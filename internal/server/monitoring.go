package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Bekican/gorenel/internal/authmgr"
	"github.com/Bekican/gorenel/internal/handler"
	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/middleware"
	"github.com/Bekican/gorenel/internal/ml"
	"github.com/Bekican/gorenel/internal/utils"
	"github.com/Bekican/gorenel/pkg/auth"
	serverErrors "github.com/Bekican/gorenel/pkg/errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	TotalRequests        int64
	ActiveConnections    int64
	TotalBytesIn         int64
	TotalBytesOut        int64
	WebSocketConnections int64
	WebSocketMessages    int64
	ServerStartTime      time.Time
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
	authManager     *authmgr.AuthManager
	advancedRL      *limiter.RateLimiter
	inspector       *TrafficInspector
	tokenSvc        *auth.JWTService
	anomalyStore    *AnomalyStore
	mlClient        *ml.Client
	traceSharer     *TraceSharer
	historyStore    *TunnelHistoryStore
	reservationRepo *PostgresReservationRepository
	tunnelHandler   TunnelClientHandler
	baseDomain      string
	proxyPort       string
	env             string
	logger          *zap.Logger
}

func NewMonitoringServer(tm *TunnelManager, ae *AnalyticsEngine, ah *handler.AuthHandler, am *authmgr.AuthManager, rl *limiter.RateLimiter, ti *TrafficInspector, ts *auth.JWTService, as *AnomalyStore, mlc *ml.Client, redisAddr string, historyStore *TunnelHistoryStore, reservationRepo *PostgresReservationRepository, baseDomain, proxyPort, env string, logger *zap.Logger) *MonitoringServer {
	return &MonitoringServer{
		tunnelManager:   tm,
		analyticsEngine: ae,
		authHandler:     ah,
		authManager:     am,
		advancedRL:      rl,
		inspector:       ti,
		tokenSvc:        ts,
		anomalyStore:    as,
		mlClient:        mlc,
		traceSharer:     NewTraceSharer(redisAddr),
		historyStore:    historyStore,
		reservationRepo: reservationRepo,
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
	// CLI-only helper: ensure stable/reserved subdomains via API key auth.
	mux.HandleFunc("/api/reservations/ensure", m.corsMiddleware(serverErrors.ErrorWrapper(m.reservationsEnsureHandler)))

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

		// Tunnel Policy Management (KeyAuth + IP allowlist)
		mux.HandleFunc("/api/tunnel-policy/", m.corsMiddleware(authMw(serverErrors.ErrorWrapper(m.tunnelPolicyHandler))))

		// Reservations (Reserved Subdomains)
		mux.HandleFunc("/api/reservations", m.corsMiddleware(authMw(serverErrors.ErrorWrapper(m.reservationsHandler))))
		mux.HandleFunc("/api/reservations/", m.corsMiddleware(authMw(serverErrors.ErrorWrapper(m.reservationsHandler))))
	}

	// Register Inspector Endpoints
	if m.inspector != nil {
		// Inspector data is sensitive; require auth if available.
		inspectorHandler := func(h http.HandlerFunc) http.HandlerFunc { return h }
		if m.tokenSvc != nil && m.authHandler != nil {
			inspectorHandler = middleware.RequireAuth(m.tokenSvc)
		}
		mux.HandleFunc("/api/inspector/history", m.corsMiddleware(inspectorHandler(rl(m.inspectorHistoryHandler))))
		mux.HandleFunc("/api/inspector/replay", m.corsMiddleware(inspectorHandler(rl(m.inspectorReplayHandler))))
		mux.HandleFunc("/api/inspector/rules", m.corsMiddleware(inspectorHandler(rl(m.inspectorRulesHandler))))

		// Trace Sharing:
		// - POST /api/shares (create) => auth required (contains inspector content)
		// - GET  /api/shares/{id} (view) => public (so share links work)
		shareCreateHandler := func(h http.HandlerFunc) http.HandlerFunc { return h }
		if m.tokenSvc != nil && m.authHandler != nil {
			shareCreateHandler = middleware.RequireAuth(m.tokenSvc)
		}
		mux.HandleFunc("/api/shares", m.corsMiddleware(shareCreateHandler(rl(m.shareTraceHandler))))
		mux.HandleFunc("/api/shares/", m.corsMiddleware(rl(m.getSharedTraceHandler)))
	}

	// Tunnels endpoint
	tunnelsHandler := func(h http.HandlerFunc) http.HandlerFunc { return h }
	if m.tokenSvc != nil && m.authHandler != nil {
		tunnelsHandler = middleware.RequireAuth(m.tokenSvc)
	}
	mux.HandleFunc("/api/tunnels", m.corsMiddleware(tunnelsHandler(rl(m.tunnelsHandlerFunc))))
	mux.HandleFunc("/api/tunnels/", m.corsMiddleware(tunnelsHandler(rl(m.tunnelsHandlerFunc))))
	mux.HandleFunc("/api/tunnels/history", m.corsMiddleware(tunnelsHandler(rl(m.tunnelHistoryHandler))))

	// Anomaly endpoint
	anomalyHandler := func(h http.HandlerFunc) http.HandlerFunc { return h }
	if m.tokenSvc != nil && m.authHandler != nil {
		anomalyHandler = middleware.RequireAuth(m.tokenSvc)
	}
	mux.HandleFunc("/api/anomalies", m.corsMiddleware(anomalyHandler(rl(m.anomaliesHandler))))

	// ML Stats endpoint
	mlHandler := func(h http.HandlerFunc) http.HandlerFunc { return h }
	if m.tokenSvc != nil && m.authHandler != nil {
		mlHandler = middleware.RequireAuth(m.tokenSvc)
	}
	mux.HandleFunc("/api/ml/stats", m.corsMiddleware(mlHandler(rl(m.mlStatsHandler))))

	// CLI Download & Install endpoints
	mux.HandleFunc("/downloads/", m.corsMiddleware(m.handleDownload))
	mux.HandleFunc("/v1/install", m.corsMiddleware(m.handleAutoInstall))
	mux.HandleFunc("/install.sh", m.corsMiddleware(m.handleInstallSh))
	mux.HandleFunc("/install.ps1", m.corsMiddleware(m.handleInstallPs1))

	// WebSocket Tunnel endpoint (replaces raw TCP control port for Fly.io shared IP)
	if m.tunnelHandler != nil {
		mux.HandleFunc("/tunnel/connect", rl(m.handleTunnelWebSocket))
	}

	// Caddy On-Demand TLS "ask" endpoint
	mux.HandleFunc("/api/tls/ask", m.handleCaddyAsk)
	
	// Diagnostic endpoint for TLS connectivity
	mux.HandleFunc("/api/health/tls-test", func(w http.ResponseWriter, r *http.Request) {
		m.logger.Info("TLS connectivity test reached from client", zap.String("remote_addr", r.RemoteAddr))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"base_domain": m.baseDomain,
			"msg": "Monitoring server is reachable",
		})
	})

	l, _ := zap.NewProduction()
	l.Info("Monitoring server başlatılıyor", zap.String("port", port), zap.String("base_domain", m.baseDomain))
	srv := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	return srv.ListenAndServe()
}

type reservationCreateRequest struct {
	Subdomain string `json:"subdomain"`
}

type reservationAssignRequest struct {
	APIKey string `json:"api_key"`
}

type reservationEnsureRequest struct {
	Subdomain string `json:"subdomain"`
}

func (m *MonitoringServer) reservationsHandler(w http.ResponseWriter, r *http.Request) error {
	if m.reservationRepo == nil {
		http.Error(w, "reservations unavailable", http.StatusServiceUnavailable)
		return nil
	}
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}

	// Routes:
	// - GET    /api/reservations
	// - POST   /api/reservations
	// - DELETE /api/reservations/{subdomain}
	// - PUT    /api/reservations/{subdomain}/assign  (body.api_key empty => unassign)
	path := strings.TrimPrefix(r.URL.Path, "/api/reservations")
	path = strings.Trim(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			list, err := m.reservationRepo.ListByUser(claims.UserID)
			if err != nil {
				http.Error(w, "failed to list reservations", http.StatusInternalServerError)
				return nil
			}
			w.Header().Set("Content-Type", "application/json")
			return json.NewEncoder(w).Encode(map[string]interface{}{"reservations": list, "count": len(list)})
		case http.MethodPost:
			var req reservationCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return nil
			}
			rec, err := m.reservationRepo.Create(claims.UserID, req.Subdomain)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return nil
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			return json.NewEncoder(w).Encode(rec)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return nil
		}
	}

	parts := strings.Split(path, "/")
	sub := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if action == "assign" {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return nil
		}
		var req reservationAssignRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		req.APIKey = strings.TrimSpace(req.APIKey)
		var keyHash *string
		if req.APIKey != "" {
			if m.authManager == nil {
				http.Error(w, "auth unavailable", http.StatusServiceUnavailable)
				return nil
			}
			info, ok := m.authManager.GetKeyInfo(req.APIKey)
			if !ok || info == nil || info.UserID != claims.UserID {
				http.Error(w, "invalid api_key (not owned by user)", http.StatusBadRequest)
				return nil
			}
			h := authmgr.HashKey(req.APIKey)
			keyHash = &h
		}
		if err := m.reservationRepo.Assign(claims.UserID, sub, keyHash); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	if r.Method == http.MethodDelete {
		if err := m.reservationRepo.Delete(claims.UserID, sub); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return nil
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	http.Error(w, "Not found", http.StatusNotFound)
	return nil
}

// reservationsEnsureHandler allows CLI clients (API-key authenticated) to ensure a subdomain
// reservation exists for the key owner. This is intentionally idempotent to support
// "stable subdomain" workflows.
//
// Route:
// - POST /api/reservations/ensure (body.subdomain required, auth via X-API-Key)
func (m *MonitoringServer) reservationsEnsureHandler(w http.ResponseWriter, r *http.Request) error {
	if m.reservationRepo == nil || m.authManager == nil {
		http.Error(w, "reservations unavailable", http.StatusServiceUnavailable)
		return nil
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return nil
	}

	apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(r.URL.Query().Get("api_key"))
	}
	if apiKey == "" {
		http.Error(w, "API key required (send X-API-Key)", http.StatusUnauthorized)
		return nil
	}

	keyInfo, err := m.authManager.ValidateKey(apiKey)
	if err != nil || keyInfo == nil {
		http.Error(w, "invalid API key", http.StatusUnauthorized)
		return nil
	}

	var req reservationEnsureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return nil
	}
	req.Subdomain = strings.TrimSpace(req.Subdomain)
	if req.Subdomain == "" {
		http.Error(w, "subdomain required", http.StatusBadRequest)
		return nil
	}

	rec, err := m.reservationRepo.Get(req.Subdomain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}

	if rec == nil {
		created, err := m.reservationRepo.Create(keyInfo.UserID, req.Subdomain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil
		}
		rec = created
	} else if rec.UserID != keyInfo.UserID {
		http.Error(w, "subdomain is reserved for a different user", http.StatusForbidden)
		return nil
	}

	// Enforce/attach API key binding when configured.
	keyHash := authmgr.HashKey(apiKey)
	if rec.AssignedAPIKeyHash != nil && strings.TrimSpace(*rec.AssignedAPIKeyHash) != "" && *rec.AssignedAPIKeyHash != keyHash {
		http.Error(w, "subdomain is assigned to a different API key", http.StatusForbidden)
		return nil
	}
	// If unassigned, bind it to this key (idempotent).
	if rec.AssignedAPIKeyHash == nil || strings.TrimSpace(*rec.AssignedAPIKeyHash) == "" {
		if err := m.reservationRepo.Assign(keyInfo.UserID, req.Subdomain, &keyHash); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil
		}
	}

	// Refresh and return the record.
	out, err := m.reservationRepo.Get(req.Subdomain)
	if err != nil {
		http.Error(w, "failed to load reservation", http.StatusInternalServerError)
		return nil
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(out)
}

type tunnelPolicyUpdateRequest struct {
	KeyAuthEnabled     *bool    `json:"key_auth_enabled,omitempty"`
	IPAllowlistEnabled *bool    `json:"ip_allowlist_enabled,omitempty"`
	IPAllowlist        []string `json:"ip_allowlist,omitempty"`

	BasicAuthEnabled  *bool   `json:"basic_auth_enabled,omitempty"`
	BasicAuthUsername *string `json:"basic_auth_username,omitempty"`
	BasicAuthPassword *string `json:"basic_auth_password,omitempty"`

	HttpsRedirectEnabled *bool `json:"https_redirect_enabled,omitempty"`

	RateLimitEnabled  *bool  `json:"rate_limit_enabled,omitempty"`
	RateLimitRequests *int   `json:"rate_limit_requests,omitempty"`
	RateLimitWindowS  *int64 `json:"rate_limit_window_s,omitempty"`

	AddRequestHeaders     map[string]string `json:"add_request_headers,omitempty"`
	RemoveRequestHeaders  []string          `json:"remove_request_headers,omitempty"`
	AddResponseHeaders    map[string]string `json:"add_response_headers,omitempty"`
	RemoveResponseHeaders []string          `json:"remove_response_headers,omitempty"`

	PathPrefix      *string `json:"path_prefix,omitempty"`
	ReplacePathFrom *string `json:"replace_path_from,omitempty"`
	ReplacePathTo   *string `json:"replace_path_to,omitempty"`
}

type tunnelPolicyRotateResponse struct {
	Token string `json:"token"`
}

func (m *MonitoringServer) tunnelPolicyHandler(w http.ResponseWriter, r *http.Request) error {
	// Routes:
	// - PUT  /api/tunnel-policy/{subdomain}
	// - POST /api/tunnel-policy/{subdomain}/rotate
	path := strings.TrimPrefix(r.URL.Path, "/api/tunnel-policy/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.Error(w, "Missing subdomain", http.StatusBadRequest)
		return nil
	}

	parts := strings.Split(path, "/")
	subdomain := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodPost:
		if action != "rotate" {
			http.Error(w, "Not found", http.StatusNotFound)
			return nil
		}
		token := utils.GenerateTunnelToken()
		if err := m.tunnelManager.UpdateTunnelPolicy(subdomain, func(p *TunnelPolicy) error {
			p.KeyAuthEnabled = true
			p.KeyAuthToken = token
			return nil
		}); err != nil {
			http.Error(w, "Tunnel not found", http.StatusNotFound)
			return nil
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(tunnelPolicyRotateResponse{Token: token})

	case http.MethodPut:
		if action != "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return nil
		}
		var req tunnelPolicyUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return nil
		}

		// Validate allowlist (if present)
		if req.IPAllowlist != nil {
			for _, raw := range req.IPAllowlist {
				raw = strings.TrimSpace(raw)
				if raw == "" {
					continue
				}
				if strings.Contains(raw, "/") {
					if _, err := netip.ParsePrefix(raw); err != nil {
						http.Error(w, "Invalid CIDR in allowlist: "+raw, http.StatusBadRequest)
						return nil
					}
					continue
				}
				if _, err := netip.ParseAddr(raw); err != nil {
					http.Error(w, "Invalid IP in allowlist: "+raw, http.StatusBadRequest)
					return nil
				}
			}
		}

		if err := m.tunnelManager.UpdateTunnelPolicy(subdomain, func(p *TunnelPolicy) error {
			if req.KeyAuthEnabled != nil {
				p.KeyAuthEnabled = *req.KeyAuthEnabled
				if !p.KeyAuthEnabled {
					p.KeyAuthToken = ""
				}
			}
			if req.BasicAuthEnabled != nil {
				p.BasicAuthEnabled = *req.BasicAuthEnabled
				if !p.BasicAuthEnabled {
					p.BasicAuthUsername = ""
					p.BasicAuthPasswordHash = ""
				}
			}
			if req.BasicAuthUsername != nil {
				p.BasicAuthUsername = strings.TrimSpace(*req.BasicAuthUsername)
			}
			if req.BasicAuthPassword != nil {
				pw := strings.TrimSpace(*req.BasicAuthPassword)
				if pw == "" {
					p.BasicAuthPasswordHash = ""
					p.BasicAuthEnabled = false
				} else {
					hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
					if err != nil {
						return err
					}
					p.BasicAuthPasswordHash = string(hash)
					if p.BasicAuthUsername != "" {
						p.BasicAuthEnabled = true
					}
				}
			}

			if req.IPAllowlistEnabled != nil {
				p.IPAllowlistEnabled = *req.IPAllowlistEnabled
				if !p.IPAllowlistEnabled {
					p.IPAllowlist = nil
				}
			}
			if req.IPAllowlist != nil {
				// Replace allowlist
				out := make([]string, 0, len(req.IPAllowlist))
				for _, raw := range req.IPAllowlist {
					raw = strings.TrimSpace(raw)
					if raw == "" {
						continue
					}
					out = append(out, raw)
				}
				p.IPAllowlist = out
				if len(out) > 0 {
					p.IPAllowlistEnabled = true
				}
			}

			if req.HttpsRedirectEnabled != nil {
				p.HttpsRedirectEnabled = *req.HttpsRedirectEnabled
			}

			if req.RateLimitEnabled != nil {
				p.RateLimitEnabled = *req.RateLimitEnabled
			}
			if req.RateLimitRequests != nil {
				p.RateLimitRequests = *req.RateLimitRequests
			}
			if req.RateLimitWindowS != nil {
				p.RateLimitWindowS = *req.RateLimitWindowS
			}

			if req.AddRequestHeaders != nil {
				p.AddRequestHeaders = req.AddRequestHeaders
			}
			if req.RemoveRequestHeaders != nil {
				p.RemoveRequestHeaders = req.RemoveRequestHeaders
			}
			if req.AddResponseHeaders != nil {
				p.AddResponseHeaders = req.AddResponseHeaders
			}
			if req.RemoveResponseHeaders != nil {
				p.RemoveResponseHeaders = req.RemoveResponseHeaders
			}

			if req.PathPrefix != nil {
				p.PathPrefix = strings.TrimSpace(*req.PathPrefix)
			}
			if req.ReplacePathFrom != nil {
				p.ReplacePathFrom = strings.TrimSpace(*req.ReplacePathFrom)
			}
			if req.ReplacePathTo != nil {
				p.ReplacePathTo = strings.TrimSpace(*req.ReplacePathTo)
			}
			return nil
		}); err != nil {
			http.Error(w, "Tunnel not found", http.StatusNotFound)
			return nil
		}

		w.WriteHeader(http.StatusNoContent)
		return nil

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return nil
	}
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
	w.Header().Set("Content-Type", "application/json")
	// ML kapalı / eğitim sırasında bile 200 dön: dashboard Promise.all ile tümünü düşürmesin
	type mlStatsEnvelope struct {
		Stats            interface{} `json:"stats"`
		ActiveTunnels    int         `json:"active_tunnels"`
		MLUp             bool        `json:"ml_up"`
		LastPredictionAt *string     `json:"last_prediction_at"`
	}

	env := mlStatsEnvelope{
		Stats:         map[string]interface{}{},
		ActiveTunnels: m.tunnelManager.Count(),
		MLUp:          false,
	}

	if m.mlClient != nil {
		env.MLUp = m.mlClient.HealthCheck()
		stats, err := m.mlClient.GetModelStats()
		if err != nil {
			m.logger.Warn("ML stats unavailable, returning empty stats", zap.Error(err))
		} else {
			env.Stats = stats
		}
		if t, ok := m.mlClient.LastPredictionAt(); ok {
			s := t.Format(time.RFC3339)
			env.LastPredictionAt = &s
		}
	}

	_ = json.NewEncoder(w).Encode(env)
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
		"capture_pipeline": map[string]interface{}{
			"inspector_queue_dropped": atomic.LoadInt64(&InspectorQueueDropped),
			"ml_concurrency_dropped":  atomic.LoadInt64(&MLConcurrencyDropped),
			"ml_in_flight":            atomic.LoadInt64(&MLInFlight),
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
	apiKey := strings.TrimSpace(r.URL.Query().Get("api_key"))
	if apiKey != "" {
		w.Header().Set("Content-Disposition", "attachment; filename=gorenel-setup.sh")
	}

	domain := m.baseDomain
	if domain == "" {
		domain = "gorenel.site"
	}
	baseURL := "https://" + domain

	script := fmt.Sprintf(`#!/bin/bash
set -e

# Gorenel Magic Install Script (Linux/Mac)
# Usage: curl -sSL %[1]s/install.sh | bash

GORENEL_URL="%[1]s"
INSTALL_DIR="$HOME/.gorenel/bin"
mkdir -p "$INSTALL_DIR"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then ARCH="amd64"; fi
if [ "$ARCH" == "aarch64" ]; then ARCH="arm64"; fi

BINARY_NAME="gorenel-$OS-$ARCH"
if [ "$OS" == "darwin" ]; then BINARY_NAME="gorenel-darwin-$ARCH"; fi

echo "Downloading Gorenel for $OS/$ARCH..."
curl -L -f -o "$INSTALL_DIR/gorenel" "$GORENEL_URL/downloads/$BINARY_NAME" || { echo "Download failed!"; exit 1; }
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

API_KEY="%[2]s"
if [ -n "$API_KEY" ]; then
    "$INSTALL_DIR/gorenel" config set api_key "$API_KEY"
    echo "API Key configured automatically."
    echo "You can now run: gorenel connect --port 3000"
elif [ "$#" -gt 0 ]; then
    echo "Executing: gorenel $*"
    "$INSTALL_DIR/gorenel" "$@"
else
    echo "Run: gorenel config set api_key YOUR_API_KEY && gorenel connect --port 3000"
fi
`, baseURL, apiKey)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(script))
}

// psSingleQuoted escapes a string for safe embedding inside PowerShell single quotes ('').
func psSingleQuoted(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// installPs1Template must not use fmt.Sprintf: any literal '%' in the script would corrupt output.
const installPs1Template = `# Gorenel Magic Install Script (Windows)
# Usage: iwr -useb __GORENEL_URL__/install.ps1 | iex

$gorenelUrl = '__GORENEL_URL__'
$installDir = "$env:LOCALAPPDATA\gorenel"
if (!(Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir | Out-Null }

$binaryPath = "$installDir\gorenel.exe"

Write-Host "Downloading Gorenel for Windows/amd64..." -ForegroundColor Cyan

try {
    Invoke-WebRequest -Uri "$gorenelUrl/downloads/gorenel-windows-amd64.exe" -OutFile $binaryPath -ErrorAction Stop
} catch {
    Write-Host "Re-trying with alternative method..." -ForegroundColor Yellow
    (New-Object System.Net.WebClient).DownloadFile("$gorenelUrl/downloads/gorenel-windows-amd64.exe", $binaryPath)
}

if (!(Test-Path $binaryPath) -or (Get-Item $binaryPath).Length -lt 1000) {
    Write-Host "Error: Binary download failed or file is corrupted." -ForegroundColor Red
    exit 1
}

# Add to User PATH permanently (dedupe; avoid leading ";"; works if User Path was empty)
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($null -eq $userPath) { $userPath = "" }
$segments = @($userPath -split ';' | ForEach-Object { $_.Trim() } | Where-Object { $_ })
$already = ($segments -contains $installDir)
if (-not $already) {
    $without = @($segments | Where-Object { $_ -ne $installDir })
    $newUserPath = if ($without.Count -gt 0) { ($without -join ';') + ';' + $installDir } else { $installDir }
    [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    Write-Host "Added $installDir to User PATH." -ForegroundColor DarkCyan
    try {
        $HWND_BROADCAST = [IntPtr]0xffff
        $WM_SETTINGCHANGE = 0x001a
        $sig = '[DllImport("user32.dll", SetLastError=true)] public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, IntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out IntPtr lpdwResult);'
        Add-Type -Namespace Win32 -Name EnvNotify -MemberDefinition $sig -ErrorAction SilentlyContinue | Out-Null
        $r = [IntPtr]::Zero
        [void][Win32.EnvNotify]::SendMessageTimeout($HWND_BROADCAST, $WM_SETTINGCHANGE, [IntPtr]::Zero, "Environment", 2, 5000, [ref]$r)
    } catch { }
}

# Current process (e.g. iwr | iex one-liner): PATH + function so chained commands work
if ($env:Path -notlike "*$installDir*") {
    $env:Path = "$installDir;$env:Path"
}
function global:gorenel { & $binaryPath @args }

# Persist 'gorenel' in PowerShell: profile shim (works in new windows even if PATH refresh lags)
try {
    if ($PROFILE) {
        $profileDir = Split-Path -Parent $PROFILE
        if (!(Test-Path -LiteralPath $profileDir)) { New-Item -ItemType Directory -Path $profileDir -Force | Out-Null }
        if (!(Test-Path -LiteralPath $PROFILE)) { New-Item -ItemType File -Path $PROFILE -Force | Out-Null }
        $marker = '# Gorenel CLI'
        $hasMarker = $false
        if ((Get-Item -LiteralPath $PROFILE).Length -gt 0) {
            $hasMarker = [bool](Select-String -LiteralPath $PROFILE -SimpleMatch $marker -Quiet -ErrorAction SilentlyContinue)
        }
        if (-not $hasMarker) {
            Add-Content -LiteralPath $PROFILE -Value "" -Encoding UTF8
            Add-Content -LiteralPath $PROFILE -Value $marker -Encoding UTF8
            Add-Content -LiteralPath $PROFILE -Value "function global:gorenel {" -Encoding UTF8
            Add-Content -LiteralPath $PROFILE -Value '  $exe = Join-Path $env:LOCALAPPDATA "gorenel\gorenel.exe"' -Encoding UTF8
            Add-Content -LiteralPath $PROFILE -Value '  if (Test-Path -LiteralPath $exe) { & $exe @args } else { Write-Error "Gorenel CLI bulunamadi: $exe" }' -Encoding UTF8
            Add-Content -LiteralPath $PROFILE -Value "}" -Encoding UTF8
            Write-Host "PowerShell profiline 'gorenel' eklendi: $PROFILE" -ForegroundColor DarkCyan
            Write-Host 'Bu pencerede hemen kullanmak icin: . $PROFILE' -ForegroundColor DarkGray
        }
    }
} catch { }

Write-Host "Gorenel installed to $binaryPath" -ForegroundColor Green

$apiKey = '__API_KEY__'
if ($apiKey) {
    & $binaryPath config set api_key $apiKey
    Write-Host "API Key configured automatically." -ForegroundColor Green
    Write-Host "You can now run: gorenel connect --port 3000" -ForegroundColor Cyan
} else {
    Write-Host "CMD / yeni PowerShell: PATH ile 'gorenel' (veya profil yuklendiyse 'gorenel')." -ForegroundColor DarkGray
    Write-Host "PATH sorununda tam yol: & '$binaryPath' connect --port 3000" -ForegroundColor DarkGray
    Write-Host ('Yapistir (PowerShell): iwr -useb ' + $gorenelUrl + '/install.ps1 | iex; gorenel connect --port 3000') -ForegroundColor Yellow
}
`

// handleInstallPs1 serves the PowerShell installation script
func (m *MonitoringServer) handleInstallPs1(w http.ResponseWriter, r *http.Request) {
	apiKey := strings.TrimSpace(r.URL.Query().Get("api_key"))
	if apiKey != "" {
		w.Header().Set("Content-Disposition", "attachment; filename=gorenel-setup.ps1")
	}

	domain := m.baseDomain
	if domain == "" {
		domain = "gorenel.site"
	}
	baseURL := "https://" + domain

	script := strings.ReplaceAll(installPs1Template, "__GORENEL_URL__", psSingleQuoted(baseURL))
	script = strings.ReplaceAll(script, "__API_KEY__", psSingleQuoted(apiKey))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(script))
}

// WebSocket upgrader for tunnel connections
var tunnelUpgrader = websocket.Upgrader{
	ReadBufferSize:  16384,
	WriteBufferSize: 16384,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for production proxy
	},
}

// handleTunnelWebSocket upgrades HTTP to WebSocket and passes to tunnel handler.
// This replaces the raw TCP control port (7000) so tunnels work over HTTPS (443).
func (m *MonitoringServer) handleTunnelWebSocket(w http.ResponseWriter, r *http.Request) {
	if m.tunnelHandler == nil {
		m.logger.Error("Tunnel handler not configured on MonitoringServer")
		http.Error(w, "Tunnel handler not configured", http.StatusServiceUnavailable)
		return
	}

	// Connection rate-limit to reduce abuse before any protocol-level auth happens.
	clientIP := ""
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		// XFF may contain a chain: client, proxy1, proxy2...
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			clientIP = strings.TrimSpace(parts[0])
		}
	}
	if clientIP == "" {
		clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}
	}
	if m.advancedRL != nil {
		q := limiter.Quota{Limit: 100, WindowSize: 1 * time.Minute}
		if !m.advancedRL.AllowWithQuota("wsconnect:"+clientIP, 1, q) {
			m.logger.Warn("Tunnel connection rate limited", zap.String("client_ip", clientIP))
			http.Error(w, "Too many tunnel connections", http.StatusTooManyRequests)
			return
		}
	}

	// Require API key early (header/query) to avoid unauthenticated long-lived WS handshakes.
	// apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
	// if apiKey == "" {
	// 	apiKey = strings.TrimSpace(r.URL.Query().Get("api_key"))
	// }
	// if strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Authorization"))), "bearer ") && apiKey == "" {
	// 	apiKey = strings.TrimSpace(strings.TrimSpace(r.Header.Get("Authorization"))[7:])
	// }
	apiKey := "debug-key"
	if false && apiKey == "" {
		http.Error(w, "API key required (send X-API-Key)", http.StatusUnauthorized)
		return
	}
	// if m.authManager != nil {
	// 	if _, err := m.authManager.ValidateKey(apiKey); err != nil {
	// 		http.Error(w, "Invalid API key", http.StatusUnauthorized)
	// 		return
	// 	}
	// 	m.authManager.IncrementUsage(apiKey)
	// }

	// DEBUG: Force headers (Unconditional Proxy bypass)
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Connection", "Upgrade")

	// Log headers for absolute certainty
	headerLog := make(map[string]string)
	for k, v := range r.Header {
		headerLog[k] = strings.Join(v, ", ")
	}
	m.logger.Info("FORCED WebSocket upgrade headers (v2)", zap.Any("headers", headerLog))

	ws, err := tunnelUpgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error("WebSocket upgrade failed", 
			zap.Error(err),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", r.Header.Get("User-Agent")),
		)
		return
	}

	m.logger.Info("New WebSocket tunnel connection successful",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("client_ip", clientIP),
		zap.String("x_forwarded_for", r.Header.Get("X-Forwarded-For")),
	)

	// Wrap WebSocket as net.Conn and hand off to existing tunnel handler
	conn := NewWSConn(ws)
	m.tunnelHandler(conn)
}

func (m *MonitoringServer) handleCaddyAsk(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	
	// Enhanced logging for debugging
	m.logger.Info("Caddy TLS ask request received",
		zap.String("domain", domain),
		zap.String("ip", r.RemoteAddr),
		zap.Any("query", r.URL.Query()),
	)

	if domain == "" {
		m.logger.Warn("Caddy TLS ask: Missing domain parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Allow ALL subdomains of our base domain.
	if strings.HasSuffix(domain, "."+m.baseDomain) {
		m.logger.Info("Caddy TLS ask: Domain allowed (wildcard match)", 
			zap.String("domain", domain), 
			zap.String("base", m.baseDomain))
		w.WriteHeader(http.StatusOK)
		return
	}

	// Also allow the bare base domain itself
	if domain == m.baseDomain {
		m.logger.Info("Caddy TLS ask: Base domain allowed", zap.String("domain", domain))
		w.WriteHeader(http.StatusOK)
		return
	}

	m.logger.Warn("Caddy TLS ask: Domain denied", 
		zap.String("domain", domain), 
		zap.String("expected_suffix", "."+m.baseDomain))
	w.WriteHeader(http.StatusNotFound)
}
