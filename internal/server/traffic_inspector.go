package server

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CapturedRequest struct {
	ID          string        `json:"id"`
	Subdomain   string        `json:"subdomain"`
	Method      string        `json:"method"`
	Path        string        `json:"path"`
	ReqHeaders  http.Header   `json:"req_headers"`
	ReqBody     []byte        `json:"req_body"`
	RespHeaders http.Header   `json:"resp_headers"`
	RespBody    []byte        `json:"resp_body"`
	StatusCode  int           `json:"status_code"`
	Timestamp   time.Time     `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
	AIMetadata  *AIMetadata   `json:"ai_metadata,omitempty"`
}

type AIMetadata struct {
	Model      string `json:"model"`
	Provider   string `json:"provider"` // OpenAI, Anthropic, etc.
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Tokens     struct {
		Prompt     int `json:"prompt"`
		Completion int `json:"completion"`
		Total      int `json:"total"`
	} `json:"tokens"`
	IsSecurityRisk bool    `json:"is_security_risk"`
	RiskScore      float64 `json:"risk_score"`
	RiskReason     string  `json:"risk_reason"`
}

type TrafficInspector struct {
	mu       sync.RWMutex
	history  []*CapturedRequest
	maxSize  int
	modifier *TrafficModifier
}

func NewTrafficInspector(maxSize int) *TrafficInspector {
	return &TrafficInspector{
		history:  make([]*CapturedRequest, 0, maxSize),
		maxSize:  maxSize,
		modifier: NewTrafficModifier(),
	}
}

// GetModifier returns the attached traffic modifier
func (ti *TrafficInspector) GetModifier() *TrafficModifier {
	return ti.modifier
}

func (ti *TrafficInspector) Record(req *CapturedRequest) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	if len(ti.history) >= ti.maxSize {
		// Use copy instead of ti.history[1:] to avoid leaking the old backing array.
		// The simple slice shift (ti.history = ti.history[1:]) keeps the old first element
		// in the underlying array, preventing the GC from reclaiming it.
		copy(ti.history, ti.history[1:])
		ti.history[len(ti.history)-1] = nil // release pointer for GC
		ti.history = ti.history[:len(ti.history)-1]
	}
	ti.history = append(ti.history, req)
}

// bütün geçmişi getiren fonksiyon
func (ti *TrafficInspector) GetHistory() []*CapturedRequest {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	out := make([]*CapturedRequest, 0, len(ti.history))
	for _, r := range ti.history {
		if r == nil {
			continue
		}
		cp := *r
		out = append(out, &cp)
	}
	return out
}

// ID ile spesifik req getirme
func (ti *TrafficInspector) GetByID(id string) *CapturedRequest {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	for _, r := range ti.history {
		if r.ID == id {
			cp := *r
			return &cp
		}
	}
	return nil
}

// geçmişte olan bir hatayı incelemeye almak için replay etmek
func (ti *TrafficInspector) Replay(id string, client *http.Client, targetBase string) (*http.Response, error) {
	captured := ti.GetByID(id)
	if captured == nil {
		return nil, errors.New("captured request not found")
	}

	req, err := http.NewRequest(captured.Method, targetBase+captured.Path, bytes.NewReader(captured.ReqBody))

	if err != nil {
		return nil, err
	}
	for k, vv := range captured.ReqHeaders {
		// Avoid forwarding hop-by-hop and host-specific headers on replay.
		kl := strings.ToLower(k)
		if kl == "host" || kl == "connection" || kl == "content-length" || kl == "transfer-encoding" || kl == "keep-alive" || kl == "proxy-authenticate" || kl == "proxy-authorization" || kl == "te" || kl == "trailers" || kl == "upgrade" {
			continue
		}
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	return client.Do(req)
}
