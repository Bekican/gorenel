package server

// import(
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"runtime"
// 	"sync/atomic"
// 	"time"
// )

// var(
// 	TotalRequests  int64
// 	ActiveConnections  int64
// 	TotalBytesIn  int64
// 	TotalBytesOut  int64
// 	ServerStartTime time.Time
// )

// func init(){
// 	ServerStartTime = time.Now()
// }

// type MonitoringServer struct{
// 	tunnelManager *TunnelManager
// }

// func NewMonitoringServer(tm *TunnelManager) *MonitoringServer{
// 	return &MonitoringServer{
// 		tunnelManager: tm,
// 	}
// }

// func(m *MonitoringServer) Start() error{
// 	mux := http.NewServeMux()

// 	mux.HandleFunc("/health",m.healthHandler)

// 	mux.HandleFunc("/metrics",m.metricsHandler)

// 	mux.HandleFunc("/info",m.infoHandler)

// 	log.Println("Monitoring serverı başlatılıyor: :9090")
// 	return http.ListenAndServe(":9090",mux)
// }

// //healthHandler
// func(m *MonitoringServer) healthHandler(w http.ResponseWriter , r *http.Request){
// 	uptime := time.Since(ServerStartTime)

// 	health := map[string]interface{}{
// 		"status" : "healthy",
// 		"uptime" : uptime.String(),
// 		"time" : time.Now().Format(time.RFC3339),
// 	}

// 	w.Header().Set("Content-Type","application/json")
// 	json.NewEncoder(w).Encode(health)
// }

// //metricsHandler
// func(m *MonitoringServer) metricsHandler(w http.ResponseWriter, r *http.Request){
// 	var memStats runtime.MemStats
// 	runtime.ReadMemStats(&memStats)

// 	metrics := map[string]interface{}{
// 		"tunnels" : map[string]interface{}{
// 			"active_count" : m.tunnelManager.Count(),
// 		},
// 		"requests" : map[string]interface{}{
// 			"total" :   atomic.LoadInt64(&TotalRequests),
// 			"active_connections" : atomic.LoadInt64(&ActiveConnections),
// 		},
// 		"bandwidth" : map[string]interface{}{
// 			"bytes_in" : atomic.LoadInt64(&TotalBytesIn),
// 			"bytes_out" : atomic.LoadInt64(&TotalBytesOut),
// 		},
// 		"system" : map[string]interface{}{
// 			"goroutines" : runtime.NumGoroutine(),
// 			"memory_alloc" : formatBytes(int64(memStats.Alloc)),
// 			"memory_sys" : formatBytes(int64(memStats.Sys)),
// 		},
// 		"uptime_seconds" : time.Since(ServerStartTime).Seconds(),
// 	}

// 	w.Header().Set("Content-Type","application/json")
// 	json.NewEncoder(w).Encode(metrics)
// }

// func(m *MonitoringServer) infoHandler(w http.ResponseWriter,r *http.Request){
// 	info := map[string]interface{}{
// 		"version" : "1.0.0",
// 		"go_version" : runtime.Version(),
// 		"platform" : fmt.Sprintf("%s/%s",runtime.GOOS,runtime.GOARCH),
// 		"start_time" : ServerStartTime.Format(time.RFC3339),
// 	}

// 	w.Header().Set("Content-Type","application/json")
// 	json.NewEncoder(w).Encode(info)
// }
