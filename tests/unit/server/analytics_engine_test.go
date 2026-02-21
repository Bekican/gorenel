package server_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Bekican/gorenel/internal/server"
	"github.com/stretchr/testify/assert"
)

// ===== EVENT CONSUMPTION =====

func TestAnalyticsEngine_ConsumeEvent(t *testing.T) {
	ae := server.NewAnalyticsEngine(1 * time.Hour)

	err := ae.Consume(&server.RequestEvent{
		Subdomain:    "test-sub",
		Method:       "GET",
		Path:         "/api/users",
		UserAgent:    "Mozilla/5.0",
		ClientIP:     "192.168.1.1",
		StatusCode:   200,
		ResponseTime: 50 * time.Millisecond,
		BytesSent:    1024,
		GeoCountry:   "Turkey",
		Timestamp:    time.Now(),
	})

	assert.NoError(t, err, "Consume should not return error")

	snapshot := ae.GetSnapshot()
	assert.Equal(t, int64(1), snapshot.TotalRequests, "Should have 1 request")
}

func TestAnalyticsEngine_MultipleEvents(t *testing.T) {
	ae := server.NewAnalyticsEngine(1 * time.Hour)

	paths := []string{"/api/users", "/api/users", "/api/products", "/api/users", "/api/orders"}
	countries := []string{"Turkey", "Germany", "Turkey", "USA", "Turkey"}

	for i, path := range paths {
		ae.Consume(&server.RequestEvent{
			Subdomain:    "test-sub",
			Method:       "GET",
			Path:         path,
			UserAgent:    "TestAgent",
			StatusCode:   200,
			ResponseTime: time.Duration(10+i*5) * time.Millisecond,
			GeoCountry:   countries[i],
			Timestamp:    time.Now(),
		})
	}

	snapshot := ae.GetSnapshot()
	assert.Equal(t, int64(5), snapshot.TotalRequests)

	// Top paths should have /api/users at #1
	if len(snapshot.TopPaths) > 0 {
		assert.Equal(t, "/api/users", snapshot.TopPaths[0].Key, "Top path should be /api/users")
		assert.Equal(t, int64(3), snapshot.TopPaths[0].Count)
	}

	// Top countries should have Turkey at #1
	if len(snapshot.TopCountries) > 0 {
		assert.Equal(t, "Turkey", snapshot.TopCountries[0].Key)
		assert.Equal(t, int64(3), snapshot.TopCountries[0].Count)
	}
}

// ===== STATUS CODE DISTRIBUTION =====

func TestAnalyticsEngine_StatusCodeDistribution(t *testing.T) {
	ae := server.NewAnalyticsEngine(1 * time.Hour)

	statusCodes := []int{200, 200, 200, 404, 500}
	for _, code := range statusCodes {
		ae.Consume(&server.RequestEvent{
			Subdomain:    "test-sub",
			Method:       "GET",
			Path:         "/test",
			StatusCode:   code,
			ResponseTime: 10 * time.Millisecond,
			Timestamp:    time.Now(),
		})
	}

	snapshot := ae.GetSnapshot()
	assert.Equal(t, int64(3), snapshot.StatusCodeDist[200], "Should have 3x 200")
	assert.Equal(t, int64(1), snapshot.StatusCodeDist[404], "Should have 1x 404")
	assert.Equal(t, int64(1), snapshot.StatusCodeDist[500], "Should have 1x 500")
}

// ===== ENGINE NAME =====

func TestAnalyticsEngine_Name(t *testing.T) {
	ae := server.NewAnalyticsEngine(1 * time.Hour)
	assert.NotEmpty(t, ae.Name(), "Engine should have a name")
}

// ===== EMPTY SNAPSHOT =====

func TestAnalyticsEngine_EmptySnapshot(t *testing.T) {
	ae := server.NewAnalyticsEngine(1 * time.Hour)

	snapshot := ae.GetSnapshot()
	assert.Equal(t, int64(0), snapshot.TotalRequests, "Empty engine should have 0 requests")
	assert.NotNil(t, snapshot.StatusCodeDist, "Status code map should not be nil")
}

// ===== CONCURRENCY =====

func TestAnalyticsEngine_ConcurrentConsume(t *testing.T) {
	ae := server.NewAnalyticsEngine(1 * time.Hour)
	var wg sync.WaitGroup

	numEvents := 100
	wg.Add(numEvents)
	for i := 0; i < numEvents; i++ {
		go func(idx int) {
			defer wg.Done()
			ae.Consume(&server.RequestEvent{
				Subdomain:    fmt.Sprintf("sub-%d", idx%5),
				Method:       "GET",
				Path:         fmt.Sprintf("/path/%d", idx%10),
				StatusCode:   200,
				ResponseTime: time.Duration(idx) * time.Millisecond,
				GeoCountry:   "Turkey",
				Timestamp:    time.Now(),
			})
		}(i)
	}
	wg.Wait()

	snapshot := ae.GetSnapshot()
	assert.Equal(t, int64(numEvents), snapshot.TotalRequests, "Should count all concurrent events")
}

func TestAnalyticsEngine_ConcurrentConsumeAndRead(t *testing.T) {
	ae := server.NewAnalyticsEngine(1 * time.Hour)
	var wg sync.WaitGroup

	// Writers
	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func(idx int) {
			defer wg.Done()
			ae.Consume(&server.RequestEvent{
				Subdomain:    "sub",
				Method:       "GET",
				Path:         "/test",
				StatusCode:   200,
				ResponseTime: 10 * time.Millisecond,
				Timestamp:    time.Now(),
			})
		}(i)
	}

	// Readers
	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func() {
			defer wg.Done()
			snapshot := ae.GetSnapshot()
			// Just verify it doesn't panic
			_ = snapshot.TotalRequests
		}()
	}

	wg.Wait()

	snapshot := ae.GetSnapshot()
	assert.Equal(t, int64(50), snapshot.TotalRequests)
}

// ===== BENCHMARKS =====

func BenchmarkAnalyticsEngine_Consume(b *testing.B) {
	ae := server.NewAnalyticsEngine(1 * time.Hour)

	event := &server.RequestEvent{
		Subdomain:    "bench-sub",
		Method:       "GET",
		Path:         "/api/bench",
		UserAgent:    "BenchAgent/1.0",
		StatusCode:   200,
		ResponseTime: 10 * time.Millisecond,
		BytesSent:    512,
		GeoCountry:   "Turkey",
		Timestamp:    time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ae.Consume(event)
	}
}

func BenchmarkAnalyticsEngine_GetSnapshot(b *testing.B) {
	ae := server.NewAnalyticsEngine(1 * time.Hour)

	// Pre-populate with data
	for i := 0; i < 1000; i++ {
		ae.Consume(&server.RequestEvent{
			Subdomain:    fmt.Sprintf("sub-%d", i%10),
			Method:       "GET",
			Path:         fmt.Sprintf("/path/%d", i%50),
			StatusCode:   200,
			ResponseTime: time.Duration(i) * time.Millisecond,
			GeoCountry:   "Turkey",
			Timestamp:    time.Now(),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ae.GetSnapshot()
	}
}
