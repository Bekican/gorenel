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
	p.logger.Info("WebSocket geçişi başlatılıyor", zap.String("subdomain", subdomain), zap.String("path", r.URL.Path))

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

	// Set/Preserve proxy headers for WS
	if r.Header.Get("X-Forwarded-For") == "" {
		r.Header.Set("X-Forwarded-For", clientIP)
	}
	// Caddy/Cloudflare sets this. We ensure it's "https" because WS is over SSL here.
	if r.Header.Get("X-Forwarded-Proto") == "" {
		r.Header.Set("X-Forwarded-Proto", "https")
	}
	r.Header.Set("X-Forwarded-Host", r.Host)

	// Write the upgrade request to the tunnel
	if err := r.Write(stream); err != nil {
		p.logger.Error("WebSocket isteği tünele iletilemedi", zap.Error(err))
		return
	}

	// Read the response from the tunnel (Expected: 101 Switching Protocols)
	respReader := bufio.NewReader(stream)
	resp, err := http.ReadResponse(respReader, r)
	if err != nil {
		p.logger.Error("Tünelden WebSocket yanıtı (101) okunamadı", zap.Error(err))
		return
	}
	
	// Forward the tunnel's response headers back to the browser
	if err := resp.Write(clientConn); err != nil {
		p.logger.Error("İstemciye (tarayıcı) WebSocket yanıtı yazılamadı", zap.Error(err))
		return
	}

	// Handle data remaining in the hijack reader (Client -> Server data)
	if bufrw.Reader.Buffered() > 0 {
		buffered := make([]byte, bufrw.Reader.Buffered())
		bufrw.Reader.Read(buffered)
		stream.Write(buffered)
	}

	done := make(chan struct{}, 2)

	// Bidirectional tunnel:
	// Browser -> Tunnel -> Localhost
	go func() {
		io.Copy(stream, clientConn)
		done <- struct{}{}
	}()

	// Localhost -> Tunnel -> Browser
	go func() {
		// Use the respReader because it might have already buffered some WebSocket frames
		io.Copy(clientConn, respReader)
		done <- struct{}{}
	}()

	<-done
	p.logger.Info("WebSocket bağlantısı sonlandırıldı", zap.String("subdomain", subdomain))
}
