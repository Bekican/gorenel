package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/ml"
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
	baseDomain     string
	acmeEmail      string
	env            string
	aiAnalyzer     *AIAnalyzer
}

func NewHTTPProxy(tm *TunnelManager, es *EventStream, gl *GeoLocator, rl *limiter.RateLimiter, ti *TrafficInspector, logger *zap.Logger, as *AnomalyStore, mlc *ml.Client, redisAddr string, baseDomain, acmeEmail, env string) *HTTPProxy {
	return &HTTPProxy{
		tunnelManager:  tm,
		advancedRL:     rl,
		eventStream:    es,
		geoLocator:     gl,
		inspector:      ti,
		mlClient:       mlc,
		logger:         logger,
		redisPublisher: NewRedisPublisher(redisAddr),
		anomalyStore:   as,
		baseDomain:     baseDomain,
		acmeEmail:      acmeEmail,
		env:            env,
		aiAnalyzer:     NewAIAnalyzer(),
	}
}

func (p *HTTPProxy) Start(port string) error {
	p.logger.Info("HTTP Proxy initiating", zap.String("port", port))

	// Certmagic (Auto-SSL) is disabled in production Docker environment
	// because SSL is handled by Fly.io LBs and Nginx proxy.
	/*
		if p.env == "production" && p.baseDomain != "" && p.acmeEmail != "" {
			certmagic.DefaultACME.Email = p.acmeEmail
			// ...
		}
	*/

	server := &http.Server{
		Addr:    port,
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
	targetKey, isCustom := resolveTargetKey(host, p.baseDomain)

	// Use a response wrapper to capture status code for analytics
	statusCode := http.StatusOK // Default, will be overwritten by errors or response
	var bytesOut int64
	var bytesReceived int64
	reqBytes := r.ContentLength
	if reqBytes < 0 {
		reqBytes = 0
	}

	// Ensure analytics are published for ALL requests, even on early return
	defer func() {
		dur := time.Since(startTime)
		p.publishEvent(targetKey, r, clientIP, statusCode, dur, reqBytes, bytesOut, "")

		// Panic recovery
		if err := recover(); err != nil {
			p.logger.Error("PANIC RECOVERED in ServeHTTP",
				zap.Any("panic", err),
				zap.String("host", r.Host),
				zap.String("path", r.URL.Path),
			)
			// If headers haven't been sent yet, send 500.
			// However, in many cases they might have been.
		}
	}()

	if targetKey == "" {
		statusCode = http.StatusBadRequest
		http.Error(w, "Invalid host or subdomain", statusCode)
		p.logger.Warn("Geçersiz host denemesi", zap.String("host", host))
		return
	}

	if p.inspector != nil && p.inspector.GetModifier() != nil {
		p.inspector.GetModifier().Apply(r)

		// Check for Status Code / Mock Body override (Gorenel Chaos & Morphing Mode)
		for _, rule := range p.inspector.GetModifier().GetRules() {
			if p.inspector.GetModifier().matches(r.URL.Path, rule.PathPattern) {
				if rule.MockBody != "" {
					sc := rule.StatusCode
					if sc == 0 {
						sc = http.StatusOK
					}
					statusCode = sc
					w.Header().Set("X-Gorenel-Morph", "Active")
					w.WriteHeader(statusCode)
					if _, err := w.Write([]byte(rule.MockBody)); err != nil {
						p.logger.Error("MockBody write error", zap.Error(err))
					}
					return
				}
				if rule.StatusCode > 0 {
					statusCode = rule.StatusCode
					w.WriteHeader(statusCode)
					fmt.Fprintf(w, "Gorenel Modifier: Overridden with status %d", statusCode)
					return
				}
			}
		}
	}

	// RateLimiter kontrolü
	if !p.advancedRL.Allow(targetKey, 1) {
		statusCode = http.StatusTooManyRequests
		http.Error(w, "Rate limit exceeded", statusCode)
		p.logger.Warn("Rate limit aşıldı", zap.String("target", targetKey))
		return
	}

	if isCustom {
		p.logger.Info("HTTP istek", zap.String("type", "custom_domain"), zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.String("host", host))
	} else {
		p.logger.Info("HTTP istek", zap.String("type", "subdomain"), zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.String("subdomain", targetKey))
	}

	session, exists := p.tunnelManager.GetTunnel(host)
	if !exists {

		session, exists = p.tunnelManager.GetTunnel(targetKey)
	}

	if !exists {
		statusCode = http.StatusNotFound
		p.logger.Warn("Tunnel not found", zap.String("host", host))
		http.Error(w, "Tunnel bulunamadı", statusCode)

		// Phase 5: Tunnel bulunmasa bile body'yi oku ve AI güvenlik taraması yap
		var aiMeta *AIMetadata
		if r.Body != nil && p.aiAnalyzer != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil && len(bodyBytes) > 0 {
				aiMeta = p.aiAnalyzer.AnalyzeRequest(r.Host, r.URL.Path, bodyBytes)
				if aiMeta != nil {
					p.logger.Info("AI Security scan (no tunnel)",
						zap.Bool("is_risk", aiMeta.IsSecurityRisk),
						zap.Float64("risk_score", aiMeta.RiskScore),
					)
				}
			}
		}
		p.triggerMLAnalysis(r.Method, r.URL.Path, r.Host, r.ContentLength, time.Since(startTime), statusCode, 0, clientIP, targetKey, aiMeta)
		return
	}

	if IsWebSocketUpgrade(r) {
		p.HandleWebSocket(w, r, session, targetKey)
		atomic.AddInt64(&WebSocketConnections, 1)
		return
	}

	stream, err := session.Open()
	if err != nil {
		p.logger.Error("Stream open error", zap.Error(err))
		http.Error(w, "Connection failed", http.StatusBadGateway)
		return
	}
	defer stream.Close()

	// Phase 5: Streaming Body Capture (Bounded to 5MB to prevent OOM)
	var reqBodyBuf bytes.Buffer
	if r.Body != nil {
		bw := &BoundedWriter{W: &reqBodyBuf, Limit: 5 * 1024 * 1024}
		r.Body = io.NopCloser(io.TeeReader(r.Body, bw))
	}

	captured := &CapturedRequest{
		ID:         uuid.New().String(),
		Subdomain:  targetKey,
		Method:     r.Method,
		Path:       r.URL.Path,
		ReqHeaders: r.Header,
		// ReqBody will be populated after forwarding
		Timestamp: startTime,
	}

	captureWriter := NewResponseCaptureWriter(w)

	// Set proxy headers
	if r.Header.Get("X-Forwarded-For") == "" {
		r.Header.Set("X-Forwarded-For", clientIP)
	}
	if r.Header.Get("X-Forwarded-Proto") == "" {
		r.Header.Set("X-Forwarded-Proto", "http")
	}
	r.Header.Set("X-Forwarded-Host", r.Host)

	if err := r.Write(stream); err != nil {
		statusCode = http.StatusBadGateway
		p.logger.Error("Request forwarding failed", zap.Error(err))
		http.Error(w, "Upstream error", statusCode)

		// Phase 5 Fix: Upstream fail olsa bile AI analizi ve anomali kaydı yap
		captured.ReqBody = reqBodyBuf.Bytes()
		if p.inspector != nil && p.aiAnalyzer != nil {
			aiMeta := p.aiAnalyzer.AnalyzeRequest(r.Host, r.URL.Path, captured.ReqBody)
			if aiMeta != nil {
				captured.AIMetadata = aiMeta
				p.logger.Info("AI Security scan on failed upstream",
					zap.String("provider", aiMeta.Provider),
					zap.Bool("is_risk", aiMeta.IsSecurityRisk),
				)
			}
			p.inspector.Record(captured)
			p.triggerMLAnalysis(r.Method, r.URL.Path, r.Host, r.ContentLength, time.Since(startTime), statusCode, 0, clientIP, targetKey, aiMeta)
		}
		return
	}

	// Now populate the captured req body
	captured.ReqBody = reqBodyBuf.Bytes()

	// Phase 6: Proper HTTP Response Parsing from Tunnel
	// Yamux stream provides raw HTTP bytes, we MUST parse them to set headers/status correctly.
	respReader := bufio.NewReader(stream)
	resp, err := http.ReadResponse(respReader, r)
	if err != nil {
		p.logger.Error("Failed to read response from tunnel", zap.Error(err))
		http.Error(w, "Tunnel response error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers from tunnel response to browser response
	for k, vv := range resp.Header {
		for _, v := range vv {
			captureWriter.Header().Add(k, v)
		}
	}
	captureWriter.WriteHeader(resp.StatusCode)

	// Stream actual body to browser
	bytesReceived, err = io.Copy(captureWriter, resp.Body)
	if err != nil {
		p.logger.Error("Response body copy error", zap.Error(err))
	}

	bytesOut = bytesReceived
	statusCode = resp.StatusCode

	captured.RespHeaders = captureWriter.Header()
	captured.RespBody = captureWriter.Body.Bytes()
	captured.StatusCode = captureWriter.StatusCode
	captured.Duration = time.Since(startTime)
	if p.inspector != nil {
		// AI Analysis Phase
		if p.aiAnalyzer != nil {
			aiMeta := p.aiAnalyzer.AnalyzeRequest(r.Host, r.URL.Path, captured.ReqBody)
			if aiMeta != nil {
				p.aiAnalyzer.AnalyzeResponse(aiMeta, captured.RespBody)
				captured.AIMetadata = aiMeta
				p.logger.Info("AI Traffic Detected",
					zap.String("provider", aiMeta.Provider),
					zap.String("model", aiMeta.Model),
					zap.Int("tokens", aiMeta.Tokens.Total),
				)
			}
		}

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
				p.logger.Error("Redis publish hatası", zap.Error(err))
			}
		}()
	}

	// ML Anomali Kontrolu
	p.triggerMLAnalysis(r.Method, r.URL.Path, r.Host, r.ContentLength, responseTime, captureWriter.StatusCode, bytesReceived, clientIP, targetKey, captured.AIMetadata)

	p.logger.Debug("İstek tamamlandı",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", responseTime),
	)
}

func (p *HTTPProxy) triggerMLAnalysis(method, path, host string, requestSize int64, duration time.Duration, statusCode int, bytesReceived int64, clientIP string, targetKey string, aiMeta *AIMetadata) {
	if p.mlClient == nil {
		return
	}
	if requestSize < 0 {
		requestSize = 0
	}

	requestData := map[string]interface{}{
		"method":        method,
		"path":          path,
		"response_time": duration.Milliseconds(),
		"status_code":   statusCode,
		"request_size":  requestSize,
		"response_size": bytesReceived,
	}

	p.mlClient.PredictCompareAsync(requestData, func(resp *ml.ComparisonResponse, err error) {
		isAnomaly := false
		if err == nil && resp.Consensus.AnyAnomaly {
			isAnomaly = true
		}

		// Also consider AI Security risk as an anomaly trigger
		if aiMeta != nil && aiMeta.IsSecurityRisk {
			isAnomaly = true
		}

		if isAnomaly {
			// Hangi modeller anomali dedi?
			flaggedBy := []string{}
			if err == nil {
				flaggedBy = append(flaggedBy, resp.Consensus.FlaggedBy...)
			}
			if aiMeta != nil && aiMeta.IsSecurityRisk {
				flaggedBy = append(flaggedBy, "AI_SECURITY_ANALYSER")
			}
			detectedBy := strings.Join(flaggedBy, ", ")

			// En yüksek anomali skorunu bul
			var maxScore float64
			var ifScore, aeScore float64
			if err == nil {
				for name, mResult := range resp.Models {
					if mResult.AnomalyScore > maxScore {
						maxScore = mResult.AnomalyScore
					}
					if name == "isolation_forest" {
						ifScore = mResult.AnomalyScore
					} else if name == "autoencoder" {
						aeScore = mResult.AnomalyScore
					}
				}
			}

			// If AI risk is higher, use it
			if aiMeta != nil && aiMeta.RiskScore > maxScore {
				maxScore = aiMeta.RiskScore
			}

			p.logger.Warn("Anomali tespit edildi!",
				zap.String("path", path),
				zap.String("method", method),
				zap.String("detected_by", detectedBy),
				zap.Float64("max_score", maxScore),
			)

			// Anomali deposuna kaydet
			if p.anomalyStore != nil {
				riskReason := ""
				if aiMeta != nil {
					riskReason = aiMeta.RiskReason
				}
				p.anomalyStore.Add(AnomalyRecord{
					ID:           uuid.New().String(),
					Timestamp:    time.Now(),
					Subdomain:    targetKey,
					Method:       method,
					Path:         path,
					ClientIP:     clientIP,
					AnomalyScore: maxScore,
					DetectedBy:   detectedBy,
					IFScore:      ifScore,
					AEScore:      aeScore,
					RiskReason:   riskReason,
				})
			}
		}
	})
	_ = host // kept for future ML enrichment dimensions
}

func resolveTargetKey(host string, baseDomain string) (key string, isCustom bool) {

	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Ensure baseDomain has a leading dot for suffix matching if it's not empty
	dotDomain := baseDomain
	if dotDomain != "" && !strings.HasPrefix(dotDomain, ".") {
		dotDomain = "." + dotDomain
	}

	if dotDomain != "" && strings.HasSuffix(host, dotDomain) {
		sub := strings.TrimSuffix(host, dotDomain)
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
	event.BytesSent = bytesOut
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

// BoundedWriter writes at most Limit bytes to the underlying writer,
// discard the rest, but always returns the original length to keep streams flowing.
type BoundedWriter struct {
	W     io.Writer
	Limit int64
	n     int64
}

func (bw *BoundedWriter) Write(p []byte) (n int, err error) {
	if bw.n >= bw.Limit {
		return len(p), nil // Discard
	}
	rem := bw.Limit - bw.n
	toWrite := len(p)
	if int64(toWrite) > rem {
		toWrite = int(rem)
	}
	nWritten, err := bw.W.Write(p[:toWrite])
	bw.n += int64(nWritten)
	if err != nil {
		return nWritten, err
	}
	return len(p), nil
}
