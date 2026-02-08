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
	"github.com/Bekican/gorenel/internal/ml"
	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type HTTPProxy struct {
	tunnelManager *TunnelManager
	advancedRL    *limiter.RateLimiter
	eventStream   *EventStream
	geoLocator    *GeoLocator
	inspector     *TrafficInspector
	mlClient      *ml.Client
	logger        *zap.Logger
}

func NewHTTPProxy(tm *TunnelManager, es *EventStream, gl *GeoLocator, rl *limiter.RateLimiter, ti *TrafficInspector, logger *zap.Logger) *HTTPProxy {
	var mlClient *ml.Client
	if logger != nil {
		mlClient = ml.NewClient("http://localhost:5000", logger)
	}

	return &HTTPProxy{
		tunnelManager: tm,
		advancedRL:    rl,
		eventStream:   es,
		geoLocator:    gl,
		inspector:     ti,
		mlClient:      mlClient,
		logger:        logger,
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
	// 1. Zaman sayacını başlat
	startTime := time.Now()

	IncrementRequest()
	IncrementActiveConnections()
	defer DecrementActiveConnections()

	// Client IP'yi al
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	host := r.Host
	// --- STEP 2 UPDATE: Host çözümleme ve Routing mantığı ---
	// Artık targetKey (oranın subdomaini veya domaini) ve tipini ayırıyoruz.
	targetKey, isCustom := resolveTargetKey(host)

	if targetKey == "" {
		http.Error(w, "Invalid host or subdomain", http.StatusBadRequest)
		log.Printf("Geçersiz host denemesi: %s", host)
		return
	}

	// --- NEW: Traffic Modification Hook ---
	if p.inspector != nil && p.inspector.GetModifier() != nil {
		p.inspector.GetModifier().Apply(r)
	}

	// RateLimiter kontrolü
	if !p.advancedRL.Allow(targetKey, 1) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		log.Printf("Ratelimit aşıldı: %s", targetKey)
		return
	}

	if isCustom {
		log.Printf("HTTP Özel Domain İstek: %s %s (Host: %s)", r.Method, r.URL.Path, host)
	} else {
		log.Printf("HTTP Subdomain İstek: %s %s (Sub: %s)", r.Method, r.URL.Path, targetKey)
	}

	// Tüneli bul (TunnelManager artık host bilgisini akıllıca sorgular)
	session, exists := p.tunnelManager.GetTunnel(host)
	if !exists {
		// Eğer tam host ile bulunamadıysa, belki portsuz veya subdomain olarak denemek gerekebilir
		session, exists = p.tunnelManager.GetTunnel(targetKey)
	}

	if !exists {
		log.Printf("[HTTP] Tunnel not found for: %s", host)
		http.Error(w, "Tunnel bulunamadı", http.StatusNotFound)
		return
	}

	// WebSocket Kontrolü
	if IsWebSocketUpgrade(r) {
		p.HandleWebSocket(w, r, session, targetKey)
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

	// Traffic Capture
	reqBody, _ := InterceptBody(r)
	captured := &CapturedRequest{
		ID:         uuid.New().String(),
		Subdomain:  targetKey,
		Method:     r.Method,
		Path:       r.URL.Path,
		ReqHeaders: r.Header,
		ReqBody:    reqBody,
		Timestamp:  startTime,
	}

	captureWriter := NewResponseCaptureWriter(w)

	if err := r.Write(stream); err != nil {
		log.Printf("[HTTP] Request forwarding failed: %v", err)
		http.Error(w, "Upstream error", http.StatusBadGateway)
		return
	}

	// Response'u kopyalarken yakala
	bytesReceived, err := io.Copy(captureWriter, stream)
	if err != nil {
		log.Printf("Response copy error: %v", err)
	}

	// Traffic Capture Finalize
	captured.RespHeaders = captureWriter.Header()
	captured.RespBody = captureWriter.Body.Bytes()
	captured.StatusCode = captureWriter.StatusCode
	captured.Duration = time.Since(startTime)
	if p.inspector != nil {
		p.inspector.Record(captured)
	}

	// Analytics
	responseTime := time.Since(startTime)
	p.publishEvent(targetKey, r, clientIP, captureWriter.StatusCode, responseTime, 0, bytesReceived, "")

	// ML Anomali Kontrolu (Async - istegi yavaslat maz)
	if p.mlClient != nil {
		requestData := map[string]interface{}{
			"method":        r.Method,
			"path":          r.URL.Path,
			"response_time": responseTime.Milliseconds(),
			"status_code":   captureWriter.StatusCode,
			"request_size":  r.ContentLength,
			"response_size": bytesReceived,
		}

		p.mlClient.PredictAsync(requestData, func(resp *ml.PredictionResponse, err error) {
			if err == nil && resp.IsAnomaly {
				p.logger.Warn("Anomali tespit edildi!",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.Float64("score", resp.AnomalyScore),
				)
			}
		})
	}

	log.Printf("İstek tamamlandı : %s %s (%v)", r.Method, r.URL.Path, responseTime)
}

// resolveTargetKey: Gelen host bilgisinin subdomain mi yoksa custom domain mi olduğunu ayırır.
func resolveTargetKey(host string) (key string, isCustom bool) {
	// Port bilgisini temizle (örn: :8080)
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Eğer host ana tünel domaini ile bitiyorsa (örn: .gorenel.io)
	// protocol.BaseDomain'in ".gorenel.io" formatında olduğunu varsayıyoruz (başına nokta eklenmiş hali)
	if strings.HasSuffix(host, protocol.BaseDomain) {
		parts := strings.Split(host, ".")
		// parts: ["abc", "gorenel", "io"]
		if len(parts) >= 3 {
			return parts[0], false // Bu bir subdomaindir
		}
	}

	// Eğer ana domain değilse, bu bir Custom Domain'dir
	return host, true
}

func (p *HTTPProxy) publishEvent(subdomain string, r *http.Request, clientIP string, statusCode int, responseTime time.Duration, bytesIn, bytesOut int64, errorMsg string) {
	if p.eventStream == nil {
		return
	}

	event := NewRequestEvent(subdomain, r.Method, r.URL.Path, r.UserAgent(), clientIP)
	event.StatusCode = statusCode
	event.ResponseTime = responseTime
	event.BytesReceived = bytesIn
	event.ByteSent = bytesOut
	event.Error = errorMsg

	if p.geoLocator != nil {
		go func() {
			if loc, err := p.geoLocator.Lookup(clientIP); err == nil {
				event.GeoCountry = loc.Country
				event.GeoCity = loc.City
			}
			p.eventStream.publish(event)
		}()
	} else {
		p.eventStream.publish(event)
	}
}
