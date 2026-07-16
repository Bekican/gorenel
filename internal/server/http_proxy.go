package server

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/ml"
	"github.com/google/uuid"
	"github.com/hashicorp/yamux"
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
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       0, // Disable ReadTimeout for long-lived tunnels/WebSockets
		WriteTimeout:      0, // Disable WriteTimeout for long-lived tunnels/WebSockets
		IdleTimeout:       300 * time.Second,
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
	clientIP := getClientIP(r)

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
	// Bypass rate limiting for common static assets to improve page load performance for web apps.
	isAsset := false
	extIdx := strings.LastIndex(r.URL.Path, ".")
	if extIdx != -1 {
		ext := strings.ToLower(r.URL.Path[extIdx:])
		switch ext {
		case ".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".otf", ".map":
			isAsset = true
		}
	}

	if !isAsset {
		// If per-tunnel policy specifies a custom quota, prefer it; otherwise use default limiter tiers.
		if pol, ok := p.tunnelManager.GetTunnelPolicy(targetKey); ok && pol.RateLimitEnabled && pol.RateLimitRequests > 0 && pol.RateLimitWindowS > 0 {
			q := limiter.Quota{Limit: pol.RateLimitRequests, WindowSize: time.Duration(pol.RateLimitWindowS) * time.Second}
			if !p.advancedRL.AllowWithQuota("tunnel:"+targetKey+":"+clientIP, 1, q) {
				statusCode = http.StatusTooManyRequests
				// Add CORS headers even on error responses to avoid breaking parallel fetch in browser
				if pol.CORSEnabled {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
				http.Error(w, "Rate limit exceeded", statusCode)
				p.logger.Warn("Per-tunnel rate limit exceeded", zap.String("target", targetKey))
				return
			}
		} else if !p.advancedRL.Allow(targetKey, 1) {
			statusCode = http.StatusTooManyRequests
			http.Error(w, "Rate limit exceeded (Gorenel)", statusCode)
			p.logger.Warn("Rate limit aşıldı", zap.String("target", targetKey))
			return
		}
	}

	if isCustom {
		p.logger.Info("HTTP istek", zap.String("type", "custom_domain"), zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.String("host", host))
	} else {
		p.logger.Info("HTTP istek", zap.String("type", "subdomain"), zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.String("subdomain", targetKey))
	}

	// Smart CORS handling
	if pol, ok := p.tunnelManager.GetTunnelPolicy(targetKey); ok && pol.CORSEnabled {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			statusCode = http.StatusNoContent
			w.WriteHeader(statusCode)
			return
		}
	}

	session, exists := p.tunnelManager.GetTunnel(host)
	if !exists {

		session, exists = p.tunnelManager.GetTunnel(targetKey)
	}

	if !exists {
		statusCode = http.StatusNotFound
		p.logger.Warn("Tunnel not found", zap.String("host", host))
		http.Error(w, "Tunnel bulunamadı", statusCode)

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
			p.triggerMLAnalysis(r.Method, r.URL.Path, r.Host, r.ContentLength, time.Since(startTime), statusCode, 0, clientIP, targetKey, "", aiMeta)
		}
		return
	}

	localHostPort := "localhost:3000"
	userID := ""
	if info, ok := p.tunnelManager.GetTunnelInfo(targetKey); ok && info != nil {
		if info.LocalPort > 0 {
			localHostPort = fmt.Sprintf("localhost:%d", info.LocalPort)
		}
		userID = info.UserID
	}

	// Phase 5: Request/Response Capture Preparation
	var reqBodyBuf bytes.Buffer
	if captureEnabled && r.Body != nil {
		bw := &BoundedWriter{W: &reqBodyBuf, Limit: p.inspectorMaxBodyBytes}
		r.Body = io.NopCloser(io.TeeReader(r.Body, bw))
	}

	captured := &CapturedRequest{
		ID:         uuid.New().String(),
		Subdomain:  targetKey,
		UserID:     userID,
		Method:     r.Method,
		Path:       r.URL.Path,
		ReqHeaders: r.Header,
		Timestamp:  startTime,
	}

	maxCaptureBytes := int64(0)
	if captureEnabled {
		maxCaptureBytes = p.inspectorMaxBodyBytes
	}
	captureWriter := NewResponseCaptureWriter(w, maxCaptureBytes)

	// --- REFAC: httputil.ReverseProxy implementation ---

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Outbound request is forwarded over yamux to the CLI, which dials the user's local port.
			// Use loopback + local port as Host so dev servers (Vite/WebSocket) accept the connection.
			publicHost := host
			req.URL.Scheme = "http"
			req.URL.Host = localHostPort
			req.Host = localHostPort

			// Force fresh HTML responses to allow transparent WebSocket patching (disable 304 cache for HTML)
			if strings.Contains(strings.ToLower(req.Header.Get("Accept")), "text/html") {
				req.Header.Del("If-None-Match")
				req.Header.Del("If-Modified-Since")
			}

			// Detect WebSocket upgrade requests
			isWebSocket := false
			if strings.EqualFold(req.Header.Get("Upgrade"), "websocket") {
				isWebSocket = true
			}

			// Set proxy headers
			if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
				req.Header.Set("X-Forwarded-For", prior+", "+clientIP)
			} else {
				req.Header.Set("X-Forwarded-For", clientIP)
			}

			// OPTIMIZATION: Next.js Server Actions and CSRF protection require Origin to match X-Forwarded-Host.
			// We set X-Forwarded-Host and X-Forwarded-Proto correctly and keep the original Origin header.
			// This "Transparent Proxy" approach is most compatible with modern frameworks.
			// ALWAYS overwrite these to match the public secure endpoint, regardless of upstream proxy settings.
			req.Header.Set("X-Forwarded-Host", publicHost)
			req.Header.Set("X-Forwarded-Proto", "https")
			req.Header.Set("X-Forwarded-Port", "443")

			// WebSocket Specific: httputil.ReverseProxy strips hop-by-hop headers.
			// Re-add them here so they reach the RoundTrip (and thus the tunnel target).
			if isWebSocket {
				req.Header.Set("Upgrade", "websocket")
				req.Header.Set("Connection", "upgrade")
				// REFIX: httputil.ReverseProxy will strip Upgrade/Connection after Director.
				// We set a custom header to signal RoundTrip to re-add them.
				req.Header.Set("X-Gorenel-Preserve-Upgrade", "websocket")

				// Rewrite Origin to match local Host to bypass Vite's strict CORS/WebSocket origin validation
				if req.Header.Get("Origin") != "" {
					req.Header.Set("Origin", "http://"+localHostPort)
				}
			}

			// Apply per-tunnel request shaping (moved from outside for cleaner structure)
			if pol, ok := p.tunnelManager.GetTunnelPolicy(host); ok {
				applyRequestPolicy(req, pol)
			} else if pol, ok := p.tunnelManager.GetTunnelPolicy(targetKey); ok {
				applyRequestPolicy(req, pol)
			}
		},
		Transport: &TunnelTransport{
			Session: session,
			Logger:  p.logger,
		},
		// CRITICAL: Disable response buffering. Without this, ReverseProxy buffers
		// the entire response body before sending anything to the browser.
		// Dev servers (Next.js, Vite) commonly use chunked transfer encoding or
		// omit Content-Length, causing the proxy to wait indefinitely for EOF.
		// -1 means "flush immediately after every Write".
		FlushInterval: -1,
		ModifyResponse: func(resp *http.Response) error {
			// Update local metrics
			statusCode = resp.StatusCode
			bodyLen := resp.ContentLength
			if bodyLen < 0 {
				bodyLen = 0
			}
			atomic.AddInt64(&bytesReceived, bodyLen)
			p.tunnelManager.UpdateStats(targetKey, reqBytes, bodyLen)

			// REFIX: Apply response policy directly to resp.Header.
			if pol, ok := p.tunnelManager.GetTunnelPolicy(targetKey); ok {
				applyResponseHeaders(resp.Header, pol)
			}

			// Rewrite Set-Cookie headers for domain/secure alignment
			rewriteCookies(resp, host)

			// Inject WebSocket HMR patch for Vite/Next.js/etc. in HTML responses
			if err := p.injectWebSocketPatch(resp); err != nil {
				p.logger.Warn("Failed to inject WebSocket HMR patch", zap.Error(err), zap.String("subdomain", targetKey))
			}

			// WebSocket Specific: httputil.ReverseProxy strips hop-by-hop headers from response.
			// Re-add them if we are switching protocols (WebSocket upgrade).
			if resp.StatusCode == http.StatusSwitchingProtocols {
				resp.Header.Set("Upgrade", "websocket")
				resp.Header.Set("Connection", "Upgrade")
			}

			// Finalize Capture metadata from response headers
			if captureEnabled {
				captured.ReqBody = reqBodyBuf.Bytes()
				captured.StatusCode = resp.StatusCode
				captured.RespHeaders = resp.Header

				// AI Analysis
				if p.aiAnalyzer != nil {
					aiMeta := p.aiAnalyzer.AnalyzeRequest(r.Host, r.URL.Path, captured.ReqBody)
					captured.AIMetadata = aiMeta
				}
			}

			return nil
		},
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			p.logger.Error("Proxy error", zap.Error(err), zap.String("subdomain", targetKey))
			statusCode = http.StatusBadGateway
			http.Error(rw, "Tunnel proxy error: "+err.Error(), statusCode)
		},
	}

	proxy.ServeHTTP(captureWriter, r)

	// Post-proxy capture finalization: capture response body and queue
	if captureEnabled && exists {
		captured.RespBody = captureWriter.Body.Bytes()
		captured.StatusCode = captureWriter.StatusCode
		captured.Duration = time.Since(startTime)

		select {
		case p.inspectQueue <- captured:
		default:
			atomic.AddInt64(&InspectorQueueDropped, 1)
		}

		p.triggerMLAnalysis(r.Method, r.URL.Path, r.Host, r.ContentLength, captured.Duration, captured.StatusCode, int64(len(captured.RespBody)), clientIP, targetKey, userID, captured.AIMetadata)
	}
}

func getClientIP(r *http.Request) string {
	// 1. Check Fly-Client-IP header
	if ip := strings.TrimSpace(r.Header.Get("Fly-Client-IP")); ip != "" {
		return ip
	}

	// 2. Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" && !isLocalIP(ip) {
				return ip
			}
		}
		// If all are local/private, fallback to the first one in the list
		if len(ips) > 0 {
			first := strings.TrimSpace(ips[0])
			if first != "" {
				return first
			}
		}
	}

	// 3. Fallback to RemoteAddr
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	return clientIP
}

func (p *HTTPProxy) triggerMLAnalysis(method, path, host string, requestSize int64, duration time.Duration, statusCode int, bytesReceived int64, clientIP string, targetKey string, userID string, aiMeta *AIMetadata) {
	if p.mlClient == nil {
		return
	}
	if p.env == "production" && shouldIgnoreAnomalyTraffic(clientIP) {
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
					UserID:       userID,
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

	userID := ""
	if info, ok := p.tunnelManager.GetTunnelInfo(subdomain); ok && info != nil {
		userID = info.UserID
	}

	event := NewRequestEvent(subdomain, userID, r.Method, r.URL.Path, r.UserAgent(), clientIP)
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

// TunnelTransport bridges Yamux streams with the standard http.RoundTripper
type TunnelTransport struct {
	Session *yamux.Session
	Logger  *zap.Logger
}

func (t *TunnelTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// REFIX: Re-add WebSocket headers if they were stripped by ReverseProxy
	if req.Header.Get("X-Gorenel-Preserve-Upgrade") == "websocket" {
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		req.Header.Del("X-Gorenel-Preserve-Upgrade")
	}

	stream, err := t.Session.Open()
	if err != nil {
		return nil, err
	}

	// httputil.ReverseProxy will write the request to this stream if it follows the HTTP protocol.
	if err := req.Write(stream); err != nil {
		stream.Close()
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(stream), req)
	if err != nil {
		stream.Close()
		return nil, err
	}

	// If it's a WebSocket upgrade (101), track the connection
	isWS := resp.StatusCode == http.StatusSwitchingProtocols
	if isWS {
		atomic.AddInt64(&WebSocketConnections, 1)
	}

	// Wrap response body to close the stream when done
	resp.Body = &StreamBody{
		ReadCloser: resp.Body,
		Stream:     stream,
		isWS:       isWS,
	}
	return resp, nil
}

type StreamBody struct {
	io.ReadCloser
	Stream net.Conn
	isWS   bool
}

// Write implements [io.Writer] so [httputil.ReverseProxy] can treat the 101 response body as
// [io.ReadWriteCloser] when switching protocols (WebSocket, etc.).
func (s *StreamBody) Write(p []byte) (int, error) {
	return s.Stream.Write(p)
}

func (s *StreamBody) Close() error {
	err := s.ReadCloser.Close()
	if s.isWS {
		atomic.AddInt64(&WebSocketConnections, -1)
	}
	s.Stream.Close()
	return err
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

func applyResponseHeaders(h http.Header, policy TunnelPolicy) {
	for k, v := range policy.AddResponseHeaders {
		if strings.TrimSpace(k) == "" {
			continue
		}
		h.Set(k, v)
	}
	for _, k := range policy.RemoveResponseHeaders {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		h.Del(k)
	}
}

func enforcePolicy(w http.ResponseWriter, r *http.Request, clientIP string, policy TunnelPolicy) (denied bool) {
	// HTTPS redirect
	if policy.HttpsRedirectEnabled {
		if strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))) == "http" {
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
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

func rewriteCookies(resp *http.Response, publicHost string) {
	cookies := resp.Header["Set-Cookie"]
	if len(cookies) == 0 {
		return
	}

	hostOnly := publicHost
	if idx := strings.Index(hostOnly, ":"); idx != -1 {
		hostOnly = hostOnly[:idx]
	}

	var newCookies []string
	for _, cookie := range cookies {
		parts := strings.Split(cookie, ";")
		hasSecure := false

		for i, part := range parts {
			trimmed := strings.TrimSpace(part)
			lower := strings.ToLower(trimmed)

			// If domain points to localhost/127.0.0.1, rewrite to the public domain name
			if strings.HasPrefix(lower, "domain=") {
				domVal := strings.TrimSpace(trimmed[7:])
				if strings.EqualFold(domVal, "localhost") || strings.EqualFold(domVal, "127.0.0.1") {
					parts[i] = " Domain=" + hostOnly
				}
			}

			if lower == "secure" {
				hasSecure = true
			}
		}

		// Ensure Secure flag is present for public HTTPS connections
		if !hasSecure {
			parts = append(parts, " Secure")
		}

		newCookies = append(newCookies, strings.Join(parts, ";"))
	}
	resp.Header["Set-Cookie"] = newCookies
}

func (p *HTTPProxy) injectWebSocketPatch(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || !strings.Contains(strings.ToLower(contentType), "text/html") {
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Check if gzip encoded
	var isGzipped bool
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Encoding")), "gzip") {
		isGzipped = true
	}

	var rawBytes []byte
	if isGzipped {
		reader, err := gzip.NewReader(bytes.NewReader(bodyBytes))
		if err != nil {
			rawBytes = bodyBytes
			isGzipped = false
		} else {
			rawBytes, err = io.ReadAll(reader)
			reader.Close()
			if err != nil {
				rawBytes = bodyBytes
				isGzipped = false
			}
		}
	} else {
		rawBytes = bodyBytes
	}

	domainMatch := "." + p.baseDomain
	jsCode := fmt.Sprintf(`<script id="gorenel-ws-patch">
(function() {
  var OrigWS = window.WebSocket;

  // Stub WebSocket for Vite HMR — never closes, prevents all reload loops
  function StubSocket(url) {
    this.url = url;
    this.readyState = 0;
    this.protocol = 'vite-hmr';
    this.extensions = '';
    this.bufferedAmount = 0;
    this.binaryType = 'blob';
    this.onopen = null;
    this.onmessage = null;
    this.onerror = null;
    this.onclose = null;
    this._ls = {};
    var self = this;
    setTimeout(function() {
      self.readyState = 1;
      var ev = new Event('open');
      self._emit('open', ev);
    }, 50);
  }
  StubSocket.prototype._emit = function(t, e) {
    if (this['on' + t]) try { this['on' + t](e); } catch(x) {}
    var a = this._ls[t] || [];
    for (var i = 0; i < a.length; i++) try { a[i].call(this, e); } catch(x) {}
  };
  StubSocket.prototype.addEventListener = function(t, fn) {
    if (!this._ls[t]) this._ls[t] = [];
    this._ls[t].push(fn);
  };
  StubSocket.prototype.removeEventListener = function(t, fn) {
    var a = this._ls[t] || [];
    this._ls[t] = a.filter(function(f) { return f !== fn; });
  };
  StubSocket.prototype.dispatchEvent = function(e) { this._emit(e.type, e); return true; };
  StubSocket.prototype.send = function() {};
  StubSocket.prototype.close = function() {
    if (this.readyState === 3) return;
    this.readyState = 3;
    this._emit('close', new CloseEvent('close', { wasClean: true, code: 1000, reason: '' }));
  };
  StubSocket.CONNECTING = 0;
  StubSocket.OPEN = 1;
  StubSocket.CLOSING = 2;
  StubSocket.CLOSED = 3;

  window.WebSocket = function(url, protocols) {
    // Detect Vite HMR socket by its sub-protocol
    var isVite = (protocols === 'vite-hmr') ||
                 (Array.isArray(protocols) && protocols.indexOf('vite-hmr') !== -1);
    if (isVite) {
      console.log('[Gorenel] Vite HMR stubbed — page is stable. Use localhost for live-reload.');
      return new StubSocket(url);
    }

    // Non-HMR sockets: rewrite URL if needed
    if (typeof url === 'string') {
      var isG = url.includes('%s');
      var isL = /localhost|127\.0\.0\.1|\[::1\]/.test(url);
      if ((isG || isL) && !/:(443|80)(\/|$)/.test(url)) {
        try {
          var u = new URL(url);
          u.protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
          u.host = location.host;
          url = u.toString();
        } catch(e) { url = url.replace(/:\d+/, ''); }
      }
    }
    return new OrigWS(url, protocols);
  };
  window.WebSocket.CONNECTING = 0;
  window.WebSocket.OPEN = 1;
  window.WebSocket.CLOSING = 2;
  window.WebSocket.CLOSED = 3;
  window.WebSocket.prototype = OrigWS.prototype;
})();
</script>`, domainMatch)
	patchScript := []byte(jsCode)

	// Inject right after <head> or <html> or at the very beginning of the document
	var newBytes []byte
	headIndex := bytes.Index(rawBytes, []byte("<head>"))
	if headIndex != -1 {
		insertPos := headIndex + len("<head>")
		newBytes = make([]byte, 0, len(rawBytes)+len(patchScript))
		newBytes = append(newBytes, rawBytes[:insertPos]...)
		newBytes = append(newBytes, patchScript...)
		newBytes = append(newBytes, rawBytes[insertPos:]...)
	} else {
		htmlIndex := bytes.Index(rawBytes, []byte("<html>"))
		if htmlIndex != -1 {
			insertPos := htmlIndex + len("<html>")
			newBytes = make([]byte, 0, len(rawBytes)+len(patchScript))
			newBytes = append(newBytes, rawBytes[:insertPos]...)
			newBytes = append(newBytes, patchScript...)
			newBytes = append(newBytes, rawBytes[insertPos:]...)
		} else {
			newBytes = append(patchScript, rawBytes...)
		}
	}

	// Re-compress if originally gzipped
	var finalBytes []byte
	if isGzipped {
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		_, err = writer.Write(newBytes)
		writer.Close()
		if err == nil {
			finalBytes = buf.Bytes()
		} else {
			finalBytes = newBytes
			resp.Header.Del("Content-Encoding")
		}
	} else {
		finalBytes = newBytes
	}

	resp.Body = io.NopCloser(bytes.NewReader(finalBytes))
	resp.Header.Set("Content-Length", strconv.Itoa(len(finalBytes)))
	return nil
}
