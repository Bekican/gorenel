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
	tunnelManager  *TunnelManager
	advancedRL     *limiter.RateLimiter
	eventStream    *EventStream
	geoLocator     *GeoLocator
	inspector      *TrafficInspector
	mlClient       *ml.Client
	logger         *zap.Logger
	redisPublisher *RedisPublisher
	anomalyStore   *AnomalyStore
}

func NewHTTPProxy(tm *TunnelManager, es *EventStream, gl *GeoLocator, rl *limiter.RateLimiter, ti *TrafficInspector, logger *zap.Logger, as *AnomalyStore) *HTTPProxy {
	var mlClient *ml.Client
	if logger != nil {
		mlClient = ml.NewClient("http://localhost:5000", logger)
	}

	return &HTTPProxy{
		tunnelManager:  tm,
		advancedRL:     rl,
		eventStream:    es,
		geoLocator:     gl,
		inspector:      ti,
		mlClient:       mlClient,
		logger:         logger,
		redisPublisher: NewRedisPublisher("localhost:6379"),
		anomalyStore:   as,
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
	targetKey, isCustom := resolveTargetKey(host)

	// Use a response wrapper to capture status code for analytics
	statusCode := http.StatusOK // Default, will be overwritten by errors or response
	var bytesOut int64

	// Ensure analytics are published for ALL requests, even on early return
	defer func() {
		dur := time.Since(startTime)
		p.publishEvent(targetKey, r, clientIP, statusCode, dur, 0, bytesOut, "")
	}()

	if targetKey == "" {
		statusCode = http.StatusBadRequest
		http.Error(w, "Invalid host or subdomain", statusCode)
		log.Printf("Geçersiz host denemesi: %s", host)
		return
	}

	if p.inspector != nil && p.inspector.GetModifier() != nil {
		p.inspector.GetModifier().Apply(r)
	}

	// RateLimiter kontrolü
	if !p.advancedRL.Allow(targetKey, 1) {
		statusCode = http.StatusTooManyRequests
		http.Error(w, "Rate limit exceeded", statusCode)
		log.Printf("Ratelimit aşıldı: %s", targetKey)
		return
	}

	if isCustom {
		log.Printf("HTTP Özel Domain İstek: %s %s (Host: %s)", r.Method, r.URL.Path, host)
	} else {
		log.Printf("HTTP Subdomain İstek: %s %s (Sub: %s)", r.Method, r.URL.Path, targetKey)
	}

	session, exists := p.tunnelManager.GetTunnel(host)
	if !exists {

		session, exists = p.tunnelManager.GetTunnel(targetKey)
	}

	if !exists {
		statusCode = http.StatusNotFound
		log.Printf("[HTTP] Tunnel not found for: %s", host)
		http.Error(w, "Tunnel bulunamadı", statusCode)
		return
	}

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
		statusCode = http.StatusBadGateway
		log.Printf("[HTTP] Request forwarding failed: %v", err)
		http.Error(w, "Upstream error", statusCode)
		return
	}

	bytesReceived, err := io.Copy(captureWriter, stream)
	if err != nil {
		log.Printf("Response copy error: %v", err)
	}
	bytesOut = bytesReceived
	statusCode = captureWriter.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	captured.RespHeaders = captureWriter.Header()
	captured.RespBody = captureWriter.Body.Bytes()
	captured.StatusCode = captureWriter.StatusCode
	captured.Duration = time.Since(startTime)
	if p.inspector != nil {
		p.inspector.Record(captured)
	}

	// Update Tunnel Stats
	p.tunnelManager.UpdateStats(targetKey, 0, bytesReceived)

	responseTime := time.Since(startTime)

	// Redis'e Ham Trafik Verisini Gönder
	if p.redisPublisher != nil {
		trafficData := TrafficData{
			Method:       r.Method,
			Path:         r.URL.Path,
			StatusCode:   captureWriter.StatusCode,
			ResponseTime: responseTime.Milliseconds(),
			RequestSize:  r.ContentLength,
			ResponseSize: bytesReceived,
			ClientIP:     clientIP,
			Timestamp:    time.Now().Format(time.RFC3339),
		}
		go func() {
			if err := p.redisPublisher.Publish(trafficData); err != nil {
				log.Printf("Redis publish hatası: %v", err)
			}
		}()
	}

	// ML Anomali Kontrolu
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

				// Anomali deposuna kaydet
				if p.anomalyStore != nil {
					p.anomalyStore.Add(AnomalyRecord{
						ID:           uuid.New().String(),
						Timestamp:    time.Now(),
						Subdomain:    targetKey,
						Method:       r.Method,
						Path:         r.URL.Path,
						ClientIP:     clientIP,
						AnomalyScore: resp.AnomalyScore,
					})
				}
			}
		})
	}

	log.Printf("İstek tamamlandı : %s %s (%v)", r.Method, r.URL.Path, responseTime)
}

func resolveTargetKey(host string) (key string, isCustom bool) {

	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Eğer host ana tünel domaini ile bitiyorsa (örn: .gorenel.io)
	// protocol.BaseDomain'in ".gorenel.io" formatında olduğunu varsayıyoruz (başına nokta eklenmiş hali)
	if strings.HasSuffix(host, protocol.BaseDomain) {
		sub := strings.TrimSuffix(host, protocol.BaseDomain)
		return sub, false
	}

	// Eğer ana domain değilse veya localhost ise
	if host == "localhost" || host == "127.0.0.1" {
		return "localhost", false
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
