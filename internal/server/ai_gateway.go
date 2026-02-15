package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"time"

	"go.uber.org/zap"
)

type AIProvider interface {
	GetName() string
	GetURL() string
	GetAuthHeader() string
}

// AICache interface for Redis or other cache implementations
type AICache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

// Analytics tracks request metrics
type Analytics struct {
	RequestID     string
	Provider      string
	StartTime     time.Time
	EndTime       time.Time
	TokensUsed    int
	PromptTokens  int
	ResponseCode  int
	CacheHit      bool
	ErrorOccurred bool
}

type AIGateway struct {
	Providers       map[string]AIProvider
	Logger          *zap.Logger
	Cache           AICache
	DefaultProvider string
	AnalyticsStore  AnalyticsStore
}

// AnalyticsStore interface for storing metrics
type AnalyticsStore interface {
	Store(ctx context.Context, metrics *Analytics) error
}

// AIRequest represents the structure of AI API requests
type AIRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
}

// AIResponse represents the structure of AI API responses
type AIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (g *AIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := generateRequestID()

	// Initialize analytics
	analytics := &Analytics{
		RequestID: requestID,
		StartTime: time.Now(),
		CacheHit:  false,
	}
	defer func() {
		analytics.EndTime = time.Now()
		if g.AnalyticsStore != nil {
			if err := g.AnalyticsStore.Store(ctx, analytics); err != nil {
				g.Logger.Error("Failed to store analytics", zap.Error(err))
			}
		}
	}()

	// 1. PROVIDER RESOLUTION
	pName := r.URL.Query().Get("provider")
	if pName == "" && g.DefaultProvider != "" {
		pName = g.DefaultProvider
		g.Logger.Info("Using default provider", zap.String("provider", pName))
	}

	provider, ok := g.Providers[pName]
	if !ok {
		g.Logger.Error("Provider not found", zap.String("provider", pName))
		http.Error(w, fmt.Sprintf("Provider '%s' not found", pName), http.StatusNotFound)
		analytics.ErrorOccurred = true
		analytics.ResponseCode = http.StatusNotFound
		return
	}
	analytics.Provider = provider.GetName()

	// 2. REQUEST INTERCEPTION
	body, err := io.ReadAll(r.Body)
	if err != nil {
		g.Logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		analytics.ErrorOccurred = true
		analytics.ResponseCode = http.StatusBadRequest
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// PII REDACTION
	redactedBody := g.redactSensitiveData(body)

	// Generate cache key from redacted body
	cacheKey := generateCacheKey(redactedBody, provider.GetName())

	// PROMPT CACHE CHECK
	if g.Cache != nil {
		if cachedResp, err := g.Cache.Get(ctx, cacheKey); err == nil && cachedResp != nil {
			g.Logger.Info("Cache hit",
				zap.String("request_id", requestID),
				zap.String("provider", provider.GetName()))

			analytics.CacheHit = true
			analytics.ResponseCode = http.StatusOK

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache-Hit", "true")
			w.WriteHeader(http.StatusOK)
			w.Write(cachedResp)
			return
		}
	}

	// 3. PROXY SETUP
	target, err := url.Parse(provider.GetURL())
	if err != nil {
		g.Logger.Error("Failed to parse provider URL", zap.Error(err))
		http.Error(w, "Invalid provider URL", http.StatusInternalServerError)
		analytics.ErrorOccurred = true
		analytics.ResponseCode = http.StatusInternalServerError
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Capture response for caching and analytics
	var responseBuffer bytes.Buffer
	originalWriter := w

	proxy.ModifyResponse = func(resp *http.Response) error {
		analytics.ResponseCode = resp.StatusCode

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			g.Logger.Error("Failed to read response body", zap.Error(err))
			return err
		}
		resp.Body.Close()

		// ANALYTICS - Extract token usage
		if resp.StatusCode == http.StatusOK {
			var aiResp AIResponse
			if err := json.Unmarshal(respBody, &aiResp); err == nil {
				analytics.TokensUsed = aiResp.Usage.TotalTokens
				analytics.PromptTokens = aiResp.Usage.PromptTokens

				g.Logger.Info("Request completed",
					zap.String("request_id", requestID),
					zap.String("provider", provider.GetName()),
					zap.Int("total_tokens", aiResp.Usage.TotalTokens),
					zap.Int("prompt_tokens", aiResp.Usage.PromptTokens),
					zap.Int("completion_tokens", aiResp.Usage.CompletionTokens),
					zap.Duration("duration", time.Since(analytics.StartTime)))

				// CACHING - Store successful response
				if g.Cache != nil {
					cacheTTL := 1 * time.Hour // Configurable TTL
					if err := g.Cache.Set(ctx, cacheKey, respBody, cacheTTL); err != nil {
						g.Logger.Warn("Failed to cache response", zap.Error(err))
					} else {
						g.Logger.Debug("Response cached successfully", zap.String("cache_key", cacheKey))
					}
				}
			}
		}

		// Restore response body for downstream
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

		// Also write to buffer for potential retry logic
		responseBuffer.Write(respBody)

		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		g.Logger.Error("Proxy error",
			zap.String("request_id", requestID),
			zap.String("provider", provider.GetName()),
			zap.Error(err))

		analytics.ErrorOccurred = true
		analytics.ResponseCode = http.StatusBadGateway

		http.Error(w, fmt.Sprintf("Gateway error: %v", err), http.StatusBadGateway)
	}

	// 4. HEADER INJECTION
	r.Header.Set("Authorization", provider.GetAuthHeader())
	r.Header.Set("X-Request-ID", requestID)
	r.Host = target.Host
	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme

	// Update request body with redacted content
	r.Body = io.NopCloser(bytes.NewBuffer(redactedBody))
	r.ContentLength = int64(len(redactedBody))

	g.Logger.Info("Proxying request",
		zap.String("request_id", requestID),
		zap.String("provider", provider.GetName()),
		zap.String("target", target.String()))

	proxy.ServeHTTP(originalWriter, r)
}

// redactSensitiveData removes PII from request body
func (g *AIGateway) redactSensitiveData(body []byte) []byte {
	bodyStr := string(body)

	// Email redaction
	emailRegex := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	bodyStr = emailRegex.ReplaceAllString(bodyStr, "[EMAIL_REDACTED]")

	// Phone number redaction (US format)
	phoneRegex := regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`)
	bodyStr = phoneRegex.ReplaceAllString(bodyStr, "[PHONE_REDACTED]")

	// SSN redaction
	ssnRegex := regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	bodyStr = ssnRegex.ReplaceAllString(bodyStr, "[SSN_REDACTED]")

	// Credit card redaction (basic pattern)
	ccRegex := regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`)
	bodyStr = ccRegex.ReplaceAllString(bodyStr, "[CC_REDACTED]")

	// IP address redaction
	ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	bodyStr = ipRegex.ReplaceAllString(bodyStr, "[IP_REDACTED]")

	g.Logger.Debug("PII redaction applied",
		zap.Int("original_size", len(body)),
		zap.Int("redacted_size", len(bodyStr)))

	return []byte(bodyStr)
}

// generateCacheKey creates a cache key from request body and provider
func generateCacheKey(body []byte, provider string) string {
	return fmt.Sprintf("ai_cache:%s:%x", provider, body)
}

// generateRequestID creates a unique request identifier
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
