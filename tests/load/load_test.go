package load_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Bekican/gorenel/internal/server"
	"github.com/hashicorp/yamux"
	"github.com/stretchr/testify/assert"
)

// ===== 1000 CONCURRENT HTTP REQUESTS =====

func TestLoad_1000ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	var requestCount int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	numRequests := 1000
	var wg sync.WaitGroup
	var successCount, errorCount int64

	start := time.Now()

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 1000,
		},
	}

	// Limit concurrency to 50 workers to avoid thundering herd on Windows loopback
	concurrencyLimit := make(chan struct{}, 50)
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(idx int) {
			defer wg.Done()
			concurrencyLimit <- struct{}{}        // Acquire
			defer func() { <-concurrencyLimit }() // Release

			resp, err := client.Get(ts.URL)
			if err != nil {
				atomic.AddInt64(&errorCount, 1)
				return
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	t.Logf("📊 Load Test Results:")
	t.Logf("   Total Requests:  %d", numRequests)
	t.Logf("   Successful:      %d", successCount)
	t.Logf("   Errors:          %d", errorCount)
	t.Logf("   Duration:        %v", duration)
	t.Logf("   Throughput:      %.0f req/sec", float64(successCount)/duration.Seconds())

	assert.Equal(t, int64(numRequests), successCount, "All 1000 requests should succeed")
	assert.Equal(t, int64(0), errorCount, "No errors expected")
}

// ===== RATE LIMITER UNDER 10K REQUESTS =====

func TestLoad_RateLimiterUnder10KRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	rl := server.NewAdvancedRateLmiter()
	total := 10000
	var allowed, blocked int64

	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			if rl.Allow("stress-user") {
				atomic.AddInt64(&allowed, 1)
			} else {
				atomic.AddInt64(&blocked, 1)
			}
		}()
	}
	wg.Wait()

	t.Logf("📊 Rate Limiter Stress Test:")
	t.Logf("   Total Requests: %d", total)
	t.Logf("   Allowed:        %d", allowed)
	t.Logf("   Blocked:        %d", blocked)

	// Free tier = 100/hour, so out of 10K, ~100 should be allowed
	assert.LessOrEqual(t, allowed, int64(101), "Should not allow more than free tier limit")
	assert.Greater(t, blocked, int64(0), "Some requests should be blocked")
}

// ===== 100 CONCURRENT TUNNELS =====

func TestLoad_100ConcurrentTunnels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	tm := server.NewTunnelManager()
	numTunnels := 100
	sessions := make([]*yamux.Session, numTunnels)
	clients := make([]*yamux.Session, numTunnels)

	start := time.Now()

	// Register all tunnels
	var wg sync.WaitGroup
	wg.Add(numTunnels)
	for i := 0; i < numTunnels; i++ {
		go func(idx int) {
			defer wg.Done()
			sConn, cConn := net.Pipe()
			s, err := yamux.Server(sConn, yamux.DefaultConfig())
			if err != nil {
				return
			}
			c, err := yamux.Client(cConn, yamux.DefaultConfig())
			if err != nil {
				s.Close()
				return
			}
			sessions[idx] = s
			clients[idx] = c
			sub := fmt.Sprintf("load-%d", idx)
			tm.RegisterTunnel(sub, s, "", 3000+idx, fmt.Sprintf("http://%s.gorenel.site:8080", sub), "", "http")
		}(i)
	}
	wg.Wait()

	regDuration := time.Since(start)

	t.Logf("📊 Tunnel Capacity Test:")
	t.Logf("   Tunnels Registered:  %d", tm.Count())
	t.Logf("   Registration Time:   %v", regDuration)
	t.Logf("   Avg Per Tunnel:      %v", regDuration/time.Duration(numTunnels))

	assert.Equal(t, numTunnels, tm.Count(), "All 100 tunnels should be registered")

	// Verify all tunnels are reachable
	for i := 0; i < numTunnels; i++ {
		_, exists := tm.GetTunnel(fmt.Sprintf("load-%d", i))
		assert.True(t, exists, "Tunnel load-%d should be reachable", i)
	}

	// Cleanup
	for i := 0; i < numTunnels; i++ {
		if sessions[i] != nil {
			sessions[i].Close()
		}
		if clients[i] != nil {
			clients[i].Close()
		}
	}
}

// ===== ANALYTICS PIPELINE THROUGHPUT =====

func TestLoad_AnalyticsPipelineThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	ae := server.NewAnalyticsEngine(1 * time.Hour)
	numEvents := 50000

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(numEvents)
	for i := 0; i < numEvents; i++ {
		go func(idx int) {
			defer wg.Done()
			ae.Consume(&server.RequestEvent{
				Subdomain:    fmt.Sprintf("sub-%d", idx%100),
				Method:       "GET",
				Path:         fmt.Sprintf("/path/%d", idx%1000),
				StatusCode:   200,
				ResponseTime: time.Duration(idx%100) * time.Millisecond,
				GeoCountry:   "Turkey",
				Timestamp:    time.Now(),
			})
		}(i)
	}
	wg.Wait()

	duration := time.Since(start)
	snapshot := ae.GetSnapshot()

	t.Logf("📊 Analytics Pipeline Throughput:")
	t.Logf("   Events Processed: %d", snapshot.TotalRequests)
	t.Logf("   Duration:         %v", duration)
	t.Logf("   Throughput:       %.0f events/sec", float64(snapshot.TotalRequests)/duration.Seconds())

	assert.Equal(t, int64(numEvents), snapshot.TotalRequests, "All events should be processed")
}

// ===== MEMORY STABILITY TEST =====

func TestLoad_MemoryStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	ae := server.NewAnalyticsEngine(1 * time.Hour)
	ti := server.NewTrafficInspector(1000) // Capped ring buffer

	// Record initial memory
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Generate sustained load
	for round := 0; round < 10; round++ {
		for i := 0; i < 5000; i++ {
			ae.Consume(&server.RequestEvent{
				Subdomain:    fmt.Sprintf("sub-%d", i%50),
				Method:       "GET",
				Path:         fmt.Sprintf("/path/%d", i),
				StatusCode:   200,
				ResponseTime: 10 * time.Millisecond,
				Timestamp:    time.Now(),
			})
			ti.Record(&server.CapturedRequest{
				ID:     fmt.Sprintf("req-%d-%d", round, i),
				Method: "GET",
				Path:   fmt.Sprintf("/path/%d", i),
			})
		}
	}

	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	allocatedMB := float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024

	t.Logf("📊 Memory Stability Test:")
	t.Logf("   Before: %d MB", memBefore.Alloc/1024/1024)
	t.Logf("   After:  %d MB", memAfter.Alloc/1024/1024)
	t.Logf("   Delta:  %.2f MB", allocatedMB)

	// TrafficInspector is capped at 1000, so memory should be bounded
	history := ti.GetHistory()
	assert.LessOrEqual(t, len(history), 1000, "Inspector should cap at maxSize")

	// Memory increase should be reasonable (< 100MB for this workload)
	assert.Less(t, allocatedMB, float64(100), "Memory growth should be bounded")
}
