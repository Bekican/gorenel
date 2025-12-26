package server

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/yamux"
)

func IsWebSocketUpgrade(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}

// websocket bağlantısını tünele yönlendiriyoruz
func (p *HTTPProxy) HandleWebSocket(w http.ResponseWriter, r *http.Request, session *yamux.Session, subdomain string) {
	log.Printf("Websocket upgrade isteği geldi : %s", subdomain)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, bufrw, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Hijack hatası : %v", err)
		return
	}
	defer clientConn.Close()

	//yamux üzerinden yeni stream aç
	stream, err := session.Open()
	if err != nil {
		log.Printf("Stream açılamadı : %v", err)
		return
	}
	defer stream.Close()

	log.Printf("Websocket stream açıldı : %s (ID:%d)", subdomain, stream.StreamID())

	if err := r.Write(stream); err != nil {
		log.Printf("Request yazılamadı: %v", err)
		return
	}

	//önceden kalan işlenmemiş veri var mı
	if bufrw.Reader.Buffered() > 0 {
		buffered := make([]byte, bufrw.Reader.Buffered())
		bufrw.Reader.Read(buffered)
		stream.Write(buffered)
	}
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(stream, clientConn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(clientConn, stream)
		done <- struct{}{}
	}()

	<-done
	log.Printf("Websocket bağlantısı kapandı :%s", subdomain)
}

var (
	WebSocketConnections int64
	WebSocketMessages    int64
)
