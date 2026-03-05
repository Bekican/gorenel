package cmd

import (
	"bytes"
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
	customDomain    string
	tunnelType      string // --- NEW: "http", "tcp" or "udp" ---
	remotePort      int    // --- NEW: Requested remote port ---

	// Metrikler (atomic - thread-safe)
	requestCount  int64
	bytesReceived int64
	bytesSent     int64
)

// startCmd - Tunnel başlatma komutu
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Tunnel'ı başlat ve localhost'u internete aç",
	Long: `Belirtilen local port'u internete açar ve size public bir URL veya Port verir.

Örnekler:
  gorenel start --port 3000
  gorenel start --port 3000 --subdomain my-cool-app
  gorenel start --port 3306 --type tcp
  gorenel start --port 5000 --type udp`,
	Run: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Flag'leri tanımla
	startCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server adresi (default: localhost:7000)")
	startCmd.Flags().IntVarP(&localPort, "port", "p", 3000, "Local port numarası")
	startCmd.Flags().StringVar(&customSubdomain, "subdomain", "", "Özel subdomain (mevcut değilse)")
	startCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API key (authentication için)")
	startCmd.Flags().StringVarP(&customDomain, "domain", "d", "", "Özel alan adı (Custom Domain)")
	startCmd.Flags().StringVarP(&tunnelType, "type", "t", "http", "Tunnel tipi (http, tcp, udp)")
	startCmd.Flags().IntVarP(&remotePort, "remote-port", "r", 0, "İstenen uzak port (raw tüneller için)")

	// Viper ile config dosyasından değerleri bağla
	viper.BindPFlag("server", startCmd.Flags().Lookup("server"))
	viper.BindPFlag("port", startCmd.Flags().Lookup("port"))
	viper.BindPFlag("api_key", startCmd.Flags().Lookup("api-key"))
	viper.BindPFlag("domain", startCmd.Flags().Lookup("domain"))
	viper.BindPFlag("type", startCmd.Flags().Lookup("type"))
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

	// Flag boşsa viper'dan (config dosyasından) çek
	if customDomain == "" {
		customDomain = viper.GetString("domain")
	}

	if tunnelType == "http" {
		t := viper.GetString("type")
		if t != "" {
			tunnelType = t
		}
	}

	// Banner
	printBanner()

	log.Printf("Gorenel Client v%s", rootCmd.Version)
	log.Printf("Type: %s", strings.ToUpper(tunnelType))
	log.Printf("Local port: localhost:%d", localPort)
	log.Printf("Server: %s", serverAddr)
	if customDomain != "" {
		log.Printf("Custom Domain: %s", customDomain)
	}

	if Verbose {
		log.Println("🔍 Verbose mod aktif")
	}

	// Context ile graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// SIGTERM/SIGINT yakalama
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Tunnel persistence loop
	go func() {
		delay := 1 * time.Second
		maxDelay := 30 * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := startTunnel(ctx, serverAddr, localPort, customDomain, tunnelType)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("⚠️ Bağlantı sorunu: %v", err)
				log.Printf("🔄 %v saniye içinde tekrar bağlanılacak...", delay.Seconds())

				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
				}

				// Exponential backoff
				delay *= 2
				if delay > maxDelay {
					delay = maxDelay
				}
			} else {
				// Reset delay on success
				delay = 1 * time.Second
			}
		}
	}()

	// Show metrics (every 10 seconds)
	if Verbose {
		go printMetrics(ctx)
	}

	// Wait for signal
	<-sigChan
	log.Println("\nGraceful shutdown başlatılıyor...")
	cancel()
	time.Sleep(1 * time.Second)

	log.Println("Tunnel kapatıldı")
	printFinalMetrics()
}

// startTunnel - Ana tunnel mantığı
func startTunnel(ctx context.Context, serverAddr string, localPort int, domain string, tType string) error {
	// 1. Server'a bağlan
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("server'a bağlanılamadı: %w", err)
	}
	defer conn.Close()

	log.Println("Server'a bağlanıldı")

	// 2. REGISTER mesajı gönder
	clientID := utils.GenerateClientID()

	// Prepare register request with type
	regReq := protocol.RegisterRequest{
		ClientID:     clientID,
		Version:      rootCmd.Version,
		APIKey:       apiKey,
		CustomDomain: domain,
		TunnelType:   tType,
	}
	reqPayload, _ := json.Marshal(regReq)

	registerMsg := protocol.Message{
		Type:    protocol.MsgTypeRegister,
		Payload: string(reqPayload),
	}

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
	if tType == "tcp" || tType == "udp" {
		printRawSuccessBanner(tType, regResp.PublicPort, localPort)
	} else {
		url := regResp.FullURL
		if domain != "" {
			url = "http://" + domain + " (Custom Domain)"
		}
		printSuccessBanner(url, localPort)
	}

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

	startTime := time.Now()
	var method, path string

	// Sniff request if it's HTTP
	if tunnelType == "http" {
		buf := make([]byte, 1024)
		n, _ := stream.Read(buf)
		if n > 0 {
			line := buf[:n]
			parts := strings.Split(string(line), " ")
			if len(parts) >= 2 {
				method = parts[0]
				path = parts[1]
			}
			// Wrap stream to not lose read data
			stream = &MultiConn{
				Reader: io.MultiReader(bytes.NewReader(buf[:n]), stream),
				Conn:   stream,
			}
		}
	}

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
		// Attempt to sniff response status code if HTTP
		var statusCode int
		if tunnelType == "http" {
			buf := make([]byte, 1024)
			n, _ := localConn.Read(buf)
			if n > 0 {
				line := buf[:n]
				parts := strings.Split(string(line), " ")
				if len(parts) >= 2 && strings.HasPrefix(parts[1], "HTTP/") {
					fmt.Sscanf(parts[1], "%d", &statusCode)
				} else if len(parts) >= 2 && strings.HasPrefix(parts[0], "HTTP/") {
					fmt.Sscanf(parts[1], "%d", &statusCode)
				}

				// Wrap localConn to not lose read data
				localConn = &MultiConn{
					Reader: io.MultiReader(bytes.NewReader(buf[:n]), localConn),
					Conn:   localConn,
				}
			}
		}

		n, _ := io.Copy(stream, localConn)
		atomic.AddInt64(&bytesSent, n)

		if tunnelType == "http" && method != "" {
			dur := time.Since(startTime)
			statusStr := fmt.Sprintf("%d", statusCode)
			if statusCode == 0 {
				statusStr = "???"
			}

			// Colorize status
			color := "\033[32m" // Green
			if statusCode >= 400 {
				color = "\033[31m"
			} // Red
			if statusCode >= 300 && statusCode < 400 {
				color = "\033[33m"
			} // Yellow
			reset := "\033[0m"

			log.Printf("[%s] %s %s -> %s%s%s (%v)",
				time.Now().Format("15:04:05"),
				method, path, color, statusStr, reset,
				dur.Round(time.Millisecond))
		}

		done <- struct{}{}
	}()

	<-done
}

// MultiConn wraps an io.Reader and a net.Conn
type MultiConn struct {
	io.Reader
	net.Conn
}

func (c *MultiConn) Read(p []byte) (int, error) {
	return c.Reader.Read(p)
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

func printRawSuccessBanner(tType string, publicPort int, localPort int) {
	cizgi := strings.Repeat("=", 60)
	fmt.Println("\n" + cizgi)
	fmt.Printf("%s TÜNELİNİZ HAZIR!\n", strings.ToUpper(tType))
	fmt.Println(cizgi)
	fmt.Printf("Public Address: gorenel.net:%d\n", publicPort)
	fmt.Printf("Local Target:   localhost:%d\n", localPort)
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
