package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
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
	startCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server adresi (default: gorenel.site:7000)")
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
	// Flag > env > config > default
	if !cmd.Flags().Changed("server") {
		serverAddr = viper.GetString("server")
	}

	if !cmd.Flags().Changed("port") {
		if p := viper.GetInt("port"); p > 0 {
			localPort = p
		}
	}

	apiKeyProvided := cmd.Flags().Changed("api-key") || cmd.Flags().Changed("key")
	if !apiKeyProvided {
		apiKey = viper.GetString("api_key")
	}
	if apiKey == "" {
		log.Fatal("API key gerekli. `gorenel config init` veya `gorenel config set api_key <KEY>` ile kaydedin.")
	}

	if !cmd.Flags().Changed("domain") {
		customDomain = viper.GetString("domain")
	}

	if !cmd.Flags().Changed("type") {
		if t := viper.GetString("type"); t != "" {
			tunnelType = t
		}
	}

	serverAddr = strings.TrimSpace(serverAddr)
	apiKey = strings.TrimSpace(apiKey)
	tunnelType = strings.ToLower(strings.TrimSpace(tunnelType))
	if tunnelType != "http" && tunnelType != "tcp" && tunnelType != "udp" {
		log.Printf("Invalid tunnel type '%s', defaulting to http", tunnelType)
		tunnelType = "http"
	}
	if localPort <= 0 || localPort > 65535 {
		log.Printf("Invalid local port %d, defaulting to 3000", localPort)
		localPort = 3000
	}
	if serverAddr == "" {
		serverAddr = "wss://gorenel.site/tunnel/connect"
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
	// 1. Server'a bağlan (WebSocket üzerinden)
	var conn net.Conn
	var err error

	if strings.HasPrefix(serverAddr, "ws://") || strings.HasPrefix(serverAddr, "wss://") {
		// WebSocket bağlantısı (Fly.io shared IP için)
		header := http.Header{}
		dialer := websocket.Dialer{
			HandshakeTimeout: 15 * time.Second,
		}
		ws, _, err := dialer.DialContext(ctx, serverAddr, header)
		if err != nil {
			return fmt.Errorf("WebSocket bağlantısı kurulamadı: %w", err)
		}
		conn = NewClientWSConn(ws)
	} else {
		// Legacy: Raw TCP bağlantısı
		conn, err = net.Dial("tcp", serverAddr)
		if err != nil {
			return fmt.Errorf("server'a bağlanılamadı: %w", err)
		}
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
		LocalPort:    localPort,
	}
	reqPayload, err := json.Marshal(regReq)
	if err != nil {
		return fmt.Errorf("REGISTER payload hazırlanamadı: %w", err)
	}

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
		if err := json.Unmarshal([]byte(response.Payload), &errResp); err != nil {
			return fmt.Errorf("server hata mesajı parse edilemedi: %w", err)
		}
		msg := strings.ToLower(errResp.Message)
		if strings.Contains(msg, "authentication failed") || strings.Contains(msg, "invalid api key") {
			return fmt.Errorf("kimlik dogrulama basarisiz: API key gecersiz veya iptal edilmis.\nCozum: gorenel config set api_key <YENI_KEY>\nKontrol: gorenel config get")
		}
		return fmt.Errorf("server hatasi: %s", errResp.Message)
	}

	var regResp protocol.RegisterResponse
	if err := json.Unmarshal([]byte(response.Payload), &regResp); err != nil {
		return fmt.Errorf("cevap parse edilemedi: %w", err)
	}

	// 4. Yamux session başlat
	yamuxConfig := yamux.DefaultConfig()
	yamuxConfig.EnableKeepAlive = true
	yamuxConfig.KeepAliveInterval = 30 * time.Second

	session, err := yamux.Client(conn, yamuxConfig)
	if err != nil {
		return fmt.Errorf("yamux başlatılamadı: %w", err)
	}
	defer session.Close()

	if Verbose {
		log.Println("Yamux session başlatıldı")
	}

	// Periyodik ping ile tünel sağlığı (verbose: RTT loglanır)
	go tunnelHealthLoop(ctx, session)

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
	streamErr := make(chan error, 1)
	go handleStreams(ctx, session, localPort, streamErr)

	// Context iptal edilene veya session kopana kadar bekle
	select {
	case <-ctx.Done():
		return nil
	case err := <-streamErr:
		if err != nil {
			return err
		}
		return fmt.Errorf("tunnel stream kapandı")
	}
}

// handleStreams - Gelen stream'leri localhost'a yönlendir
func handleStreams(ctx context.Context, session *yamux.Session, localPort int, streamErr chan<- error) {
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
			select {
			case streamErr <- err:
			default:
			}
			return
		}

		atomic.AddInt64(&requestCount, 1)

		if Verbose {
			log.Printf("Yeni istek (Stream ID: %d)", stream.StreamID())
		}

		if tunnelType == "udp" {
			go proxyUDPStream(stream, localPort)
		} else {
			go proxyToLocalhost(stream, localAddr)
		}
	}
}

// tunnelHealthLoop runs application-level yamux pings (in addition to yamux keepalive frames).
func tunnelHealthLoop(ctx context.Context, session *yamux.Session) {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rtt, err := session.Ping()
			if err != nil {
				if Verbose {
					log.Printf("⚠️ Tunnel ping başarısız: %v", err)
				}
				return
			}
			if Verbose {
				log.Printf("🔌 Tunnel RTT: %s", rtt.Round(time.Millisecond))
			}
		}
	}
}

// proxyUDPStream relays length-framed UDP datagrams between yamux and local UDP.
func proxyUDPStream(stream net.Conn, localPort int) {
	raddr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: localPort}
	udpConn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Printf("❌ Yerel UDP bağlantısı kurulamadı (127.0.0.1:%d): %v", localPort, err)
		return
	}
	defer udpConn.Close()
	defer stream.Close()

	done := make(chan struct{}, 2)

	go func() {
		for {
			payload, err := protocol.ReadUDPFrame(stream)
			if err != nil {
				break
			}
			if _, err := udpConn.Write(payload); err != nil {
				break
			}
			atomic.AddInt64(&bytesReceived, int64(len(payload)))
		}
		_ = udpConn.Close()
		done <- struct{}{}
	}()

	go func() {
		buf := make([]byte, 65507)
		for {
			n, err := udpConn.Read(buf)
			if err != nil {
				break
			}
			if err := protocol.WriteUDPFrame(stream, buf[:n]); err != nil {
				break
			}
			atomic.AddInt64(&bytesSent, int64(n))
		}
		done <- struct{}{}
	}()

	<-done
	<-done
}

// proxyToLocalhost - Stream'i localhost'a bağla
func proxyToLocalhost(stream net.Conn, localAddr string) {
	defer stream.Close()

	// 'localhost' yerine '127.0.0.1' kullanmak IPv6/IPv4 karmaşasını önler
	targetAddr := localAddr
	if strings.HasPrefix(targetAddr, "localhost:") {
		targetAddr = strings.Replace(targetAddr, "localhost:", "127.0.0.1:", 1)
	}

	startTime := time.Now()
	var method, path string

	// Sniff request if it's HTTP
	if tunnelType == "http" {
		buf := make([]byte, 1024)
		n, _ := stream.Read(buf)
		if n > 0 {
			line := buf[:n]
			lineStr := string(line)
			parts := strings.Split(lineStr, " ")
			if len(parts) >= 2 {
				method = parts[0]
				path = parts[1]
			}

			// --- ÖNEMLİ: Host header'ını localhost'a çevir ---
			// Next.js vb. frameworkler Host header'ı uyuşmazsa redirect atabilir veya hata verebilir.
			modifiedBuf := line
			hostIdx := strings.Index(lineStr, "Host: ")
			if hostIdx != -1 {
				endIdx := strings.Index(lineStr[hostIdx:], "\r\n")
				if endIdx != -1 {
					oldHostLine := lineStr[hostIdx : hostIdx+endIdx]
					newHostLine := "Host: " + localAddr
					lineStr = strings.Replace(lineStr, oldHostLine, newHostLine, 1)
					modifiedBuf = []byte(lineStr)
				}
			}

			// Wrap stream to not lose read data
			stream = &MultiConn{
				Reader: io.MultiReader(bytes.NewReader(modifiedBuf), stream),
				Conn:   stream,
			}
		}
	}

	log.Printf("Yerel servise bağlanılıyor: %s...", targetAddr)
	localConn, err := net.DialTimeout("tcp", targetAddr, 5*time.Second)
	if err != nil {
		log.Printf("❌ Yerel servise BAĞLANILAMADI (%s): %v. Lütfen projenizin bu adreste çalıştığından emin olun.", targetAddr, err)
		return
	}
	defer localConn.Close()
	log.Printf("✅ Yerel servise BAĞLANILDI: %s", targetAddr)

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
	fmt.Printf("Public Address: gorenel.site:%d\n", publicPort)
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
