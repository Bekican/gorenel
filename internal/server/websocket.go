package server

import (
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
	// Chrome uyumluluğu için "Contains" kullanıyoruz
	connectionHeader := strings.ToLower(r.Header.Get("Connection"))
	upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
	return strings.Contains(connectionHeader, "upgrade") && upgradeHeader == "websocket"
}

// HandleWebSocket: WebSocket bağlantısını yönetir.
// Not: HTTPProxy struct'ının bir metodu olarak tanımladık, böylece diğer dosyadan çağrılabilir.
func (p *HTTPProxy) HandleWebSocket(w http.ResponseWriter, r *http.Request, session *yamux.Session, subdomain string) {
	p.logger.Info("WebSocket upgrade isteği", zap.String("subdomain", subdomain))

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
		p.logger.Error("Stream açılamadı", zap.Error(err))
		return
	}
	defer stream.Close()

	atomic.AddInt64(&WebSocketConnections, 1)
	defer atomic.AddInt64(&WebSocketConnections, -1)

	var streamID uint32
	if ys, ok := interface{}(stream).(*yamux.Stream); ok {
		streamID = ys.StreamID()
	}
	p.logger.Info("WebSocket stream açıldı", zap.String("subdomain", subdomain), zap.Uint32("stream_id", streamID))

	if err := r.Write(stream); err != nil {
		p.logger.Error("Request yazılamadı", zap.Error(err))
		return
	}

	// Buffer Drain (Kalan verileri temizle)
	if bufrw.Reader.Buffered() > 0 {
		buffered := make([]byte, bufrw.Reader.Buffered())
		bufrw.Reader.Read(buffered)
		stream.Write(buffered)
	}

	done := make(chan struct{}, 2)

	// Client -> Tünel
	go func() {
		io.Copy(stream, clientConn)
		done <- struct{}{}
	}()

	// Tünel -> Client
	go func() {
		io.Copy(clientConn, stream)
		done <- struct{}{}
	}()

	<-done
	p.logger.Info("WebSocket bağlantısı kapandı", zap.String("subdomain", subdomain))
}
