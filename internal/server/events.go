package server

import (
	"sync"
	"time"
)

//her http isteği için request oluşturacağımız event

type RequestEvent struct {
	EventID   string    `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`

	Subdomain string `json:"subdomain"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	UserAgent string `json:"user_agent"`

	ClientIP   string `json:"client_ip"`
	GeoCountry string `json:"geo_country,omitempty"`
	GeoCity    string `json:"geo_city,omitempty"`

	StatusCode    int           `json:"status_code"`
	ResponseTime  time.Duration `json:"response_time_ms"`
	ByteSent      int64         `json:"bytes_sent"`
	BytesReceived int64         `json:"bytes_received"`

	SessionID string `json:"session_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`

	IsWebSocket       bool          `json:"is_websocket"`
	WebSocketDuration time.Duration `json:"websocket_duration_ms,omitempty"`

	Error string `json:"error,omitempty"`
}

type EventStream struct {
	events chan *RequestEvent

	subscribers []EventConsumer
	subMu       sync.RWMutex

	TotalEvents   int64
	DroppedEvents int64
	mu            sync.RWMutex

	done chan struct{}
}

type EventConsumer interface {
	Consume(event *RequestEvent) error
	Name() string
}

//yeni event streeam oluşturuyoruz
