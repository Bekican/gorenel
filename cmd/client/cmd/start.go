package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/Bekican/gorenel/internal/utils"
)

var (
	// Komut flag'leri
	serverAddr      string
	localPort       int
	customSubdomain string
	apiKey          string

	// Metrikler (atomic - thread-safe)
	requestCount  int64
	bytesReceived int64
	bytesSent     int64
)

// startCmd - Tunnel başlatma komutu
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Tunnel'ı başlat ve localhost'u internete aç",
	Long: `Belirtilen local port'u internete açar ve size public bir URL verir.

Örnekler:
  gorenel start --port 3000
  gorenel start --port 8080 --server tunnel.example.com:7000
  gorenel start --port 3000 --subdomain my-cool-app`,
	Run: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Flag'leri tanımla
	startCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server adresi (default: localhost:7000)")
	startCmd.Flags().IntVarP(&localPort, "port", "p", 3000, "Local port numarası")
	startCmd.Flags().StringVar(&customSubdomain, "subdomain", "", "Özel subdomain (mevcut değilse)")
	startCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API key (authentication için)")

	// Viper ile config dosyasından değerleri bağla
	viper.BindPFlag("server", startCmd.Flags().Lookup("server"))
	viper.BindPFlag("port", startCmd.Flags().Lookup("port"))
	viper.BindPFlag("api_key", startCmd.Flags().Lookup("api-key"))
}

func runStart(cmd *cobra.Command, args []string) {
	// Config'den veya flag'den değerleri al
	if serverAddr == "" {
		serverAddr = viper.GetString("server")
		if serverAddr == "" {
			serverAddr = "localhost" + protocol.ControlPort
		}
	}

	if localPort == 0 {
		localPort = viper.GetInt("port")
		if localPort == 0 {
			localPort = 3000
		}
	}

	if apiKey == "" {
		apiKey = viper.GetString("api_key")
		if apiKey == "" {
			log.Fatal("API key gerekli. --api-key flag'i veya config dosyasında 'api_key' belirtin.")
		}
	}

	// Banner
	printBanner()

	log.Printf("Gorenel Client v%s", rootCmd.Version)
	log.Printf("Local port: localhost:%d", localPort)
	log.Printf("Server: %s", serverAddr)

	if Verbose {
		log.Println("🔍 Verbose mod aktif")
	}

	// Context ile graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// SIGTERM/SIGINT yakalama
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Tunnel'ı başlat (goroutine içinde)
	errChan := make(chan error, 1)
	go func() {
		errChan <- startTunnel(ctx, serverAddr, localPort)
	}()

	// Metrikleri göster (her 10 saniyede bir)
	if Verbose {
		go printMetrics(ctx)
	}

	// Bekleme
	select {
	case <-sigChan:
		log.Println("\nGraceful shutdown başlatılıyor...")
		cancel()
		time.Sleep(2 * time.Second) // Cleanup için zaman tanı
	case err := <-errChan:
		if err != nil {
			log.Fatalf("Hata: %v", err)
		}
	}

	log.Println("Tunnel kapatıldı")
	printFinalMetrics()
}

// startTunnel - Ana tunnel mantığı
func startTunnel(ctx context.Context, serverAddr string, localPort int) error {
	// 1. Server'a bağlan
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("server'a bağlanılamadı: %w", err)
	}
	defer conn.Close()

	log.Println("Server'a bağlanıldı")

	// 2. REGISTER mesajı gönder
	clientID := utils.GenerateClientID()
	registerMsg := protocol.NewRegisterMessage(clientID, rootCmd.Version, apiKey)

	if err := protocol.WriteMessage(conn, registerMsg); err != nil {
		return fmt.Errorf("REGISTER gönderilemedi: %w", err)
	}

	if Verbose {
		log.Println("REGISTER mesajı gönderildi")
	}

	// 3. REGISTERED cevabını bekle
	response, err := protocol.ReadMessage(conn)
	if err != nil {
		return fmt.Errorf("cevap alınamadı: %w", err)
	}

	if response.Type == protocol.MsgTypeError {
		var errResp protocol.ErrorResponse
		json.Unmarshal([]byte(response.Payload), &errResp)
		return fmt.Errorf("server hatası: %s", errResp.Message)
	}

	var regResp protocol.RegisterResponse
	if err := json.Unmarshal([]byte(response.Payload), &regResp); err != nil {
		return fmt.Errorf("cevap parse edilemedi: %w", err)
	}

	// 4. Yamux session başlat
	yamuxConfig := yamux.DefaultConfig()
	// yamuxConfig.LogOutput = nil

	session, err := yamux.Client(conn, yamuxConfig)
	if err != nil {
		return fmt.Errorf("yamux başlatılamadı: %w", err)
	}
	defer session.Close()

	if Verbose {
		log.Println("Yamux session başlatıldı")
	}

	// 5. Başarı mesajı
	printSuccessBanner(regResp.FullURL, localPort)

	// 6. Stream'leri handle et
	go handleStreams(ctx, session, localPort)

	// Context iptal edilene kadar bekle
	<-ctx.Done()
	return nil
}

// handleStreams - Gelen stream'leri localhost'a yönlendir
func handleStreams(ctx context.Context, session *yamux.Session, localPort int) {
	localAddr := fmt.Sprintf("localhost:%d", localPort)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		stream, err := session.AcceptStream()
		if err != nil {
			if Verbose {
				log.Printf("Stream kabul edilemedi: %v", err)
			}
			return
		}

		atomic.AddInt64(&requestCount, 1)

		if Verbose {
			log.Printf("Yeni istek (Stream ID: %d)", stream.StreamID())
		}

		go proxyToLocalhost(stream, localAddr)
	}
}

// proxyToLocalhost - Stream'i localhost'a bağla
func proxyToLocalhost(stream net.Conn, localAddr string) {
	defer stream.Close()

	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Printf("Localhost'a bağlanılamadı: %v", err)
		return
	}
	defer localConn.Close()

	// Bidirectional copy with metrics
	done := make(chan struct{}, 2)

	// Stream → Localhost
	go func() {
		n, _ := io.Copy(localConn, stream)
		atomic.AddInt64(&bytesReceived, n)
		done <- struct{}{}
	}()

	// Localhost → Stream
	go func() {
		n, _ := io.Copy(stream, localConn)
		atomic.AddInt64(&bytesSent, n)
		done <- struct{}{}
	}()

	<-done
}

// --- UTILITY FONKSİYONLARI ---

func printBanner() {
	banner := `
   ____                            _
  / ___| ___  _ __ ___ _ __   ___| |
 | |  _ / _ \| '__/ _ \ '_ \ / _ \ |
 | |_| | (_) | | |  __/ | | |  __/ |
  \____|\___/|_|  \___|_| |_|\___|_|
                                     
  Modern Tunnel System - Built with Go
`
	fmt.Println(banner)
}

func printSuccessBanner(url string, port int) {
	cizgi := strings.Repeat("=", 60)
	fmt.Println("\n" + cizgi)
	fmt.Println("TÜNELİNİZ HAZIR!")
	fmt.Println(cizgi)
	fmt.Printf("Public URL:  %s\n", url)
	fmt.Printf("Local Port:  localhost:%d\n", port)
	fmt.Println(cizgi)
	fmt.Println("Tunnel çalışıyor. Kapatmak için Ctrl+C basın..")
}

func printMetrics(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count := atomic.LoadInt64(&requestCount)
			received := atomic.LoadInt64(&bytesReceived)
			sent := atomic.LoadInt64(&bytesSent)

			log.Printf(" Metrikler: %d istek | ⬇️  %s | ⬆️  %s",
				count,
				formatBytes(received),
				formatBytes(sent),
			)
		}
	}
}

func printFinalMetrics() {
	count := atomic.LoadInt64(&requestCount)
	received := atomic.LoadInt64(&bytesReceived)
	sent := atomic.LoadInt64(&bytesSent)

	fmt.Println("\n Son Metrikler:")
	fmt.Printf("   Toplam İstek:  %d\n", count)
	fmt.Printf("   Alınan Veri:   %s\n", formatBytes(received))
	fmt.Printf("   Gönderilen:    %s\n", formatBytes(sent))
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
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
