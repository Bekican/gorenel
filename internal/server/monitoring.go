package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

var (
	TotalRequests     int64
	ActiveConnections int64
	TotalBytesIn      int64
	TotalBytesOut     int64
	ServerStartTime   time.Time
)

func init() {
	ServerStartTime = time.Now()
}

type MonitoringServer struct {
	tunnelManager   *TunnelManager
	analyticsEngine *AnalyticsEngine
}

func NewMonitoringServer(tm *TunnelManager, ae *AnalyticsEngine) *MonitoringServer {
	return &MonitoringServer{
		tunnelManager:   tm,
		analyticsEngine: ae,
	}
}

func (m *MonitoringServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", m.corsMiddleware(m.healthHandler))

	mux.HandleFunc("/metrics", m.corsMiddleware(m.metricsHandler))

	mux.HandleFunc("/info", m.corsMiddleware(m.infoHandler))

	mux.HandleFunc("/analytics", m.corsMiddleware(m.analyticsHandler))

	mux.HandleFunc("/api/analytics/realtime", m.corsMiddleware(m.realtimeAnalyticsHandler))

	log.Println("Monitoring serverı başlatılıyor: :9090")
	return http.ListenAndServe(":9090", mux)
}

// corsMiddleware adds CORS headers to allow cross-origin requests
func (m *MonitoringServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin (for development)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next(w, r)
	}
}

// healthHandler -- healthCheck
func (m *MonitoringServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(ServerStartTime)

	health := map[string]interface{}{
		"status": "healthy",
		"uptime": uptime.String(),
		"time":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// metricsHandler
func (m *MonitoringServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := map[string]interface{}{
		"tunnels": map[string]interface{}{
			"active_count": m.tunnelManager.Count(),
		},
		"requests": map[string]interface{}{
			"total":              atomic.LoadInt64(&TotalRequests),
			"active_connections": atomic.LoadInt64(&ActiveConnections),
		},
		"bandwidth": map[string]interface{}{
			"bytes_in":  atomic.LoadInt64(&TotalBytesIn),
			"bytes_out": atomic.LoadInt64(&TotalBytesOut),
		},

		"websocket": map[string]interface{}{
			"connections": atomic.LoadInt64(&WebSocketConnections),
			"messages":    atomic.LoadInt64(&WebSocketMessages),
		},
		"system": map[string]interface{}{
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": formatBytes(int64(memStats.Alloc)),
			"memory_sys":   formatBytes(int64(memStats.Sys)),
		},
		"uptime_seconds": time.Since(ServerStartTime).Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// infoHandler
func (m *MonitoringServer) infoHandler(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"version":    "1.0.0",
		"go_version": runtime.Version(),
		"platform":   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		"start_time": ServerStartTime.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (m *MonitoringServer) analyticsHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
	<html>
<head>
    <title>Gorenel Analytics</title>
    <style>
        body { font-family: Arial; margin: 20px; background: #1a1a1a; color: #fff; }
        .card { background: #2d2d2d; padding: 20px; margin: 10px 0; border-radius: 8px; }
        .metric { font-size: 2em; color: #4CAF50; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #444; }
        th { background: #333; }
    </style>
</head>
<body>
    <h1>🚀 Gorenel Analytics Dashboard</h1>
    <div class="card">
        <h2>📊 Real-Time Stats</h2>
        <div id="stats">Loading...</div>
    </div>
    <script>
        async function loadStats() {
            const res = await fetch('/api/analytics/realtime');
            const data = await res.json();
            document.getElementById('stats').innerHTML = 
                '<p>Total Requests: <span class="metric">' + data.total_requests + '</span></p>' +
                '<p>Avg Response Time: <span class="metric">' + (data.avg_response_time_ms/1000000).toFixed(2) + ' ms</span></p>';
        }
        loadStats();
        setInterval(loadStats, 5000);
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (m *MonitoringServer) realtimeAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	if m.analyticsEngine == nil {
		http.Error(w, "Analytics engine not initialized", http.StatusServiceUnavailable)
		return
	}
	snapshot := m.analyticsEngine.GetSnapshot()
	json.NewEncoder(w).Encode(snapshot)
}

// increment request
func IncrementRequest() {
	atomic.AddInt64(&TotalRequests, 1)
}

// aktif bağlantı sayacı
func IncrementActiveConnections() {
	atomic.AddInt64(&ActiveConnections, 1)
}

// aktif bağlantı sayacı
func DecrementActiveConnections() {
	atomic.AddInt64(&ActiveConnections, -1)
}

// gelen bytelar
func AddBytesIn(bytes int64) {
	atomic.AddInt64(&TotalBytesIn, bytes)
}

// giden bytelar
func AddBytesOut(bytes int64) {
	atomic.AddInt64(&TotalBytesOut, bytes)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
