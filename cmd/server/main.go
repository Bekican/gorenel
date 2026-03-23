package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Bekican/gorenel/internal/analytics"
	"github.com/Bekican/gorenel/internal/authmgr"
	"github.com/Bekican/gorenel/internal/config"
	"github.com/Bekican/gorenel/internal/handler"
	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/ml"
	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/Bekican/gorenel/internal/server"
	"github.com/Bekican/gorenel/internal/utils"
	"github.com/Bekican/gorenel/pkg/auth"
	"github.com/Bekican/gorenel/pkg/logger"
	"github.com/hashicorp/yamux"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type MockOAuth struct{}

func (m *MockOAuth) GetAuthURL(state string) string {
	return "/api/callback?state=" + state + "&code=mock_code"
}
func (m *MockOAuth) GetUserProfile(code string) (*auth.UserProfile, error) {
	return &auth.UserProfile{
		Email:    "demo@gorenel.site",
		Name:     "Demo User",
		Provider: "mock",
	}, nil
}

func initOAuthProviders(logger *zap.Logger, cfg *config.Config) map[string]auth.OAuthProvider {
	providers := make(map[string]auth.OAuthProvider)

	if cfg.Env == "production" {
		// Google
		if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" && cfg.GoogleRedirectURL != "" {
			logger.Info("Google OAuth provider initialized", zap.String("client_id", cfg.GoogleClientID[:5]+"..."))
			providers["google"] = auth.NewGoogleOAuth(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
		}

		// GitHub
		if cfg.GithubClientID != "" && cfg.GithubClientSecret != "" && cfg.GithubRedirectURL != "" {
			logger.Info("GitHub OAuth provider initialized", zap.String("client_id", cfg.GithubClientID[:5]+"..."))
			providers["github"] = auth.NewGitHubOAuth(cfg.GithubClientID, cfg.GithubClientSecret, cfg.GithubRedirectURL)
		}

		if len(providers) == 0 {
			logger.Warn("GO_ENV=production but no OAuth providers (Google/GitHub) are configured.")
		}
		return providers
	}

	logger.Info("Mock OAuth provider initialized (dev mode)")
	providers["google"] = &MockOAuth{}
	providers["github"] = &MockOAuth{}
	return providers
}

func main() {
	// Initialize global logger
	logger.Init(logger.DefaultConfig())
	defer logger.Sync()

	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	zapLogger.Info("Gorenel Server başlatılıyor")

	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("Konfigürasyon yüklenemedi", zap.Error(err))
	}

	// Database / Persistence
	// 1. PostgreSQL
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		zapLogger.Fatal("PostgreSQL bağlantısı açılamadı", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		zapLogger.Warn("PostgreSQL ping başarısız", zap.Error(err))
	}

	// Optimize connection pool for production
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 2. Repositories
	userRepo := handler.NewPostgresUserRepository(db)
	if err := userRepo.Init(); err != nil {
		zapLogger.Error("PostgresUserRepository init hatası", zap.Error(err))
	}

	apiKeyRepo := server.NewPostgresAPIKeyRepository(db)
	if err := apiKeyRepo.Init(); err != nil {
		zapLogger.Error("PostgresAPIKeyRepository init hatası", zap.Error(err))
	}

	// 3. Analytics (ClickHouse)
	chRepo, err := analytics.NewClickHouseRepo(cfg.ClickHouseAddr, cfg.ClickHouseDB, cfg.ClickHouseUser, cfg.ClickHousePassword)
	if err != nil {
		zapLogger.Warn("Clickhouse bağlantısı kurulamadı, yüksek hacimli analiz kısıtlı olabilir", zap.Error(err))
	} else {
		if err := chRepo.InitSchema(); err != nil {
			zapLogger.Error("Clickhouse şema init hatası", zap.Error(err))
		}
	}

	// Core components
	tm := server.NewTunnelManager()
	authManager := authmgr.NewAuthManager(apiKeyRepo)

	// Database / Persistence (Redis)
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	// Ping Redis to ensure connection
	ctxPing, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisClient.Ping(ctxPing).Err(); err != nil {
		zapLogger.Warn("Redis bağlantısı kurulamadı, bazı özellikler kısıtlı olabilir", zap.Error(err))
	}
	cancel()

	// Initialize Advanced Rate Limiter (Redis-backed)
	rateLimiter := limiter.NewRateLimiter(redisClient, cfg.RateLimitRequests, cfg.RateLimitWindow)

	// Initialize Traffic Inspector
	inspector := server.NewTrafficInspector(cfg.InspectorHistorySize)

	//eventStreaming
	eventStream := server.NewEventStream(1000)
	defer eventStream.Close()
	//analytics engine
	analyticsEngine := server.NewAnalyticsEngine(24 * time.Hour)
	eventStream.Subscribe(analyticsEngine)

	//batch logger
	batchLogger, err := server.NewBatchLogger("./logs/batches", 1000, 5*time.Minute, chRepo)
	if err != nil {
		zapLogger.Fatal("Batch logger başlatılamadı", zap.Error(err))
	}
	defer batchLogger.Close()
	eventStream.Subscribe(batchLogger)

	//Data archive
	archiver, err := server.NewDataArchiver("./logs/archives", 1*time.Hour, 30)
	if err != nil {
		zapLogger.Fatal("Data archiver başlatılamadı", zap.Error(err))
	}
	defer archiver.Close()
	eventStream.Subscribe(archiver)

	//Geo location service
	geoLocator := server.NewGeoLocator(true)

	// Auth components
	jwtSvc := auth.NewJWTService(cfg.JWTSecret)

	oauthProviders := initOAuthProviders(zapLogger, cfg)
	authHandler := handler.NewAuthHandler(oauthProviders, jwtSvc, userRepo, authManager, cfg.Env == "production")

	// ML client uses the same shared logger

	// Initialize shared ML client
	mlClient := ml.NewClient(cfg.MLURL, zapLogger)

	// Proxy servers
	tcpProxy := server.NewTCPProxy()
	udpProxy := server.NewUDPProxy()
	anomalyStore := server.NewAnomalyStore(100) // Son 100 anomali kaydı
	httpProxy := server.NewHTTPProxy(tm, eventStream, geoLocator, rateLimiter, inspector, zapLogger, anomalyStore, mlClient, cfg.RedisAddr, cfg.BaseDomain, cfg.AcmeEmail, cfg.Env)

	go func() {
		zapLogger.Info("HTTP Proxy başlatılıyor", zap.String("port", cfg.ProxyPort))
		if err := httpProxy.Start(cfg.ProxyPort); err != nil {
			zapLogger.Fatal("HTTP Proxy hatası", zap.Error(err))
		}
	}()

	// Monitoring server
	monitor := server.NewMonitoringServer(tm, analyticsEngine, authHandler, rateLimiter, inspector, jwtSvc, anomalyStore, mlClient, cfg.RedisAddr, cfg.BaseDomain, cfg.ProxyPort, cfg.Env, zapLogger)

	// Set WebSocket tunnel handler - allows tunnel connections over HTTPS (replaces raw TCP for Fly.io shared IP)
	monitor.SetTunnelHandler(func(conn net.Conn) {
		handleClient(conn, tm, authManager, tcpProxy, udpProxy, zapLogger, cfg)
	})

	go func() {
		if err := monitor.Start(cfg.MonitorPort); err != nil {
			zapLogger.Fatal("Monitoring server hatası", zap.Error(err))
		}
	}()

	// Control Port'u dinle (Client'lar buraya bağlanacak)
	listener, err := net.Listen("tcp", cfg.ControlPort)
	if err != nil {
		zapLogger.Fatal("Port dinlenemedi", zap.Error(err))
	}
	defer listener.Close()

	zapLogger.Info("Gorenel Server hazır",
		zap.String("control_port", cfg.ControlPort),
		zap.String("proxy_port", cfg.ProxyPort),
		zap.String("monitoring_port", cfg.MonitorPort),
		zap.Bool("auth_enabled", true),
		zap.String("rate_limiter", "sliding_window"),
	)

	// Graceful Shutdown: sinyal dinleme
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Client handling (ayrı goroutine'de)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return // Shutdown sinyali geldi
				default:
					zapLogger.Error("Bağlantı hatası", zap.Error(err))
					continue
				}
			}

			// Security: Apply rate limiting to raw TCP connections
			clientIP, _, splitErr := net.SplitHostPort(conn.RemoteAddr().String())
			if splitErr != nil {
				clientIP = conn.RemoteAddr().String()
			}

			// Control Port connections are rare (only on client startup), so we limit token usage strictly
			if !rateLimiter.Allow("control_tcp_"+clientIP, 5) {
				zapLogger.Warn("Control Port TCP connection rate limit exceeded", zap.String("ip", clientIP))
				conn.Close()
				continue
			}

			zapLogger.Info("Yeni bağlantı", zap.String("remote_addr", conn.RemoteAddr().String()))
			go handleClient(conn, tm, authManager, tcpProxy, udpProxy, zapLogger, cfg)
		}
	}()

	// Ana goroutine sinyal bekler
	<-ctx.Done()
	zapLogger.Info("Shutdown sinyali alındı, kapatılıyor...")

	// Listener'ı kapat → accept loop durur
	listener.Close()

	// Defer'ler (eventStream.Close, batchLogger.Close, archiver.Close) burada çalışır
	zapLogger.Info("Gorenel Server kapatıldı")
}

func handleClient(conn net.Conn, tm *server.TunnelManager, authManager *authmgr.AuthManager, tcpProx *server.TCPProxy, udpProx *server.UDPProxy, logger *zap.Logger, cfg *config.Config) {
	defer conn.Close()

	// 1. REGISTER mesajını oku
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		logger.Error("Mesaj okunamadı", zap.Error(err))
		return
	}

	if msg.Type != protocol.MsgTypeRegister {
		logger.Warn("Beklenmeyen mesaj tipi", zap.String("type", msg.Type))
		errMsg := protocol.NewErrorMessage(400, "İlk mesaj REGISTER olmalı")
		if err := protocol.WriteMessage(conn, errMsg); err != nil {
			logger.Warn("Error message gönderilemedi", zap.Error(err))
		}
		return
	}

	logger.Info("REGISTER mesajı alındı")

	// Parse REGISTER request
	var regReq protocol.RegisterRequest
	if err := json.Unmarshal([]byte(msg.Payload), &regReq); err != nil {
		logger.Error("REGISTER parse edilemedi", zap.Error(err))
		errMsg := protocol.NewErrorMessage(400, "Invalid REGISTER payload")
		if err := protocol.WriteMessage(conn, errMsg); err != nil {
			logger.Warn("Error message gönderilemedi", zap.Error(err))
		}
		return
	}

	// AUTH: API key kontrolü
	authKey, err := authManager.ValidateKey(regReq.APIKey)
	if err != nil {
		logger.Warn("Authentication başarısız", zap.Error(err))
		errMsg := protocol.NewErrorMessage(401, fmt.Sprintf("Authentication failed: %v", err))
		if err := protocol.WriteMessage(conn, errMsg); err != nil {
			logger.Warn("Auth error message gönderilemedi", zap.Error(err))
		}
		return
	}

	apiKeyPrefix := regReq.APIKey
	if len(apiKeyPrefix) > 8 {
		apiKeyPrefix = apiKeyPrefix[:8]
	}
	logger.Info("Authentication başarılı", zap.String("api_key_prefix", apiKeyPrefix+"..."), zap.String("user_id", authKey.UserID))
	authManager.IncrementUsage(regReq.APIKey)

	// 4. Yamux Session başlat
	yamuxConfig := yamux.DefaultConfig()
	yamuxConfig.LogOutput = io.Discard
	yamuxConfig.EnableKeepAlive = true
	yamuxConfig.KeepAliveInterval = 30 * time.Second

	session, err := yamux.Server(conn, yamuxConfig)
	if err != nil {
		logger.Error("Yamux session başlatılamadı", zap.Error(err))
		return
	}
	defer session.Close()

	// 2. Tünel tipine göre işlem yap
	var subdomain string
	var fullURL string
	var publicPort int

	if regReq.TunnelType == "tcp" || regReq.TunnelType == "udp" {
		// Port tahsis et
		publicPort, err = tm.AllocatePort()
		if err != nil {
			logger.Error("Port tahsis hatası", zap.Error(err))
			if err := protocol.WriteMessage(conn, protocol.NewErrorMessage(500, "Boş port bulunamadı")); err != nil {
				logger.Warn("Port tahsis hatası response gönderilemedi", zap.Error(err))
			}
			return
		}
		defer tm.ReleasePort(publicPort, regReq.TunnelType)

		if regReq.TunnelType == "tcp" {
			if err := tcpProx.ListenAndForward(publicPort, session); err != nil {
				logger.Error("TCP Proxy hatası", zap.Error(err))
				if err := protocol.WriteMessage(conn, protocol.NewErrorMessage(500, "TCP Proxy başlatılamadı")); err != nil {
					logger.Warn("TCP proxy hata response gönderilemedi", zap.Error(err))
				}
				return
			}
		} else {
			if err := udpProx.ListenAndForward(publicPort, session); err != nil {
				logger.Error("UDP Proxy hatası", zap.Error(err))
				if err := protocol.WriteMessage(conn, protocol.NewErrorMessage(500, "UDP Proxy başlatılamadı")); err != nil {
					logger.Warn("UDP proxy hata response gönderilemedi", zap.Error(err))
				}
				return
			}
		}
		logger.Info("Tünel oluşturuldu", zap.String("type", strings.ToUpper(regReq.TunnelType)), zap.Int("port", publicPort))
	} else {
		// DEFAULT: HTTP Subdomain
		subdomain = utils.GenerateSubDomain(8)
		if cfg.Env == "production" || cfg.BaseDomain == "gorenel.site" {
			fullURL = fmt.Sprintf("https://%s.%s", subdomain, cfg.BaseDomain)
		} else {
			fullURL = fmt.Sprintf("http://%s.%s%s", subdomain, cfg.BaseDomain, cfg.ProxyPort)
		}
		logger.Info("HTTP Subdomain atandı", zap.String("url", fullURL))
	}

	// 3. Client'a yanıt ver
	respPayload := protocol.RegisterResponse{
		Subdomain:  subdomain,
		FullURL:    fullURL,
		PublicPort: publicPort,
	}
	respJson, err := json.Marshal(respPayload)
	if err != nil {
		logger.Error("REGISTERED payload marshal edilemedi", zap.Error(err))
		return
	}
	if err := protocol.WriteMessage(conn, protocol.Message{
		Type:    protocol.MsgTypeRegistered,
		Payload: string(respJson),
	}); err != nil {
		logger.Error("REGISTERED mesajı gönderilemedi", zap.Error(err))
		return
	}

	logger.Info("Yamux session başlatıldı", zap.String("subdomain", subdomain))

	// 5. Session'ı kaydet
	if subdomain != "" {
		tm.RegisterTunnel(subdomain, session, regReq.CustomDomain, regReq.LocalPort, fullURL)
		defer tm.RemoveTunnel(subdomain)
	} else {
		// TCP/UDP için de kaydetmek gerekebilir (opsiyonel, monitoring için iyi olur)
	}

	logger.Info("Aktif tünel sayısı", zap.Int("count", tm.Count()))

	// 6. Session kapanana kadar bekle
	for {
		if session.IsClosed() {
			<-session.CloseChan()
			logger.Info("Session kapandı", zap.String("subdomain", subdomain))
			return
		}
		_, err := session.AcceptStream()
		if err != nil {
			logger.Info("Client bağlantısı kesildi", zap.String("subdomain", subdomain))
			return
		}
	}
}
