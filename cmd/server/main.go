package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/Bekican/gorenel/internal/handler"
	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/Bekican/gorenel/internal/server"
	"github.com/Bekican/gorenel/internal/utils"
	"github.com/Bekican/gorenel/pkg/auth"
	"github.com/Bekican/gorenel/pkg/logger"
	"github.com/hashicorp/yamux"
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

func main() {
	// Initialize global logger
	logger.Init(logger.DefaultConfig())
	defer logger.Sync()

	log.Println(" Gorenel Server başlatılıyor...")

	// Core components
	tm := server.NewTunnelManager()
	authManager := server.NewAuthManager()

	// Initialize Advanced Rate Limiter (60 req/min default)
	rateLimiter := limiter.NewRateLimiter(60, 1*time.Minute)

	// Initialize Traffic Inspector (keep last 100 requests)
	inspector := server.NewTrafficInspector(100)

	//eventStreaming
	eventStream := server.NewEventStream(1000)
	defer eventStream.Close()
	//analytics engine
	analyticsEngine := server.NewAnalyticsEngine(24 * time.Hour)
	eventStream.Subscribe(analyticsEngine)

	//batch logger
	batchLogger, err := server.NewBatchLogger("./logs/batches", 1000, 5*time.Minute)
	if err != nil {
		log.Fatalf("Batch logger başlatılamadı : %v", err)
	}
	defer batchLogger.Close()
	eventStream.Subscribe(batchLogger)

	//Data archive
	archiver, err := server.NewDataArchiver("./logs/archives", 1*time.Hour, 30)
	if err != nil {
		log.Fatalf("Data archiver başlatılamadı : %v", err)
	}
	defer archiver.Close()
	eventStream.Subscribe(archiver)

	//Geo location service
	geoLocator := server.NewGeoLocator(true)

	// Auth components
	jwtSvc := auth.NewJWTService("SUPER_SECRET_KEY_CHANGE_THIS_IN_PROD")
	userRepo := handler.NewInMemoryUserRepo()
	// Using MockOAuth to prevent nil pointer panics
	authHandler := handler.NewAuthHandler(&MockOAuth{}, jwtSvc, userRepo)

	// Initialize zap logger for ML service
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	// Proxy servers
	tcpProxy := server.NewTCPProxy()
	udpProxy := server.NewUDPProxy()
	httpProxy := server.NewHTTPProxy(tm, eventStream, geoLocator, rateLimiter, inspector, zapLogger)

	go func() {
		log.Println(" HTTP Proxy başlatılıyor...")
		if err := httpProxy.Start(); err != nil {
			log.Fatalf(" HTTP Proxy hatası: %v", err)
		}
	}()

	// Monitoring server (using auth, shared limiter, and inspector)
	monitor := server.NewMonitoringServer(tm, analyticsEngine, authHandler, rateLimiter, inspector, jwtSvc)
	go func() {
		if err := monitor.Start(); err != nil {
			log.Fatalf(" Monitoring server hatası: %v", err)
		}
	}()

	// Control Port'u dinle (Client'lar buraya bağlanacak)
	listener, err := net.Listen("tcp", protocol.ControlPort)
	if err != nil {
		log.Fatalf(" Port dinlenemedi: %v", err)
	}
	defer listener.Close()

	cizgi := strings.Repeat("=", 60)
	log.Printf("Control port dinleniyor: %s", protocol.ControlPort)
	log.Printf(" HTTP Proxy dinleniyor: %s", protocol.ProxyPort)
	log.Printf("Monitoring endpoint: http://localhost:9090/metrics")
	log.Printf("Authentication: ENABLED")
	log.Printf("Rate Limiting: ADVANCED (Sliding Window)")
	log.Println(" Client bağlantıları bekleniyor...")
	log.Println(cizgi)

	// Client handling
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf(" Bağlantı hatası: %v", err)
			continue
		}

		log.Printf("Yeni bağlantı: %s", conn.RemoteAddr())
		go handleClient(conn, tm, authManager, tcpProxy, udpProxy)
	}
}

func handleClient(conn net.Conn, tm *server.TunnelManager, authManager *server.AuthManager, tcpProx *server.TCPProxy, udpProx *server.UDPProxy) {
	defer conn.Close()

	// 1. REGISTER mesajını oku
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		log.Printf("Mesaj okunamadı: %v", err)
		return
	}

	if msg.Type != protocol.MsgTypeRegister {
		log.Printf(" Beklenmeyen mesaj tipi: %s", msg.Type)
		errMsg := protocol.NewErrorMessage(400, "İlk mesaj REGISTER olmalı")
		protocol.WriteMessage(conn, errMsg)
		return
	}

	log.Printf("📥 REGISTER mesajı alındı")

	// Parse REGISTER request
	var regReq protocol.RegisterRequest
	if err := json.Unmarshal([]byte(msg.Payload), &regReq); err != nil {
		log.Printf(" REGISTER parse edilemedi: %v", err)
		errMsg := protocol.NewErrorMessage(400, "Invalid REGISTER payload")
		protocol.WriteMessage(conn, errMsg)
		return
	}

	// AUTH: API key kontrolü
	authKey, err := authManager.ValidateKey(regReq.APIKey)
	if err != nil {
		log.Printf(" Authentication başarısız: %v", err)
		errMsg := protocol.NewErrorMessage(401, fmt.Sprintf("Authentication failed: %v", err))
		protocol.WriteMessage(conn, errMsg)
		return
	}

	log.Printf(" Authentication başarılı: %s (User: %s)", regReq.APIKey[:8]+"...", authKey.UserID)
	authManager.IncrementUsage(regReq.APIKey)

	// 4. Yamux Session başlat
	yamuxConfig := yamux.DefaultConfig()
	yamuxConfig.LogOutput = io.Discard

	session, err := yamux.Server(conn, yamuxConfig)
	if err != nil {
		log.Printf(" Yamux session başlatılamadı: %v", err)
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
			log.Printf(" Port tahsis hatası: %v", err)
			protocol.WriteMessage(conn, protocol.NewErrorMessage(500, "Boş port bulunamadı"))
			return
		}

		if regReq.TunnelType == "tcp" {
			if err := tcpProx.ListenAndForward(publicPort, session); err != nil {
				log.Printf(" TCP Proxy hatası: %v", err)
				protocol.WriteMessage(conn, protocol.NewErrorMessage(500, "TCP Proxy başlatılamadı"))
				return
			}
		} else {
			if err := udpProx.ListenAndForward(publicPort, session); err != nil {
				log.Printf(" UDP Proxy hatası: %v", err)
				protocol.WriteMessage(conn, protocol.NewErrorMessage(500, "UDP Proxy başlatılamadı"))
				return
			}
		}
		log.Printf("✅ %s tüneli oluşturuldu: :%d", strings.ToUpper(regReq.TunnelType), publicPort)
	} else {
		// DEFAULT: HTTP Subdomain
		subdomain = utils.GenerateSubDomain(8)
		fullURL = fmt.Sprintf("http://%s.%s%s", subdomain, protocol.BaseDomain, protocol.ProxyPort)
		log.Printf("✅ HTTP Subdomain atandı: %s", fullURL)
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

	log.Printf(" Yamux session başlatıldı: %s", subdomain)

	// 5. Session'ı kaydet
	if subdomain != "" {
		tm.RegisterTunnel(subdomain, session, regReq.CustomDomain, regReq.LocalPort, fullURL)
		defer tm.RemoveTunnel(subdomain)
	} else {
		// TCP/UDP için de kaydetmek gerekebilir (opsiyonel, monitoring için iyi olur)
	}

	cizgi := strings.Repeat("=", 60)
	log.Printf(" Aktif tünel sayısı: %d", tm.Count())
	log.Println(cizgi)

	// 6. Session kapanana kadar bekle
	for {
		if session.IsClosed() {
			<-session.CloseChan()
			log.Printf("🔌 Session kapandı: %s", subdomain)
			return
		}
		_, err := session.AcceptStream()
		if err != nil {
			log.Printf("🔌 Client bağlantısı kesildi: %s", subdomain)
			return
		}
	}
}
