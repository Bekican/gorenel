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
	connectionHeader := strings.ToLower(r.Header.Get("Connection"))
	upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
	return strings.Contains(connectionHeader, "upgrade") && upgradeHeader == "websocket"
}

// HandleWebSocket: WebSocket bağlantısını yönetir.
func (p *HTTPProxy) HandleWebSocket(w http.ResponseWriter, r *http.Request, session *yamux.Session, subdomain string) {
	p.logger.Info("WebSocket raw-proxy başlatılıyor", zap.String("subdomain", subdomain), zap.String("path", r.URL.Path))

	// Enforce policy
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

	// Set necessary headers for the local server
	if r.Header.Get("X-Forwarded-For") == "" {
		r.Header.Set("X-Forwarded-For", clientIP)
	}
	r.Header.Set("X-Forwarded-Proto", "https")
	r.Header.Set("X-Forwarded-Host", r.Host)

	// Write the upgrade request to the tunnel
	if err := r.Write(stream); err != nil {
		p.logger.Error("WebSocket isteği iletilemedi", zap.Error(err))
		return
	}

	// Buffer Drain (Client data already in bufrw)
	if bufrw.Reader.Buffered() > 0 {
		buffered := make([]byte, bufrw.Reader.Buffered())
		bufrw.Reader.Read(buffered)
		stream.Write(buffered)
	}

	done := make(chan struct{}, 2)

	// Raw Bidirectional Data Pipe
	// This will forward the 101 Switching Protocols response and all subsequent WS frames as raw binary.
	go func() {
		io.Copy(stream, clientConn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(clientConn, stream)
		done <- struct{}{}
	}()

	<-done
	p.logger.Info("WebSocket raw-bağlantı kapandı", zap.String("subdomain", subdomain))
}
