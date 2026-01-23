package server

import (
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/hashicorp/yamux"
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
	log.Printf("🔌 Websocket upgrade isteği geldi: %s", subdomain)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, bufrw, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Hijack hatası: %v", err)
		return
	}
	defer clientConn.Close()

	stream, err := session.Open()
	if err != nil {
		log.Printf(" Stream açılamadı: %v", err)
		return
	}
	defer stream.Close()

	atomic.AddInt64(&WebSocketConnections, 1)
	defer atomic.AddInt64(&WebSocketConnections, -1)

	var streamID uint32
	if ys, ok := interface{}(stream).(*yamux.Stream); ok {
		streamID = ys.StreamID()
	}
	log.Printf("Websocket stream açıldı: %s (ID:%d)", subdomain, streamID)

	if err := r.Write(stream); err != nil {
		log.Printf("Request yazılamadı: %v", err)
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
	log.Printf("Websocket bağlantısı kapandı: %s", subdomain)
}
