package server

import (
	"bytes"
	"net/http"
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
}

type TrafficInspector struct {
	mu      *sync.RWMutex
	history []*CapturedRequest
	maxSize int
}

func NewTrafficInspector(maxSize int) *TrafficInspector {
	return &TrafficInspector{
		history: make([]*CapturedRequest, 0, maxSize),
		maxSize: maxSize,
	}
}

func (ti *TrafficInspector) Record(req *CapturedRequest) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	if len(ti.history) >= ti.maxSize {
		ti.history = ti.history[1:]
	}
	ti.history = append(ti.history, req)
}

// bütün geçmişi getiren fonksiyon
func (ti *TrafficInspector) GetHistory() []*CapturedRequest {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	return ti.history
}

// ID ile spesifik req getirme
func (ti *TrafficInspector) GetByID(id string) *CapturedRequest {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	for _, r := range ti.history {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// geçmişte olan bir hatayı incelemeye almak için replay etmek
func (ti *TrafficInspector) Replay(id string, client http.Client, targetBase string) (*http.Response, error) {
	captured := ti.GetByID(id)
	if captured == nil {
		return nil, http.ErrNoLocation
	}

	req, err := http.NewRequest(captured.Method, targetBase+captured.Path, bytes.NewReader(captured.ReqBody))

	if err != nil {
		return nil, err
	}
	for k, vv := range captured.ReqHeaders {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	return client.Do(req)
}
