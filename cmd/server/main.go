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
	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/Bekican/gorenel/internal/server"
	"github.com/Bekican/gorenel/internal/utils"
	"github.com/Bekican/gorenel/pkg/auth"
	"github.com/hashicorp/yamux"
)

func main() {
	log.Println(" Gorenel Server başlatılıyor...")

	// Core components
	tm := server.NewTunnelManager()
	authManager := server.NewAuthManager()

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
	// HTTP Proxy'yi başlat (Port 8080)

	// Auth components
	jwtSvc := auth.NewJWTService("SUPER_SECRET_KEY_CHANGE_THIS_IN_PROD")
	userRepo := handler.NewInMemoryUserRepo()
	authHandler := handler.NewAuthHandler(jwtSvc, userRepo)

	proxy := server.NewHTTPProxy(tm, eventStream, geoLocator)
	go func() {
		log.Println(" HTTP Proxy başlatılıyor...")
		if err := proxy.Start(); err != nil {
			log.Fatalf(" HTTP Proxy hatası: %v", err)
		}
	}()

	// Monitoring server'ı başlat (Port 9090)
	monitor := server.NewMonitoringServer(tm, analyticsEngine, authHandler)
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
	log.Printf("Rate Limiting: ENABLED")
	log.Println(" Client bağlantıları bekleniyor...")
	log.Println(cizgi)

	// Her client için ayrı goroutine
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf(" Bağlantı hatası: %v", err)
			continue
		}

		log.Printf("Yeni bağlantı: %s", conn.RemoteAddr())
		go handleClient(conn, tm, authManager)
	}
}

func handleClient(conn net.Conn, tm *server.TunnelManager, authManager *server.AuthManager) {
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
	apiKey, err := authManager.ValidateKey(regReq.APIKey)
	if err != nil {
		log.Printf(" Authentication başarısız: %v", err)
		errMsg := protocol.NewErrorMessage(401, fmt.Sprintf("Authentication failed: %v", err))
		protocol.WriteMessage(conn, errMsg)
		return
	}

	log.Printf(" Authentication başarılı: %s (User: %s)", regReq.APIKey[:8]+"...", apiKey.UserID)
	authManager.IncrementUsage(regReq.APIKey)

	// 2. Subdomain üret
	subdomain := utils.GenerateSubDomain(8)
	fullURL := fmt.Sprintf("http://%s.%s%s", subdomain, protocol.BaseDomain, protocol.ProxyPort)

	// 3. Client'a subdomain'i bildir
	response := protocol.NewRegisteredMessage(subdomain, fullURL)
	if err := protocol.WriteMessage(conn, response); err != nil {
		log.Printf(" Cevap gönderilemedi: %v", err)
		return
	}

	log.Printf("✅ Subdomain atandı: %s", fullURL)

	// 4. Yamux Session başlat
	yamuxConfig := yamux.DefaultConfig()
	yamuxConfig.LogOutput = io.Discard

	session, err := yamux.Server(conn, yamuxConfig)
	if err != nil {
		log.Printf(" Yamux session başlatılamadı: %v", err)
		return
	}
	defer session.Close()

	log.Printf(" Yamux session başlatıldı: %s", subdomain)

	// 5. Session'ı kaydet
	tm.RegisterTunnel(subdomain, session)
	defer tm.RemoveTunnel(subdomain)

	cizgi := strings.Repeat("=", 60)
	log.Printf(" Aktif tünel sayısı: %d", tm.Count())
	log.Println(cizgi)

	// 6. Session kapanana kadar bekle
	// HTTP Proxy stream açtığında client tarafı yakalayacak
	for {
		if session.IsClosed() {
			<-session.CloseChan()
			log.Printf("🔌 Session kapandı: %s", subdomain)
			return
		}

		// Yamux kendi keepalive'ı yönetir
		// Burada sadece session'ın açık olup olmadığını kontrol ediyoruz
		_, err := session.AcceptStream()
		if err != nil {
			log.Printf("🔌 Client bağlantısı kesildi: %s", subdomain)
			return
		}
		// Not: Stream'leri HTTP Proxy açıyor, biz kapatıyoruz
	}
}
