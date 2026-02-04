// package server

// import (
// 	"io"
// 	"log"
// 	"net/http"
// 	"strings"

// 	"github.com/Bekican/gorenel/internal/protocol"
// 	"github.com/hashicorp/yamux"
// )

// type HTTPProxy struct {
// 	tunnelManager *TunnelManager
// 	rateLimiter   *RateLimiter
// 	eventStream   *EventStream
// 	geoLocator    *GeoLocator
// }

// func NewHTTPProxy(tm *TunnelManager, es *EventStream, gl *GeoLocator) *HTTPProxy {
// 	return &HTTPProxy{
// 		tunnelManager: tm,
// 		rateLimiter:   NewRateLimiter(100, 10),
// 		eventStream:   es,
// 		geoLocator:    gl,
// 	}
// }

// func (p *HTTPProxy) Start() error {
// 	log.Printf("[HTTP] Proxy listening on %s", protocol.ProxyPort)

// 	server := &http.Server{
// 		Addr:    protocol.ProxyPort,
// 		Handler: p,
// 	}

// 	return server.ListenAndServe()
// }

// func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

// 	IncrementRequest()
// 	IncrementActiveConnections()
// 	defer DecrementActiveConnections()

// 	host := r.Host
// 	subdomain := extractSubdomain(host)

// 	if subdomain == "" {
// 		http.Error(w, "Invalid subdomain", http.StatusBadRequest)
// 		log.Printf("Geçersiz host : %s", host)
// 		return
// 	}

// 	// RateLimiter kontrolü
// 	if !p.rateLimiter.Allow(subdomain, 1) {
// 		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
// 		log.Printf("Ratelimit aşıldı : %s", subdomain)
// 		return
// 	}

// 	log.Printf("HTTP istek : %s %s (subdomain:%s)", r.Method, r.URL.Path, subdomain)

// 	session, exists := p.tunnelManager.GetTunnel(subdomain)
// 	if !exists {
// 		log.Printf("[HTTP] Tunnel not found: %s", subdomain)
// 		http.Error(w, "Tunnel bulunamadı", http.StatusNotFound)
// 		return
// 	}

// 	if IsWebSocketUpgrade(r) {
// 		p.HandleWebSocket(w, r, session, subdomain)
// 		return
// 	}

// 	stream, err := session.Open()
// 	if err != nil {
// 		log.Printf("[HTTP] Stream open error: %v", err)
// 		http.Error(w, "Connection failed", http.StatusBadGateway)
// 		return
// 	}
// 	defer stream.Close()

// 	var streamID uint32

// 	if ys, ok := interface{}(stream).(*yamux.Stream); ok {
// 		streamID = ys.StreamID()
// 	}
// 	log.Printf("Stream açıldı : %s (ID:%d)", subdomain, streamID)

// 	if err := r.Write(stream); err != nil {
// 		log.Printf("[HTTP] Request forwarding failed: %v", err)
// 		http.Error(w, "Upstream error", http.StatusBadGateway)
// 		return
// 	}

// 	io.Copy(w, stream)

// 	log.Printf("İstek tamamlandı : %s %s", r.Method, r.URL.Path)
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
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/google/uuid"
	"github.com/hashicorp/yamux"
)

type HTTPProxy struct {
	tunnelManager *TunnelManager
	advancedRL    *limiter.RateLimiter
	eventStream   *EventStream
	geoLocator    *GeoLocator
	inspector     *TrafficInspector
}

func NewHTTPProxy(tm *TunnelManager, es *EventStream, gl *GeoLocator, rl *limiter.RateLimiter, ti *TrafficInspector) *HTTPProxy {
	return &HTTPProxy{
		tunnelManager: tm,
		advancedRL:    rl,
		eventStream:   es,
		geoLocator:    gl,
		inspector:     ti,
	}
}

func (p *HTTPProxy) Start() error {
	log.Printf("[HTTP] Proxy listening on %s", protocol.ProxyPort)

	server := &http.Server{
		Addr:    protocol.ProxyPort,
		Handler: p,
	}

	return server.ListenAndServe()
}

func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Zaman sayacını başlat (Analytics için gerekli)
	startTime := time.Now()

	IncrementRequest()
	IncrementActiveConnections()
	defer DecrementActiveConnections()

	// Client IP'yi al (Event logları için)
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	host := r.Host
	subdomain := extractSubdomain(host)

	if subdomain == "" {
		http.Error(w, "Invalid subdomain", http.StatusBadRequest)
		log.Printf("Geçersiz host : %s", host)
		return
	}

	// RateLimiter kontrolü
	if !p.advancedRL.Allow(subdomain, 1) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		log.Printf("Ratelimit aşıldı : %s", subdomain)
		return
	}

	log.Printf("HTTP istek : %s %s (subdomain:%s)", r.Method, r.URL.Path, subdomain)

	session, exists := p.tunnelManager.GetTunnel(subdomain)
	if !exists {
		log.Printf("[HTTP] Tunnel not found: %s", subdomain)
		http.Error(w, "Tunnel bulunamadı", http.StatusNotFound)
		return
	}

	// WebSocket Kontrolü
	if IsWebSocketUpgrade(r) {
		p.HandleWebSocket(w, r, session, subdomain)
		// Atomic counter artırımı eklendi
		atomic.AddInt64(&WebSocketConnections, 1)
		return
	}

	stream, err := session.Open()
	if err != nil {
		log.Printf("[HTTP] Stream open error: %v", err)
		http.Error(w, "Connection failed", http.StatusBadGateway)
		return
	}
	defer stream.Close()

	var streamID uint32
	if ys, ok := interface{}(stream).(*yamux.Stream); ok {
		streamID = ys.StreamID()
	}
	log.Printf("Stream açıldı : %s (ID:%d)", subdomain, streamID)

	// --- NEW: Traffic Capture (Request) ---
	reqBody, _ := InterceptBody(r)
	captured := &CapturedRequest{
		ID:         uuid.New().String(),
		Subdomain:  subdomain,
		Method:     r.Method,
		Path:       r.URL.Path,
		ReqHeaders: r.Header,
		ReqBody:    reqBody,
		Timestamp:  startTime,
	}

	// Wrapper for response capture
	captureWriter := NewResponseCaptureWriter(w)

	if err := r.Write(stream); err != nil {
		log.Printf("[HTTP] Request forwarding failed: %v", err)
		http.Error(w, "Upstream error", http.StatusBadGateway)
		return
	}

	// Response'u kopyala (using captureWriter)
	bytesReceived, err := io.Copy(captureWriter, stream)
	if err != nil {
		log.Printf("Response copy error: %v", err)
	}

	// --- NEW: Traffic Capture (Finalize) ---
	captured.RespHeaders = captureWriter.Header()
	captured.RespBody = captureWriter.Body.Bytes()
	captured.StatusCode = captureWriter.StatusCode
	captured.Duration = time.Since(startTime)
	if p.inspector != nil {
		p.inspector.Record(captured)
	}

	// 2. İstatistikleri Hesapla ve Kaydet (Integration Kısmı Burası)
	responseTime := time.Since(startTime)

	// Event Stream'e gönder (AnalyticsEngine bunu yakalayacak)
	// Not: Status code'u io.Copy yaparken yakalamak zordur, varsayılan 200 kabul ediyoruz
	// veya özel bir ResponseWriter wrapper yazmak gerekir. Şimdilik 200 geçiyoruz.
	p.publishEvent(subdomain, r, clientIP, 200, responseTime, 0, bytesReceived, "")

	log.Printf("İstek tamamlandı : %s %s (%v)", r.Method, r.URL.Path, responseTime)
}

// --- Yeni Eklenen Yardımcı Fonksiyonlar ---

// publishEvent - İstatistikleri EventStream'e gönderir
func (p *HTTPProxy) publishEvent(subdomain string, r *http.Request, clientIP string, statusCode int, responseTime time.Duration, bytesIn, bytesOut int64, errorMsg string) {
	if p.eventStream == nil {
		return
	}

	// RequestEvent struct'ının senin 'internal/protocol' veya 'server' paketinde tanımlı olduğunu varsayıyorum.
	// Eğer yoksa tanımlaman gerekebilir. Buradaki yapı referans koddaki NewRequestEvent kullanımıdır.
	event := NewRequestEvent(subdomain, r.Method, r.URL.Path, r.UserAgent(), clientIP)
	event.StatusCode = statusCode
	event.ResponseTime = responseTime
	event.BytesReceived = bytesIn // Response size (Tunnel -> Client)
	event.ByteSent = bytesOut     // Request size (Client -> Tunnel)
	event.Error = errorMsg

	// Geo-location lookup (Asenkron çalışır, ana akışı bloklamaz)
	if p.geoLocator != nil {
		go func() {
			if loc, err := p.geoLocator.Lookup(clientIP); err == nil {
				event.GeoCountry = loc.Country
				event.GeoCity = loc.City
			}
			// Lokasyon bilgisi eklendikten sonra publish ediliyor
			p.eventStream.publish(event)
		}()
	} else {
		// GeoLocator yoksa direkt gönder
		p.eventStream.publish(event)
	}
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
