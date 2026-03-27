package server

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/ml"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	InspectorQueueDropped int64
	MLConcurrencyDropped  int64
	MLInFlight            int64
)

type HTTPProxy struct {
	tunnelManager         *TunnelManager
	advancedRL            *limiter.RateLimiter
	eventStream           *EventStream
	geoLocator            *GeoLocator
	inspector             *TrafficInspector
	mlClient              *ml.Client
	logger                *zap.Logger
	redisPublisher        *RedisPublisher
	anomalyStore          *AnomalyStore
	baseDomain            string
	acmeEmail             string
	env                   string
	aiAnalyzer            *AIAnalyzer
	inspectorMaxBodyBytes int64
	inspectorSamplingRate float64
	fullCaptureUntil      sync.Map // subdomain -> unixnano expiry
	inspectQueue          chan *CapturedRequest
	rngMu                 sync.Mutex
	rng                   *rand.Rand
	mlSem                 chan struct{}
}

func NewHTTPProxy(tm *TunnelManager, es *EventStream, gl *GeoLocator, rl *limiter.RateLimiter, ti *TrafficInspector, logger *zap.Logger, as *AnomalyStore, mlc *ml.Client, redisAddr string, baseDomain, acmeEmail, env string, inspectorMaxBodyBytes int64, inspectorSamplingRate float64) *HTTPProxy {
	p := &HTTPProxy{
		tunnelManager:         tm,
		advancedRL:            rl,
		eventStream:           es,
		geoLocator:            gl,
		inspector:             ti,
		mlClient:              mlc,
		logger:                logger,
		redisPublisher:        NewRedisPublisher(redisAddr),
		anomalyStore:          as,
		baseDomain:            baseDomain,
		acmeEmail:             acmeEmail,
		env:                   env,
		aiAnalyzer:            NewAIAnalyzer(),
		inspectorMaxBodyBytes: inspectorMaxBodyBytes,
		inspectorSamplingRate: inspectorSamplingRate,
		inspectQueue:          make(chan *CapturedRequest, 512),
		// Use a per-proxy RNG; seeded to avoid deterministic sampling across restarts.
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		// Bound ML request concurrency to protect the proxy under load/spikes.
		mlSem: make(chan struct{}, 32),
	}

	// Async workers: move inspector/AI analysis off request path.
	// Keep this small and bounded; drops are preferable to request latency.
	const inspectorWorkers = 2
	for i := 0; i < inspectorWorkers; i++ {
		go func() {
			for req := range p.inspectQueue {
				if req == nil {
					continue
				}
				if p.aiAnalyzer != nil {
					aiMeta := p.aiAnalyzer.AnalyzeRequest(req.Subdomain, req.Path, req.ReqBody)
					if aiMeta != nil {
						p.aiAnalyzer.AnalyzeResponse(aiMeta, req.RespBody)
						req.AIMetadata = aiMeta
					}
				}
				if p.inspector != nil {
					p.inspector.Record(req)
				}
			}
		}()
	}

	return p
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
		Addr:              port,
		Handler:           p,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
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

	// Sampling gate: heavy inspector/ML work is executed only for a fraction of requests.
	captureEnabled := false
	if p.inspector != nil {
		if p.inspectorSamplingRate >= 1 {
			captureEnabled = true
		} else if p.inspectorSamplingRate > 0 {
			p.rngMu.Lock()
			captureEnabled = p.rng.Float64() < p.inspectorSamplingRate
			p.rngMu.Unlock()
		}
	}
	// If an anomaly was detected recently for this tunnel, temporarily force full capture.
	if !captureEnabled {
		if v, ok := p.fullCaptureUntil.Load(targetKey); ok {
			if exp, ok2 := v.(int64); ok2 && exp > time.Now().UnixNano() {
				captureEnabled = true
			}
		}
	}

	// Use a response wrapper to capture status code for analytics
	statusCode := http.StatusOK // Default, will be overwritten by errors or response
	var bytesOut int64
	var bytesReceived int64
	reqBytes := r.ContentLength
	if reqBytes < 0 {
		reqBytes = 0
	}

	// Enforce per-tunnel policies (auth, allowlist, redirect, etc.)
	// Apply early so we avoid spending resources on unauthorized traffic.
	if policy, ok := p.tunnelManager.GetTunnelPolicy(host); ok {
		if enforcePolicy(w, r, clientIP, policy) {
			statusCode = http.StatusForbidden
			return
		}
	} else if policy, ok := p.tunnelManager.GetTunnelPolicy(targetKey); ok {
		if enforcePolicy(w, r, clientIP, policy) {
			statusCode = http.StatusForbidden
			return
		}
	}

	// Per-tunnel request shaping (path + request headers)
	if policy, ok := p.tunnelManager.GetTunnelPolicy(host); ok {
		applyRequestPolicy(r, policy)
	} else if policy, ok := p.tunnelManager.GetTunnelPolicy(targetKey); ok {
		applyRequestPolicy(r, policy)
	}

	// Panic recovery MUST be registered first (runs last in LIFO) so it catches
	// panics from any subsequent deferred work as well.
	defer func() {
		if err := recover(); err != nil {
			p.logger.Error("PANIC RECOVERED in ServeHTTP",
				zap.Any("panic", err),
				zap.String("host", r.Host),
				zap.String("path", r.URL.Path),
			)
		}
	}()

	// Ensure analytics are published for ALL requests, even on early return
	defer func() {
		dur := time.Since(startTime)
		p.publishEvent(targetKey, r, clientIP, statusCode, dur, reqBytes, bytesOut, "")
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
	// If per-tunnel policy specifies a custom quota, prefer it; otherwise use default limiter tiers.
	if pol, ok := p.tunnelManager.GetTunnelPolicy(targetKey); ok && pol.RateLimitEnabled && pol.RateLimitRequests > 0 && pol.RateLimitWindowS > 0 {
		q := limiter.Quota{Limit: pol.RateLimitRequests, WindowSize: time.Duration(pol.RateLimitWindowS) * time.Second}
		if !p.advancedRL.AllowWithQuota("tunnel:"+targetKey+":"+clientIP, 1, q) {
			statusCode = http.StatusTooManyRequests
			http.Error(w, "Rate limit exceeded", statusCode)
			p.logger.Warn("Per-tunnel rate limit exceeded", zap.String("target", targetKey))
			return
		}
	} else if !p.advancedRL.Allow(targetKey, 1) {
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

		// Tunnel bulunmasa bile (sampling'e göre) bounded body consume edip AI güvenlik taraması yap.
		if captureEnabled && r.Body != nil && p.aiAnalyzer != nil {
			var reqBodyBuf bytes.Buffer
			bw := &BoundedWriter{W: &reqBodyBuf, Limit: p.inspectorMaxBodyBytes}
			r.Body = io.NopCloser(io.TeeReader(r.Body, bw))
			_, _ = io.Copy(io.Discard, r.Body) // drain without unbounded buffering

			aiMeta := p.aiAnalyzer.AnalyzeRequest(r.Host, r.URL.Path, reqBodyBuf.Bytes())
			if aiMeta != nil {
				p.logger.Info("AI Security scan (no tunnel)",
					zap.Bool("is_risk", aiMeta.IsSecurityRisk),
					zap.Float64("risk_score", aiMeta.RiskScore),
				)
			}
			p.triggerMLAnalysis(r.Method, r.URL.Path, r.Host, r.ContentLength, time.Since(startTime), statusCode, 0, clientIP, targetKey, aiMeta)
		}
		return
	}

	if IsWebSocketUpgrade(r) {
		p.HandleWebSocket(w, r, session, targetKey)
		// Note: WebSocketConnections counter is managed inside HandleWebSocket (increment + defer decrement).
		return
	}

	stream, err := session.Open()
	if err != nil {
		p.logger.Error("Stream open error", zap.Error(err))
		http.Error(w, "Connection failed", http.StatusBadGateway)
		return
	}
	defer stream.Close()

	// Phase 5: Streaming Body Capture (bounded, sampling-gated)
	var reqBodyBuf bytes.Buffer
	if captureEnabled && r.Body != nil {
		bw := &BoundedWriter{W: &reqBodyBuf, Limit: p.inspectorMaxBodyBytes}
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

	maxCaptureBytes := int64(0)
	if captureEnabled {
		maxCaptureBytes = p.inspectorMaxBodyBytes
	}
	captureWriter := NewResponseCaptureWriter(w, maxCaptureBytes)

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

		// Upstream fail olsa bile (sampling'e göre) AI analizi ve inspector/ML kaydı yap.
		if captureEnabled && p.inspector != nil && p.aiAnalyzer != nil {
			captured.ReqBody = reqBodyBuf.Bytes()
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
	if captureEnabled {
		captured.ReqBody = reqBodyBuf.Bytes()
	}

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

	// Apply per-tunnel response header edits BEFORE WriteHeader so they reach the client.
	if pol, ok := p.tunnelManager.GetTunnelPolicy(targetKey); ok {
		applyResponsePolicy(captureWriter, pol)
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

	if captureEnabled {
		select {
		case p.inspectQueue <- captured:
		default:
			// If queue is full, avoid blocking request path; drop capture.
			atomic.AddInt64(&InspectorQueueDropped, 1)
		}
	}

	// Update Tunnel Stats (bytesIn = request body size, bytesOut = response body size)
	p.tunnelManager.UpdateStats(targetKey, reqBytes, bytesReceived)

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
	if captureEnabled {
		p.triggerMLAnalysis(r.Method, r.URL.Path, r.Host, r.ContentLength, responseTime, captureWriter.StatusCode, bytesReceived, clientIP, targetKey, captured.AIMetadata)
	}

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
	if shouldIgnoreAnomalyTraffic(clientIP) {
		return
	}
	// Protect the ML service (and us) with bounded concurrency.
	select {
	case p.mlSem <- struct{}{}:
		atomic.AddInt64(&MLInFlight, 1)
	default:
		atomic.AddInt64(&MLConcurrencyDropped, 1)
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
		defer func() {
			<-p.mlSem
			atomic.AddInt64(&MLInFlight, -1)
		}()

		isAnomaly := false
		if err == nil && resp.Consensus.AnyAnomaly {
			isAnomaly = true
		}

		// Also consider AI Security risk as an anomaly trigger
		if aiMeta != nil && aiMeta.IsSecurityRisk {
			isAnomaly = true
		}

		if isAnomaly {
			// Force full capture for a short window after anomaly.
			p.fullCaptureUntil.Store(targetKey, time.Now().Add(60*time.Second).UnixNano())

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

func shouldIgnoreAnomalyTraffic(clientIP string) bool {
	raw := strings.TrimSpace(clientIP)
	if raw == "" {
		return false
	}

	// Handle values that may include "ip:port".
	if strings.Contains(raw, ":") {
		if host, _, err := net.SplitHostPort(raw); err == nil {
			raw = host
		}
	}

	addr, err := netip.ParseAddr(raw)
	if err != nil {
		return false
	}

	// Ignore non-routable/internal traffic for anomaly generation.
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsUnspecified() {
		return true
	}
	return false
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
		// Do geo lookup synchronously but with a short timeout to avoid blocking.
		// This avoids the race condition where the event fields were mutated in a goroutine
		// while the event could be read concurrently by subscribers.
		if loc, err := p.geoLocator.Lookup(clientIP); err == nil {
			event.GeoCountry = loc.Country
			event.GeoCity = loc.City
		}
	}
	p.eventStream.publish(event)
}

// BoundedWriter writes at most Limit bytes to the underlying writer,
// discard the rest, but always returns the original length to keep streams flowing.
type BoundedWriter struct {
	W     io.Writer
	Limit int64
	n     int64
}

func applyRequestPolicy(r *http.Request, policy TunnelPolicy) {
	// Path rewrite
	if strings.TrimSpace(policy.ReplacePathFrom) != "" {
		from := strings.TrimSpace(policy.ReplacePathFrom)
		to := policy.ReplacePathTo
		if strings.HasPrefix(r.URL.Path, from) {
			r.URL.Path = to + strings.TrimPrefix(r.URL.Path, from)
		}
	}
	if strings.TrimSpace(policy.PathPrefix) != "" {
		prefix := strings.TrimSpace(policy.PathPrefix)
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		if !strings.HasPrefix(r.URL.Path, prefix) {
			r.URL.Path = prefix + r.URL.Path
		}
	}

	// Header edits (request)
	for k, v := range policy.AddRequestHeaders {
		if strings.TrimSpace(k) == "" {
			continue
		}
		r.Header.Set(k, v)
	}
	for _, k := range policy.RemoveRequestHeaders {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		r.Header.Del(k)
	}
}

func applyResponsePolicy(w http.ResponseWriter, policy TunnelPolicy) {
	for k, v := range policy.AddResponseHeaders {
		if strings.TrimSpace(k) == "" {
			continue
		}
		w.Header().Set(k, v)
	}
	for _, k := range policy.RemoveResponseHeaders {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		w.Header().Del(k)
	}
}

func enforcePolicy(w http.ResponseWriter, r *http.Request, clientIP string, policy TunnelPolicy) (denied bool) {
	// HTTPS redirect
	if policy.HttpsRedirectEnabled {
		if strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))) == "http" {
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return true
		}
	}

	// IP allowlist
	if policy.IPAllowlistEnabled && len(policy.IPAllowlist) > 0 {
		addr, err := netip.ParseAddr(strings.TrimSpace(clientIP))
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return true
		}
		allowed := false
		for _, raw := range policy.IPAllowlist {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			if strings.Contains(raw, "/") {
				prefix, err := netip.ParsePrefix(raw)
				if err != nil {
					continue
				}
				if prefix.Contains(addr) {
					allowed = true
					break
				}
				continue
			}
			ip, err := netip.ParseAddr(raw)
			if err != nil {
				continue
			}
			if ip == addr {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return true
		}
	}

	// Basic auth
	if policy.BasicAuthEnabled && strings.TrimSpace(policy.BasicAuthUsername) != "" && strings.TrimSpace(policy.BasicAuthPasswordHash) != "" {
		authz := strings.TrimSpace(r.Header.Get("Authorization"))
		const prefix = "Basic "
		if !strings.HasPrefix(authz, prefix) {
			w.Header().Set("WWW-Authenticate", `Basic realm="gorenel"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return true
		}
		raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authz, prefix))
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return true
		}
		parts := strings.SplitN(string(raw), ":", 2)
		if len(parts) != 2 || parts[0] != policy.BasicAuthUsername {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return true
		}
		if bcrypt.CompareHashAndPassword([]byte(policy.BasicAuthPasswordHash), []byte(parts[1])) != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return true
		}
	}

	// KeyAuth (token)
	if policy.KeyAuthEnabled && strings.TrimSpace(policy.KeyAuthToken) != "" {
		got := strings.TrimSpace(r.Header.Get("X-TOKEN"))
		if got == "" || got != strings.TrimSpace(policy.KeyAuthToken) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return true
		}
	}

	return false
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
