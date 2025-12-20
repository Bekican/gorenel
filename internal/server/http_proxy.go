// package server

// import (
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"strings"

// 	"github.com/Bekican/gorenel/internal/protocol"
// )

// type HTTPProxy struct {
// 	tunnelManager *TunnelManager
// }

// func NewHTTPProxy(tm *TunnelManager) *HTTPProxy {
// 	return &HTTPProxy{
// 		tunnelManager: tm,
// 	}
// }

// func (p *HTTPProxy) Start() error {
// 	addr := fmt.Sprintf(":%d", protocol.ProxyPort)
// 	log.Printf("[HTTP] Proxy listening on %s", addr)

// 	server := &http.Server{
// 		Addr:    addr,
// 		Handler: p,
// 	}

// 	return server.ListenAndServe()
// }

// func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	host := r.Host
// 	subdomain := extractSubdomain(host)

// 	if subdomain == "" {
// 		http.Error(w, "Invalid subdomain", http.StatusBadRequest)
// 		return
// 	}

// 	session, exists := p.tunnelManager.GetTunnel(subdomain)
// 	if !exists {
// 		log.Printf("[HTTP] Tunnel not found: %s", subdomain)
// 		http.Error(w, "Tunnel not active", http.StatusNotFound)
// 		return
// 	}

// 	stream, err := session.Open()
// 	if err != nil {
// 		log.Printf("[HTTP] Stream open error: %v", err)
// 		http.Error(w, "Connection failed", http.StatusBadGateway)
// 		return
// 	}
// 	defer stream.Close()

// 	if err := r.Write(stream); err != nil {
// 		log.Printf("[HTTP] Request forwarding failed: %v", err)
// 		http.Error(w, "Upstream error", http.StatusBadGateway)
// 		return
// 	}

// 	log.Printf("[HTTP] %s %s -> %s", r.Method, r.URL.Path, subdomain)

// 	// Backend'den gelen ham yanıtı client'a ilet
// 	io.Copy(w, stream)
// }

// func extractSubdomain(host string) string {
// 	if idx := strings.Index(host, ":"); idx != -1 {
// 		host = host[:idx]
// 	}

// 	parts := strings.Split(host, ".")
// 	if len(parts) < 2 {
// 		return ""
// 	}

// 	return parts[0]
// }

package server

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Bekican/gorenel/internal/protocol"
)

type HTTPProxy struct {
	tunnelManager *TunnelManager
}

func NewHTTPProxy(tm *TunnelManager) *HTTPProxy {
	return &HTTPProxy{
		tunnelManager: tm,
	}
}

func (p *HTTPProxy) Start() error {
	// Düzeltme: ProxyPort zaten string (":8080") olduğu için formatlamaya gerek yok.
	// Hatalı kullanım: fmt.Sprintf(":%d", protocol.ProxyPort) -> HATA VERİR
	addr := protocol.ProxyPort

	log.Printf("[HTTP] Proxy listening on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: p,
	}

	return server.ListenAndServe()
}

func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	subdomain := extractSubdomain(host)

	if subdomain == "" {
		http.Error(w, "Invalid subdomain", http.StatusBadRequest)
		return
	}

	session, exists := p.tunnelManager.GetTunnel(subdomain)
	if !exists {
		log.Printf("[HTTP] Tunnel not found: %s", subdomain)
		http.Error(w, "Tunnel not active", http.StatusNotFound)
		return
	}

	stream, err := session.Open()
	if err != nil {
		log.Printf("[HTTP] Stream open error: %v", err)
		http.Error(w, "Connection failed", http.StatusBadGateway)
		return
	}
	defer stream.Close()

	if err := r.Write(stream); err != nil {
		log.Printf("[HTTP] Request forwarding failed: %v", err)
		http.Error(w, "Upstream error", http.StatusBadGateway)
		return
	}

	log.Printf("[HTTP] %s %s -> %s", r.Method, r.URL.Path, subdomain)

	io.Copy(w, stream)
}

func extractSubdomain(host string) string {
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return ""
	}

	return parts[0]
}
