package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/ml"
	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/caddyserver/certmagic"
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

	if p.env == "production" && p.baseDomain != "" && p.acmeEmail != "" {
		certmagic.DefaultACME.Email = p.acmeEmail
		// production'da let's encrypt kullanılmalı, staging testi için:
		// certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA

		magic := certmagic.NewDefault()

		// On-demand certificates for subdomains
		magic.OnDemand = &certmagic.OnDemandConfig{
			DecisionFunc: func(ctx context.Context, name string) error {
				if strings.HasSuffix(name, p.baseDomain) {
					return nil
				}
				return fmt.Errorf("domain not allowed")
			},
		}

		p.logger.Info("HTTPS automation enabled via Certmagic", zap.String("base_domain", p.baseDomain))

		// Serve both HTTP (redirect) and HTTPS
		return certmagic.HTTPS([]string{p.baseDomain, "*" + p.baseDomain}, p)
	}

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
	targetKey, isCustom := resolveTargetKey(host)

	// Use a response wrapper to capture status code for analytics
	statusCode := http.StatusOK // Default, will be overwritten by errors or response
	var bytesOut int64

	// Ensure analytics are published for ALL requests, even on early return
	defer func() {
		dur := time.Since(startTime)
		p.publishEvent(targetKey, r, clientIP, statusCode, dur, 0, bytesOut, "")

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
					w.Write([]byte(rule.MockBody))
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
		p.triggerMLAnalysis(r, time.Since(startTime), statusCode, 0, clientIP, targetKey, aiMeta)
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
			p.triggerMLAnalysis(r, time.Since(startTime), statusCode, 0, clientIP, targetKey, aiMeta)
		}
		return
	}

	// Now populate the captured req body
	captured.ReqBody = reqBodyBuf.Bytes()

	bytesReceived, err := io.Copy(captureWriter, stream)
	if err != nil {
		p.logger.Error("Response copy error", zap.Error(err))
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
	p.triggerMLAnalysis(r, responseTime, captureWriter.StatusCode, bytesReceived, clientIP, targetKey, captured.AIMetadata)

	p.logger.Debug("İstek tamamlandı",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", responseTime),
	)
}

func (p *HTTPProxy) triggerMLAnalysis(r *http.Request, duration time.Duration, statusCode int, bytesReceived int64, clientIP string, targetKey string, aiMeta *AIMetadata) {
	if p.mlClient == nil {
		return
	}

	requestData := map[string]interface{}{
		"method":        r.Method,
		"path":          r.URL.Path,
		"response_time": duration.Milliseconds(),
		"status_code":   statusCode,
		"request_size":  r.ContentLength,
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
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
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
					Method:       r.Method,
					Path:         r.URL.Path,
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
