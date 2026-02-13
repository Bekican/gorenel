package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"go.uber.org/zap"
)

type AIProvider interface {
	GetName() string
	GetURL() string
	GetAUthHeader() string
}

type AICache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

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

type AnalyticsStore interface {
	Store(ctx context.Context, metrics *Analytics) error
}

type AIRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
}

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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		g.Logger.Error("Failed to request body", zap.Error(err))
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		analytics.ErrorOccurred = true
		analytics.ResponseCode = http.StatusBadRequest
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	//PII REDACTION
	redactBody := g.redactSensitiveData(body)

	cacheKey := generateCacheKey(redactBody, provider.GetName())

	//prompt cache
	if g.Cache != nil {
		if cachedResp, err := g.Cache(ctx, cacheKey); err == nil && cachedResp != nil {
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

	target, err := url.Parse(provider.GetURL())
	if err != nil {
		g.Logger.Error("Failed to parse provider URL", zap.Error(err))
		http.Error(w, "Invalid provider URL", http.StatusInternalServerError)
		analytics.ErrorOccurred = true
		analytics.ResponseCode = http.StatusInternalServerError
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	//cachleme ve analytic için response alma

}
