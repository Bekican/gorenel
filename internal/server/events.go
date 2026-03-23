package server

import (
	"crypto/rand"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Bekican/gorenel/internal/protocol"
	"go.uber.org/zap"
)

// Re-export or use directly
type RequestEvent = protocol.RequestEvent

type consumerSlot struct {
	c  EventConsumer
	ch chan *RequestEvent
}

type EventStream struct {
	events chan *RequestEvent

	slots []consumerSlot
	subMu sync.RWMutex
	perCh int // buffer size per consumer queue

	TotalEvents   int64
	DroppedEvents int64
	mu            sync.RWMutex

	done   chan struct{}
	logger *zap.Logger
	closed atomic.Bool
}

type EventConsumer interface {
	Consume(event *RequestEvent) error
	Name() string
}

// NewEventStream creates a pub/sub stream with a bounded main buffer and per-consumer worker queues.
func NewEventStream(bufferSize int) *EventStream {
	l, _ := zap.NewProduction()
	perCh := 256
	if bufferSize > 0 && bufferSize < perCh {
		perCh = bufferSize
	}
	es := &EventStream{
		events: make(chan *RequestEvent, bufferSize),
		done:   make(chan struct{}),
		logger: l,
		perCh:  perCh,
	}

	go es.dispatcher()

	return es
}

func (es *EventStream) publish(event *RequestEvent) {
	if es.closed.Load() {
		return
	}
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

// Subscribe registers a consumer backed by its own buffered queue and a single worker goroutine.
func (es *EventStream) Subscribe(consumer EventConsumer) {
	ch := make(chan *RequestEvent, es.perCh)

	es.subMu.Lock()
	es.slots = append(es.slots, consumerSlot{c: consumer, ch: ch})
	es.subMu.Unlock()

	go es.consumerLoop(consumer, ch)
}

func (es *EventStream) consumerLoop(c EventConsumer, ch <-chan *RequestEvent) {
	for {
		select {
		case <-es.done:
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			if event == nil {
				continue
			}
			if err := c.Consume(event); err != nil {
				es.logger.Error("Event processing error", zap.String("consumer", c.Name()), zap.Error(err))
			}
		}
	}
}

func (es *EventStream) dispatcher() {
	for {
		select {
		case event, ok := <-es.events:
			if !ok {
				return
			}
			if event == nil {
				continue
			}
			es.subMu.RLock()
			for _, slot := range es.slots {
				select {
				case slot.ch <- event:
				default:
					es.mu.Lock()
					es.DroppedEvents++
					es.mu.Unlock()
				}
			}
			es.subMu.RUnlock()

		case <-es.done:
			return
		}
	}
}

func (es *EventStream) Close() {
	if !es.closed.CompareAndSwap(false, true) {
		return
	}
	close(es.done)
}

func (es *EventStream) Stats() map[string]interface{} {
	es.mu.RLock()
	defer es.mu.RUnlock()

	es.subMu.RLock()
	n := len(es.slots)
	es.subMu.RUnlock()

	return map[string]interface{}{
		"total_events":   es.TotalEvents,
		"dropped_events": es.DroppedEvents,
		"buffer_size":    cap(es.events),
		"buffer_usage":   len(es.events),
		"subscribers":    n,
	}
}

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

func generateEventID() string {
	randomPart, err := randomstring(8)
	if err != nil {
		return time.Now().Format("20060102150405") + "-fallback"
	}
	return time.Now().Format("20060102150405") + "-" + randomPart
}

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
