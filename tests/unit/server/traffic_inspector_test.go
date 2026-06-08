package server_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Bekican/gorenel/internal/server"
	"github.com/stretchr/testify/assert"
)

// ===== RECORD & RETRIEVE =====

func TestTrafficInspector_RecordAndGetHistory(t *testing.T) {
	ti := server.NewTrafficInspector(100)

	ti.Record(&server.CapturedRequest{
		ID:        "req-1",
		Subdomain: "test-sub",
		Method:    "GET",
		Path:      "/api/users",
		Timestamp: time.Now(),
	})

	history := ti.GetHistory()
	assert.Len(t, history, 1, "Should have 1 captured request")
	assert.Equal(t, "req-1", history[0].ID)
	assert.Equal(t, "GET", history[0].Method)
}

func TestTrafficInspector_GetByID(t *testing.T) {
	ti := server.NewTrafficInspector(100)

	ti.Record(&server.CapturedRequest{ID: "target-id", Method: "POST", Path: "/api/create"})
	ti.Record(&server.CapturedRequest{ID: "other-id", Method: "GET", Path: "/api/list"})

	found := ti.GetByID("target-id")
	assert.NotNil(t, found, "Should find request by ID")
	assert.Equal(t, "POST", found.Method)
	assert.Equal(t, "/api/create", found.Path)
}

func TestTrafficInspector_GetByID_NotFound(t *testing.T) {
	ti := server.NewTrafficInspector(100)

	found := ti.GetByID("ghost-id")
	assert.Nil(t, found, "Non-existent ID should return nil")
}

// ===== RING BUFFER OVERFLOW =====

func TestTrafficInspector_MaxSizeEviction(t *testing.T) {
	maxSize := 5
	ti := server.NewTrafficInspector(maxSize)

	// Record more than maxSize
	for i := 0; i < 10; i++ {
		ti.Record(&server.CapturedRequest{
			ID:   fmt.Sprintf("req-%d", i),
			Path: fmt.Sprintf("/path/%d", i),
		})
	}

	history := ti.GetHistory()
	assert.Len(t, history, maxSize, "History should not exceed maxSize")

	// Oldest entries should be evicted, newest remain
	assert.Equal(t, "req-5", history[0].ID, "First entry should be req-5 (req-0 to req-4 evicted)")
	assert.Equal(t, "req-9", history[maxSize-1].ID, "Last entry should be req-9")
}

// ===== REPLAY =====

func TestTrafficInspector_Replay(t *testing.T) {
	ti := server.NewTrafficInspector(100)

	// Record a captured request
	ti.Record(&server.CapturedRequest{
		ID:         "replay-me",
		Method:     "POST",
		Path:       "/api/submit",
		ReqHeaders: http.Header{"Content-Type": {"application/json"}},
		ReqBody:    []byte(`{"key":"value"}`),
	})

	// Create a mock target server
	received := make(chan *http.Request, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("replayed"))
	}))
	defer target.Close()

	resp, err := ti.Replay("replay-me", target.Client(), target.URL)
	assert.NoError(t, err, "Replay should succeed")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify the replayed request preserved original method and headers
	select {
	case req := <-received:
		assert.Equal(t, "POST", req.Method, "Replayed request should use original method")
		assert.Equal(t, "/api/submit", req.URL.Path)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for replayed request")
	}
}

func TestTrafficInspector_ReplayNotFound(t *testing.T) {
	ti := server.NewTrafficInspector(100)

	_, err := ti.Replay("ghost", &http.Client{}, "http://localhost")
	assert.Error(t, err, "Replay of non-existent ID should fail")
}

// ===== CONCURRENCY =====

func TestTrafficInspector_ConcurrentReadWrite(t *testing.T) {
	ti := server.NewTrafficInspector(100)
	var wg sync.WaitGroup

	// Concurrent writers
	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func(idx int) {
			defer wg.Done()
			ti.Record(&server.CapturedRequest{
				ID:   fmt.Sprintf("concurrent-%d", idx),
				Path: fmt.Sprintf("/path/%d", idx),
			})
		}(i)
	}

	// Concurrent readers
	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func() {
			defer wg.Done()
			_ = ti.GetHistory()
		}()
	}

	wg.Wait()

	history := ti.GetHistory()
	assert.LessOrEqual(t, len(history), 100, "Should not exceed capacity")
	assert.Greater(t, len(history), 0, "Should have some records")
}

// ===== MODIFIER INTEGRATION =====

func TestTrafficInspector_HasModifier(t *testing.T) {
	ti := server.NewTrafficInspector(100)

	mod := ti.GetModifier()
	assert.NotNil(t, mod, "Inspector should have a modifier attached")

	// Verify the modifier is functional
	mod.AddRule(server.ModificationRule{ID: "test", PathPattern: "/api/*"})
	rules := mod.GetRules()
	assert.Len(t, rules, 1)
}

// ===== BENCHMARK =====

func BenchmarkTrafficInspector_Record(b *testing.B) {
	ti := server.NewTrafficInspector(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ti.Record(&server.CapturedRequest{
			ID:     fmt.Sprintf("bench-%d", i),
			Method: "GET",
			Path:   "/bench",
		})
	}
}

func BenchmarkTrafficInspector_GetByID(b *testing.B) {
	ti := server.NewTrafficInspector(1000)
	for i := 0; i < 1000; i++ {
		ti.Record(&server.CapturedRequest{ID: fmt.Sprintf("bench-%d", i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ti.GetByID("bench-500")
	}
}
