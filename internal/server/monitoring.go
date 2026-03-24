package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Bekican/gorenel/internal/handler"
	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/middleware"
	"github.com/Bekican/gorenel/internal/ml"
	"github.com/Bekican/gorenel/pkg/auth"
	serverErrors "github.com/Bekican/gorenel/pkg/errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
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

// TunnelClientHandler is the callback type for handling tunnel client connections.
type TunnelClientHandler func(conn net.Conn)

type MonitoringServer struct {
	tunnelManager   *TunnelManager
	analyticsEngine *AnalyticsEngine
	authHandler     *handler.AuthHandler
	advancedRL      *limiter.RateLimiter
	inspector       *TrafficInspector
	tokenSvc        *auth.JWTService
	anomalyStore    *AnomalyStore
	mlClient        *ml.Client
	traceSharer     *TraceSharer
	historyStore    *TunnelHistoryStore
	tunnelHandler   TunnelClientHandler
	baseDomain      string
	proxyPort       string
	env             string
	logger          *zap.Logger
}

func NewMonitoringServer(tm *TunnelManager, ae *AnalyticsEngine, ah *handler.AuthHandler, rl *limiter.RateLimiter, ti *TrafficInspector, ts *auth.JWTService, as *AnomalyStore, mlc *ml.Client, redisAddr string, historyStore *TunnelHistoryStore, baseDomain, proxyPort, env string, logger *zap.Logger) *MonitoringServer {
	return &MonitoringServer{
		tunnelManager:   tm,
		analyticsEngine: ae,
		authHandler:     ah,
		advancedRL:      rl,
		inspector:       ti,
		tokenSvc:        ts,
		anomalyStore:    as,
		mlClient:        mlc,
		traceSharer:     NewTraceSharer(redisAddr),
		historyStore:    historyStore,
		baseDomain:      baseDomain,
		proxyPort:       proxyPort,
		env:             env,
		logger:          logger,
	}
}

// SetTunnelHandler sets the callback function for handling tunnel client connections over WebSocket.
func (m *MonitoringServer) SetTunnelHandler(handler TunnelClientHandler) {
	m.tunnelHandler = handler
}

func (m *MonitoringServer) Start(port string) error {
	mux := http.NewServeMux()

	// Rate limit wrapper
	rl := middleware.RateLimitMiddleware(m.advancedRL)

	mux.HandleFunc("/health", m.corsMiddleware(m.healthHandler))
	mux.HandleFunc("/metrics", m.corsMiddleware(rl(m.metricsHandler)))
	mux.HandleFunc("/info", m.corsMiddleware(m.infoHandler))
	mux.HandleFunc("/analytics", m.corsMiddleware(rl(m.analyticsHandler)))
	mux.HandleFunc("/api/analytics/realtime", m.corsMiddleware(rl(m.realtimeAnalyticsHandler)))

	// Register Auth Endpoints with CORS
	if m.authHandler != nil {
		mux.HandleFunc("/api/login", m.corsMiddleware(serverErrors.ErrorWrapper(m.authHandler.Login)))
		mux.HandleFunc("/api/register", m.corsMiddleware(serverErrors.ErrorWrapper(m.authHandler.Register)))
		mux.HandleFunc("/api/callback", m.corsMiddleware(serverErrors.ErrorWrapper(m.authHandler.Callback)))
		mux.HandleFunc("/api/logout", m.corsMiddleware(serverErrors.ErrorWrapper(m.authHandler.Logout)))

		authMw := middleware.RequireAuth(m.tokenSvc)
		mux.HandleFunc("/api/me", m.corsMiddleware(authMw(serverErrors.ErrorWrapper(m.authHandler.Me))))

		// Key Management
		mux.HandleFunc("/api/keys", m.corsMiddleware(authMw(serverErrors.ErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
			switch r.Method {
			case http.MethodGet:
				return m.authHandler.ListAPIKeys(w, r)
			case http.MethodPost:
				return m.authHandler.CreateAPIKey(w, r)
			case http.MethodDelete:
				return m.authHandler.DeleteAPIKey(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return nil
			}
		}))))
	}

	// Register Inspector Endpoints
	if m.inspector != nil {
		mux.HandleFunc("/api/inspector/history", m.corsMiddleware(rl(m.inspectorHistoryHandler)))
		mux.HandleFunc("/api/inspector/replay", m.corsMiddleware(rl(m.inspectorReplayHandler)))
		mux.HandleFunc("/api/inspector/rules", m.corsMiddleware(rl(m.inspectorRulesHandler)))

		// Trace Sharing
		mux.HandleFunc("/api/shares", m.corsMiddleware(rl(m.shareTraceHandler)))
		mux.HandleFunc("/api/shares/", m.corsMiddleware(rl(m.getSharedTraceHandler)))
	}

	// Tunnels endpoint
	mux.HandleFunc("/api/tunnels", m.corsMiddleware(rl(m.tunnelsHandlerFunc)))
	mux.HandleFunc("/api/tunnels/", m.corsMiddleware(rl(m.tunnelsHandlerFunc)))
	mux.HandleFunc("/api/tunnels/history", m.corsMiddleware(rl(m.tunnelHistoryHandler)))

	// Anomaly endpoint
	mux.HandleFunc("/api/anomalies", m.corsMiddleware(rl(m.anomaliesHandler)))

	// ML Stats endpoint
	mux.HandleFunc("/api/ml/stats", m.corsMiddleware(rl(m.mlStatsHandler)))

	// CLI Download & Install endpoints
	mux.HandleFunc("/downloads/", m.corsMiddleware(m.handleDownload))
	mux.HandleFunc("/v1/install", m.corsMiddleware(m.handleAutoInstall))
	mux.HandleFunc("/install.sh", m.corsMiddleware(m.handleInstallSh))
	mux.HandleFunc("/install.ps1", m.corsMiddleware(m.handleInstallPs1))

	// WebSocket Tunnel endpoint (replaces raw TCP control port for Fly.io shared IP)
	if m.tunnelHandler != nil {
		mux.HandleFunc("/tunnel/connect", m.handleTunnelWebSocket)
	}

	l, _ := zap.NewProduction()
	l.Info("Monitoring server başlatılıyor", zap.String("port", port))
	return http.ListenAndServe(port, mux)
}

// corsMiddleware adds CORS headers to allow cross-origin requests
func (m *MonitoringServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. WWW to Apex Redirect (Production Only)
		if m.env == "production" && strings.HasPrefix(r.Host, "www.") {
			url := "https://gorenel.site" + r.URL.Path
			if r.URL.RawQuery != "" {
				url += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, url, http.StatusMovedPermanently)
			return
		}

		// Set specific origin for CORS with credentials
		origin := r.Header.Get("Origin")

		// Security: Production secure CORS whitelist
		isAllowed := false
		if origin != "" {
			// Localhost specifically for dev environment, disallow in production unless explicitly needed
			if m.env != "production" && (origin == "http://localhost" || strings.HasPrefix(origin, "http://localhost:") || origin == "http://127.0.0.1" || strings.HasPrefix(origin, "http://127.0.0.1:")) {
				isAllowed = true
			} else if strings.HasSuffix(origin, ".fly.dev") || strings.HasSuffix(origin, ".gorenel.site") || origin == "https://gorenel.site" {
				isAllowed = true
			} else if m.baseDomain != "" && (origin == "https://"+m.baseDomain || strings.HasSuffix(origin, "."+m.baseDomain)) {
				isAllowed = true
			}
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin != "" {
			// Log rejected origin for debugging
			m.logger.Warn("CORS request rejected: origin not in whitelist",
				zap.String("origin", origin),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path))

			// Disallow unauthorized origins in production
			if m.env == "production" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Security Headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// 🛡️ Content Security Policy (CSP)
		// Restrict scripts, styles and connections to trusted domains
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://fonts.googleapis.com; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://rsms.me; " +
			"font-src 'self' https://fonts.gstatic.com https://rsms.me; " +
			"connect-src 'self' wss://*.gorenel.site wss://*.fly.dev https://*.gorenel.site https://*.fly.dev; " +
			"img-src 'self' data: https:; "
		w.Header().Set("Content-Security-Policy", csp)

		// 🛡️ HTTP Strict Transport Security (HSTS) - Force HTTPS for 1 year
		if m.env == "production" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next(w, r)
	}
}

func (m *MonitoringServer) tunnelsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/tunnels/history") || r.URL.Query().Get("history") == "1" {
		m.tunnelHistoryHandler(w, r)
		return
	}
	tunnels := m.tunnelManager.GetTunnels()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tunnels": tunnels,
		"count":   len(tunnels),
	})
}

func (m *MonitoringServer) tunnelHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if m.historyStore == nil {
		http.Error(w, "tunnel history unavailable", http.StatusServiceUnavailable)
		return
	}
	records, err := m.historyStore.ListRecentSessions(100)
	if err != nil {
		http.Error(w, "failed to load tunnel history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": records,
		"count":    len(records),
	})
}

func (m *MonitoringServer) anomaliesHandler(w http.ResponseWriter, r *http.Request) {
	if m.anomalyStore == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"anomalies": []AnomalyRecord{},
			"count":     0,
		})
		return
	}

	anomalies := m.anomalyStore.GetRecent(50)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"anomalies": anomalies,
		"count":     len(anomalies),
	})
}

func (m *MonitoringServer) mlStatsHandler(w http.ResponseWriter, r *http.Request) {
	if m.mlClient == nil {
		http.Error(w, "ML client not initialized", http.StatusServiceUnavailable)
		return
	}

	stats, err := m.mlClient.GetModelStats()
	if err != nil {
		http.Error(w, "ML service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
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

// --- Inspector Handlers ---

func (m *MonitoringServer) inspectorHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if m.inspector == nil {
		http.Error(w, "Inspector not initialized", http.StatusServiceUnavailable)
		return
	}
	history := m.inspector.GetHistory()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (m *MonitoringServer) inspectorReplayHandler(w http.ResponseWriter, r *http.Request) {
	if m.inspector == nil {
		http.Error(w, "Inspector not initialized", http.StatusServiceUnavailable)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing request ID", http.StatusBadRequest)
		return
	}

	// Replay target: usually the local proxy or direct to client
	client := &http.Client{Timeout: 10 * time.Second}
	targetBase := "http://localhost" + m.proxyPort // Use dynamic proxy port

	resp, err := m.inspector.Replay(id, client, targetBase)
	if err != nil {
		http.Error(w, fmt.Sprintf("Replay failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (m *MonitoringServer) inspectorRulesHandler(w http.ResponseWriter, r *http.Request) {
	if m.inspector == nil || m.inspector.GetModifier() == nil {
		http.Error(w, "Inspector not initialized", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		rules := m.inspector.GetModifier().GetRules()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rules)
	case http.MethodPost:
		var rule ModificationRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		if rule.ID == "" {
			rule.ID = uuid.New().String()
		}
		m.inspector.GetModifier().AddRule(rule)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rule)
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing rule ID", http.StatusBadRequest)
			return
		}
		m.inspector.GetModifier().RemoveRule(id)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Trace Sharing Handlers ---

func (m *MonitoringServer) shareTraceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing request ID", http.StatusBadRequest)
		return
	}

	captured := m.inspector.GetByID(id)
	if captured == nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	shareID, err := m.traceSharer.Share(r.Context(), captured)
	if err != nil {
		http.Error(w, "Failed to share trace", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"share_id": shareID,
		"url":      fmt.Sprintf("/share/%s", shareID),
	})
}

func (m *MonitoringServer) getSharedTraceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path /api/shares/{id}
	path := r.URL.Path
	prefix := "/api/shares/"
	if len(path) < len(prefix) || path[:len(prefix)] != prefix {
		http.Error(w, "Invalid share path", http.StatusBadRequest)
		return
	}
	shareID := path[len(prefix):]
	if shareID == "" {
		http.Error(w, "Missing share ID", http.StatusBadRequest)
		return
	}

	captured, err := m.traceSharer.Get(r.Context(), shareID)
	if err != nil {
		http.Error(w, "Shared trace not found or expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(captured)
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

// handleDownload serves the CLI binary
func (m *MonitoringServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	m.logger.Info("Download request received", zap.String("path", r.URL.Path))

	fileName := strings.TrimPrefix(r.URL.Path, "/downloads/")
	if fileName == "" {
		http.NotFound(w, r)
		return
	}

	// Security: Prevent path traversal
	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		m.logger.Warn("Malicious download path detected", zap.String("path", r.URL.Path))
		http.Error(w, "Invalid path", http.StatusForbidden)
		return
	}

	filePath := fmt.Sprintf("./bin/%s", fileName)
	m.logger.Debug("Serving binary file", zap.String("file", filePath))

	// Set appropriate headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filePath)
}

// handleAutoInstall detects OS and serves the appropriate script or instructions
func (m *MonitoringServer) handleAutoInstall(w http.ResponseWriter, r *http.Request) {
	ua := strings.ToLower(r.Header.Get("User-Agent"))

	if strings.Contains(ua, "windows") {
		http.Redirect(w, r, "/install.ps1", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/install.sh", http.StatusTemporaryRedirect)
}

// handleInstallSh serves the Bash installation script
func (m *MonitoringServer) handleInstallSh(w http.ResponseWriter, r *http.Request) {
	script := `#!/bin/bash
set -e

# Gorenel Magic Install Script (Linux/Mac)
# Usage: curl -sSL https://gorenel.site/install.sh | bash

INSTALL_DIR="$HOME/.gorenel/bin"
mkdir -p "$INSTALL_DIR"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then ARCH="amd64"; fi
if [ "$ARCH" == "aarch64" ]; then ARCH="arm64"; fi

BINARY_NAME="gorenel-$OS-$ARCH"
if [ "$OS" == "darwin" ]; then BINARY_NAME="gorenel-darwin-$ARCH"; fi

echo "Downloading Gorenel for $OS/$ARCH..."
curl -L -f -o "$INSTALL_DIR/gorenel" "https://gorenel.site/downloads/$BINARY_NAME" || { echo "Download failed!"; exit 1; }
chmod +x "$INSTALL_DIR/gorenel"

# Add to PATH via shell profile
SHELL_PROFILE=""
if [ -f "$HOME/.zshrc" ]; then SHELL_PROFILE="$HOME/.zshrc"
elif [ -f "$HOME/.bashrc" ]; then SHELL_PROFILE="$HOME/.bashrc"
elif [ -f "$HOME/.profile" ]; then SHELL_PROFILE="$HOME/.profile"
fi

if [ -n "$SHELL_PROFILE" ] && ! grep -q ".gorenel/bin" "$SHELL_PROFILE" 2>/dev/null; then
    echo 'export PATH="$HOME/.gorenel/bin:$PATH"' >> "$SHELL_PROFILE"
    echo "Added to PATH in $SHELL_PROFILE (restart terminal or run: source $SHELL_PROFILE)"
fi
export PATH="$INSTALL_DIR:$PATH"

echo "Gorenel installed to $INSTALL_DIR/gorenel"

if [ "$#" -gt 0 ]; then
    echo "Executing: gorenel $*"
    "$INSTALL_DIR/gorenel" "$@"
else
    echo "Run: gorenel config set api_key YOUR_API_KEY && gorenel connect --port 3000"
fi
`
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(script))
}

// handleInstallPs1 serves the PowerShell installation script
func (m *MonitoringServer) handleInstallPs1(w http.ResponseWriter, r *http.Request) {
	script := `# Gorenel Magic Install Script (Windows)
# Usage: iwr -useb https://gorenel.site/install.ps1 | iex

$installDir = "$env:LOCALAPPDATA\gorenel"
if (!(Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir | Out-Null }

$binaryPath = "$installDir\gorenel.exe"

Write-Host "Downloading Gorenel for Windows/amd64..." -ForegroundColor Cyan

try {
    Invoke-WebRequest -Uri "https://gorenel.site/downloads/gorenel-windows-amd64.exe" -OutFile $binaryPath -ErrorAction Stop
} catch {
    Write-Host "Re-trying with alternative method..." -ForegroundColor Yellow
    (New-Object System.Net.WebClient).DownloadFile("https://gorenel.site/downloads/gorenel-windows-amd64.exe", $binaryPath)
}

if (!(Test-Path $binaryPath) -or (Get-Item $binaryPath).Length -lt 1000) {
    Write-Host "Error: Binary download failed or file is corrupted." -ForegroundColor Red
    exit 1
}

# Add to User PATH permanently if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added $installDir to User PATH (restart terminal for permanent effect)" -ForegroundColor DarkCyan
}

# Add to current session PATH
if ($env:Path -notlike "*$installDir*") {
    $env:Path = "$installDir;$env:Path"
}

# Set alias for current session
function global:gorenel { & $binaryPath @args }

Write-Host "Gorenel installed to $binaryPath" -ForegroundColor Green
Write-Host "Run: gorenel config set api_key YOUR_API_KEY; gorenel connect --port 3000" -ForegroundColor Yellow
`
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(script))
}

// WebSocket upgrader for tunnel connections
var tunnelUpgrader = websocket.Upgrader{
	ReadBufferSize:  16384,
	WriteBufferSize: 16384,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for CLI clients
	},
}

// handleTunnelWebSocket upgrades HTTP to WebSocket and passes to tunnel handler.
// This replaces the raw TCP control port (7000) so tunnels work over HTTPS (443).
func (m *MonitoringServer) handleTunnelWebSocket(w http.ResponseWriter, r *http.Request) {
	if m.tunnelHandler == nil {
		http.Error(w, "Tunnel handler not configured", http.StatusServiceUnavailable)
		return
	}

	ws, err := tunnelUpgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	m.logger.Info("New WebSocket tunnel connection",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("x_forwarded_for", r.Header.Get("X-Forwarded-For")),
	)

	// Wrap WebSocket as net.Conn and hand off to existing tunnel handler
	conn := NewWSConn(ws)
	m.tunnelHandler(conn)
}
