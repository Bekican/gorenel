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
	tunnelManager *TunnelManager
}

func NewMonitoringServer(tm *TunnelManager) *MonitoringServer {
	return &MonitoringServer{
		tunnelManager: tm,
	}
}

func (m *MonitoringServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", m.healthHandler)

	mux.HandleFunc("/metrics", m.metricsHandler)

	mux.HandleFunc("/info", m.infoHandler)

	log.Println("Monitoring serverı başlatılıyor: :9090")
	return http.ListenAndServe(":9090", mux)
}

// healthHandler
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

func IncrementRequest() {
	atomic.AddInt64(&TotalRequests, 1)
}

func IncrementActiveConnections() {
	atomic.AddInt64(&ActiveConnections, 1)
}

func DecrementActiveConnections() {
	atomic.AddInt64(&ActiveConnections, -1)
}

func AddBytesIn(bytes int64) {
	atomic.AddInt64(&TotalBytesIn, bytes)
}

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
