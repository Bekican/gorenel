package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
		Email:    "demo@gorenel.io",
		Name:     "Demo User",
		Provider: "mock",
	}, nil
}

func initOAuthProvider(logger *zap.Logger, cfg *config.Config) auth.OAuthProvider {
	if cfg.Env == "production" {
		if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" || cfg.GoogleRedirectURL == "" {
			logger.Warn("GO_ENV=production but GOOGLE_CLIENT_ID/SECRET/URL is missing. Falling back to MockOAuth for safety.")
			return &MockOAuth{}
		}

		logger.Info("Google OAuth provider initialized", zap.String("client_id", cfg.GoogleClientID[:5]+"..."))
		return auth.NewGoogleOAuth(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	}

	logger.Info("Mock OAuth provider initialized (dev mode)")
	return &MockOAuth{}
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

	// Core components
	tm := server.NewTunnelManager()
	authManager := server.NewAuthManager()

	// Initialize Advanced Rate Limiter
	rateLimiter := limiter.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)

	// Initialize Traffic Inspector
	inspector := server.NewTrafficInspector(cfg.InspectorHistorySize)

	//eventStreaming
	eventStream := server.NewEventStream(1000)
	defer eventStream.Close()
	//analytics engine
	analyticsEngine := server.NewAnalyticsEngine(24 * time.Hour)
	eventStream.Subscribe(analyticsEngine)

	//batch logger
	batchLogger, err := server.NewBatchLogger("./logs/batches", 1000, 5*time.Minute)
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

	// Database / Persistence
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	// Ping Redis to ensure connection
	ctxPing, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisClient.Ping(ctxPing).Err(); err != nil {
		zapLogger.Warn("Redis bağlantısı kurulamadı, bazı özellikler kısıtlı olabilir", zap.Error(err))
	}
	cancel()

	userRepo := handler.NewRedisUserRepository(redisClient)

	// Auth components
	jwtSvc := auth.NewJWTService(cfg.JWTSecret)

	oauthProvider := initOAuthProvider(zapLogger, cfg)
	authHandler := handler.NewAuthHandler(oauthProvider, jwtSvc, userRepo, cfg.Env == "production")

	// ML client uses the same shared logger

	// Initialize shared ML client
	mlClient := ml.NewClient(cfg.MLURL, zapLogger)

	// Proxy servers
	tcpProxy := server.NewTCPProxy()
	udpProxy := server.NewUDPProxy()
	anomalyStore := server.NewAnomalyStore(100) // Son 100 anomali kaydı
	httpProxy := server.NewHTTPProxy(tm, eventStream, geoLocator, rateLimiter, inspector, zapLogger, anomalyStore, mlClient, cfg.RedisAddr)

	go func() {
		zapLogger.Info("HTTP Proxy başlatılıyor", zap.String("port", cfg.ProxyPort))
		if err := httpProxy.Start(cfg.ProxyPort); err != nil {
			zapLogger.Fatal("HTTP Proxy hatası", zap.Error(err))
		}
	}()

	// Monitoring server
	monitor := server.NewMonitoringServer(tm, analyticsEngine, authHandler, rateLimiter, inspector, jwtSvc, anomalyStore, mlClient)
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

func handleClient(conn net.Conn, tm *server.TunnelManager, authManager *server.AuthManager, tcpProx *server.TCPProxy, udpProx *server.UDPProxy, logger *zap.Logger, cfg *config.Config) {
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
		protocol.WriteMessage(conn, errMsg)
		return
	}

	logger.Info("REGISTER mesajı alındı")

	// Parse REGISTER request
	var regReq protocol.RegisterRequest
	if err := json.Unmarshal([]byte(msg.Payload), &regReq); err != nil {
		logger.Error("REGISTER parse edilemedi", zap.Error(err))
		errMsg := protocol.NewErrorMessage(400, "Invalid REGISTER payload")
		protocol.WriteMessage(conn, errMsg)
		return
	}

	// AUTH: API key kontrolü
	authKey, err := authManager.ValidateKey(regReq.APIKey)
	if err != nil {
		logger.Warn("Authentication başarısız", zap.Error(err))
		errMsg := protocol.NewErrorMessage(401, fmt.Sprintf("Authentication failed: %v", err))
		protocol.WriteMessage(conn, errMsg)
		return
	}

	logger.Info("Authentication başarılı", zap.String("api_key_prefix", regReq.APIKey[:8]+"..."), zap.String("user_id", authKey.UserID))
	authManager.IncrementUsage(regReq.APIKey)

	// 4. Yamux Session başlat
	yamuxConfig := yamux.DefaultConfig()
	yamuxConfig.LogOutput = io.Discard

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
			protocol.WriteMessage(conn, protocol.NewErrorMessage(500, "Boş port bulunamadı"))
			return
		}

		if regReq.TunnelType == "tcp" {
			if err := tcpProx.ListenAndForward(publicPort, session); err != nil {
				logger.Error("TCP Proxy hatası", zap.Error(err))
				protocol.WriteMessage(conn, protocol.NewErrorMessage(500, "TCP Proxy başlatılamadı"))
				return
			}
		} else {
			if err := udpProx.ListenAndForward(publicPort, session); err != nil {
				logger.Error("UDP Proxy hatası", zap.Error(err))
				protocol.WriteMessage(conn, protocol.NewErrorMessage(500, "UDP Proxy başlatılamadı"))
				return
			}
		}
		logger.Info("Tünel oluşturuldu", zap.String("type", strings.ToUpper(regReq.TunnelType)), zap.Int("port", publicPort))
	} else {
		// DEFAULT: HTTP Subdomain
		subdomain = utils.GenerateSubDomain(8)
		fullURL = fmt.Sprintf("http://%s.%s%s", subdomain, protocol.BaseDomain, cfg.ProxyPort)
		logger.Info("HTTP Subdomain atandı", zap.String("url", fullURL))
	}

	// 3. Client'a yanıt ver
	respPayload := protocol.RegisterResponse{
		Subdomain:  subdomain,
		FullURL:    fullURL,
		PublicPort: publicPort,
	}
	respJson, _ := json.Marshal(respPayload)
	protocol.WriteMessage(conn, protocol.Message{
		Type:    protocol.MsgTypeRegistered,
		Payload: string(respJson),
	})

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
