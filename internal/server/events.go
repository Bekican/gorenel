package server

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
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

// yeni event streeam oluşturuyoruz
func NewEventStream(bufferSize int) *EventStream {
	es := &EventStream{
		events:      make(chan *RequestEvent, bufferSize),
		subscribers: make([]EventConsumer, 0),
		done:        make(chan struct{}),
	}

	go es.dispatcher()

	return es
}

// event yayınlıyoruz
func (es *EventStream) publish(event *RequestEvent) {
	select {
	case es.events <- event:
		es.mu.Lock()
		es.TotalEvents++
		es.mu.Unlock()
	default:
		es.mu.Lock()
		es.DroppedEvents++
		es.mu.Unlock()
	}
}

// yeni consumer ekleme fonksiyonumuz
func (es *EventStream) Subscribe(consumer EventConsumer) {
	es.subMu.Lock()
	defer es.subMu.Unlock()

	es.subscribers = append(es.subscribers, consumer)
}

// dispatcher
func (es *EventStream) dispatcher() {
	for {
		select {
		case event := <-es.events:
			es.subMu.RLock()
			for _, consumer := range es.subscribers {
				go func(c EventConsumer, e *RequestEvent) {
					if err := c.Consume(e); err != nil {
						log.Printf("Error processing event: %v", err)
					}
				}(consumer, event)
			}
			es.subMu.RUnlock()

		case <-es.done:
			return
		}
	}
}

// streami kapatma fonksiyonu
func (es *EventStream) Close() {
	close(es.done)
	close(es.events)
}

// stream istatistiklerini gördüğümüz fonksiyon
func (es *EventStream) Stats() map[string]interface{} {
	es.mu.RLock()
	defer es.mu.RUnlock()

	es.subMu.RLock()
	defer es.subMu.RUnlock()

	return map[string]interface{}{
		"total_events":   es.TotalEvents,
		"dropped_events": es.DroppedEvents,
		"buffer_size":    cap(es.events),
		"buffer_usage":   len(es.events),
		"subscribers":    len(es.subscribers),
	}
}

// yeni request event oluşturuyoruz
func NewRequestEvent(subdomain, method, path, userAgent, clientIP string) *RequestEvent {
	return &RequestEvent{
		EventID:   generateEventID(),
		Timestamp: time.Now(),
		Subdomain: subdomain,
		Method:    method,
		Path:      path,
		UserAgent: userAgent,
		ClientIP:  clientIP,
	}
}

// event'i jsona çeviriyoruz
func (e *RequestEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func generateEventID() string {
	randomPart, err := randomstring(8)
	if err != nil {
		return time.Now().Format("20060102150405") + "-fallback"
	}
	return time.Now().Format("20060102150405") + "-" + randomPart
}

// bu fonksiyondaki güvenlik açığı tespit edildi,değiştirildi
func randomstring(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, n)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}
