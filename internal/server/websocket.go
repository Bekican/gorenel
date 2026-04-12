package server

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/hashicorp/yamux"
	"go.uber.org/zap"
)

// Global değişkenler
var (
	WebSocketConnections int64
	WebSocketMessages    int64
)

// IsWebSocketUpgrade: İsteğin WebSocket olup olmadığını kontrol eder.
func IsWebSocketUpgrade(r *http.Request) bool {
	connectionHeader := strings.ToLower(r.Header.Get("Connection"))
	upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
	return strings.Contains(connectionHeader, "upgrade") && upgradeHeader == "websocket"
}

// HandleWebSocket: WebSocket bağlantısını yönetir.
func (p *HTTPProxy) HandleWebSocket(w http.ResponseWriter, r *http.Request, session *yamux.Session, subdomain string) {
	p.logger.Info("WebSocket upgrade isteği başlatılıyor", zap.String("subdomain", subdomain), zap.String("path", r.URL.Path))

	// Enforce same tunnel policies for WS upgrades.
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	if policy, ok := p.tunnelManager.GetTunnelPolicy(r.Host); ok {
		if enforcePolicy(w, r, clientIP, policy) {
			return
		}
	} else if policy, ok := p.tunnelManager.GetTunnelPolicy(subdomain); ok {
		if enforcePolicy(w, r, clientIP, policy) {
			return
		}
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, bufrw, err := hijacker.Hijack()
	if err != nil {
		p.logger.Error("Hijack hatası", zap.Error(err))
		return
	}
	defer clientConn.Close()

	stream, err := session.Open()
	if err != nil {
		p.logger.Error("Tunnel stream açılamadı", zap.Error(err))
		return
	}
	defer stream.Close()

	atomic.AddInt64(&WebSocketConnections, 1)
	defer atomic.AddInt64(&WebSocketConnections, -1)

	// Set proxy headers for WS
	if r.Header.Get("X-Forwarded-For") == "" {
		r.Header.Set("X-Forwarded-For", clientIP)
	}
	r.Header.Set("X-Forwarded-Proto", "https") // WebSocket upgrades over Caddy are always HTTPS
	r.Header.Set("X-Forwarded-Host", r.Host)

	// Write the upgrade request to the tunnel
	if err := r.Write(stream); err != nil {
		p.logger.Error("Tünele WebSocket isteği yazılamadı", zap.Error(err))
		return
	}

	// Read the response from the tunnel (Expected: 101 Switching Protocols)
	respReader := bufio.NewReader(stream)
	resp, err := http.ReadResponse(respReader, r)
	if err != nil {
		p.logger.Error("Tünelden WebSocket yanıtı okunamadı", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	// Forward the tunnel's response headers back to the client
	// This sends the "101 Switching Protocols" and necessary Sec-WebSocket-Accept headers
	if err := resp.Write(clientConn); err != nil {
		p.logger.Error("İstemciye WebSocket yanıtı yazılamadı", zap.Error(err))
		return
	}

	// Buffer Drain (If there's any data already buffered in the client request)
	if bufrw.Reader.Buffered() > 0 {
		buffered := make([]byte, bufrw.Reader.Buffered())
		bufrw.Reader.Read(buffered)
		stream.Write(buffered)
	}

	done := make(chan struct{}, 2)

	// Bidirectional tunnel
	// Client -> Tünel
	go func() {
		io.Copy(stream, clientConn)
		done <- struct{}{}
	}()

	// Tünel -> Client
	go func() {
		// Note: We use the already existing respReader if it has buffered data from the body
		// but Switch Protocols response usually doesn't have a body.
		io.Copy(clientConn, respReader)
		done <- struct{}{}
	}()

	<-done
	p.logger.Info("WebSocket bağlantısı kapandı", zap.String("subdomain", subdomain))
}
