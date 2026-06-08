package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bekican/gorenel/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestMonitoringServer sets up a minimal MonitoringServer for testing.
func createTestMonitoringServer(t *testing.T) *httptest.Server {
	t.Helper()
	tm := server.NewTunnelManager()
	ae := server.NewAnalyticsEngine(1 * time.Hour)

	mux := http.NewServeMux()

	// /health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// /metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tunnels": map[string]interface{}{
				"active_count": tm.Count(),
			},
			"requests": map[string]interface{}{
				"total": 0,
			},
		})
	})

	// /api/tunnels endpoint
	mux.HandleFunc("/api/tunnels", func(w http.ResponseWriter, r *http.Request) {
		tunnels := tm.GetTunnels()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   len(tunnels),
			"tunnels": tunnels,
		})
	})

	// /analytics/realtime endpoint
	mux.HandleFunc("/analytics/realtime", func(w http.ResponseWriter, r *http.Request) {
		snapshot := ae.GetSnapshot()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshot)
	})

	return httptest.NewServer(mux)
}

// ===== HEALTH ENDPOINT =====

func TestMonitoringAPI_Health(t *testing.T) {
	ts := createTestMonitoringServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err, "Health request should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "healthy", body["status"])
	assert.NotEmpty(t, body["time"])
}

// ===== METRICS ENDPOINT =====

func TestMonitoringAPI_Metrics(t *testing.T) {
	ts := createTestMonitoringServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	// Verify JSON structure
	tunnels, ok := body["tunnels"].(map[string]interface{})
	assert.True(t, ok, "Should have 'tunnels' key")
	assert.NotNil(t, tunnels["active_count"])

	requests, ok := body["requests"].(map[string]interface{})
	assert.True(t, ok, "Should have 'requests' key")
	assert.NotNil(t, requests["total"])
}

// ===== TUNNELS ENDPOINT =====

func TestMonitoringAPI_Tunnels_Empty(t *testing.T) {
	ts := createTestMonitoringServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/tunnels")
	require.NoError(t, err)
	defer resp.Body.Close()

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, float64(0), body["count"], "Should have 0 tunnels")
}

// ===== ANALYTICS REALTIME =====

func TestMonitoringAPI_Analytics_Realtime(t *testing.T) {
	ts := createTestMonitoringServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/analytics/realtime")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	assert.NotEmpty(t, bodyBytes, "Response should not be empty")

	var snapshot map[string]interface{}
	err = json.Unmarshal(bodyBytes, &snapshot)
	assert.NoError(t, err, "Response should be valid JSON")
}

// ===== CONCURRENT REQUESTS TO API =====

func TestMonitoringAPI_ConcurrentRequests(t *testing.T) {
	ts := createTestMonitoringServer(t)
	defer ts.Close()

	done := make(chan int, 50)
	for i := 0; i < 50; i++ {
		go func() {
			resp, err := http.Get(ts.URL + "/health")
			if err != nil {
				done <- 0
				return
			}
			resp.Body.Close()
			done <- resp.StatusCode
		}()
	}

	successCount := 0
	for i := 0; i < 50; i++ {
		status := <-done
		if status == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(t, 50, successCount, "All 50 concurrent requests should succeed")
}
