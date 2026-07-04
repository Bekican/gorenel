package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
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
	keyAuthToken    string
	ipWhitelist     []string
	corsEnabled     bool
	preferRegion    string
	stableSubdomain bool
	projectName     string
	authCredentials string
	publicAccess    bool

	// Metrikler (atomic - thread-safe)
	requestCount  int64
	bytesReceived int64
	bytesSent     int64

	// Centralized buffer pools for zero-allocation stream copy
	headerBufPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 64*1024) // 64KB
			return &b
		},
	}
	copyBufPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 32*1024) // 32KB
			return &b
		},
	}
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
	startCmd.Flags().StringVar(&keyAuthToken, "key-auth", "", "Tunnel key auth token (public requests must send header X-TOKEN)")
	startCmd.Flags().StringArrayVar(&ipWhitelist, "ip-whitelist", []string{}, "Allowed client IP/CIDR (repeatable). Example: --ip-whitelist 1.2.3.4 --ip-whitelist 10.0.0.0/24")
	startCmd.Flags().BoolVar(&corsEnabled, "cors", false, "Enable built-in Smart CORS handling at the proxy level")
	startCmd.Flags().StringVar(&preferRegion, "region", "", "Prefer a Fly.io region for tunnel control-plane (sets Fly-Prefer-Region header, e.g. fra, ams, iad)")
	startCmd.Flags().BoolVar(&stableSubdomain, "stable", false, "Sabit subdomain kullan (proje+port'a göre). İlk çalıştırmada rezervasyon otomatik oluşturulur.")
	startCmd.Flags().StringVar(&projectName, "project", "", "Stable subdomain için proje adı (default: current folder name)")
	startCmd.Flags().StringVar(&authCredentials, "auth", "", "Password protect your tunnel (form: 'user:pass')")
	startCmd.Flags().BoolVar(&publicAccess, "public", true, "Make the tunnel publicly accessible (no X-TOKEN auth required)")

	// Viper ile config dosyasından değerleri bağla
	viper.BindPFlag("server", startCmd.Flags().Lookup("server"))
	viper.BindPFlag("port", startCmd.Flags().Lookup("port"))
	viper.BindPFlag("api_key", startCmd.Flags().Lookup("api-key"))
	viper.BindPFlag("domain", startCmd.Flags().Lookup("domain"))
	viper.BindPFlag("type", startCmd.Flags().Lookup("type"))
	viper.BindPFlag("region", startCmd.Flags().Lookup("region"))
	viper.BindPFlag("stable", startCmd.Flags().Lookup("stable"))
	viper.BindPFlag("project", startCmd.Flags().Lookup("project"))
	viper.BindPFlag("public", startCmd.Flags().Lookup("public"))
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

	if !cmd.Flags().Changed("region") {
		preferRegion = viper.GetString("region")
	}
	if !cmd.Flags().Changed("stable") {
		stableSubdomain = viper.GetBool("stable")
	}
	if !cmd.Flags().Changed("project") {
		projectName = viper.GetString("project")
	}
	if !cmd.Flags().Changed("auth") {
		authCredentials = viper.GetString("auth")
	}
	if !cmd.Flags().Changed("public") {
		if viper.IsSet("public") {
			publicAccess = viper.GetBool("public")
		} else {
			publicAccess = true
		}
	}

	serverAddr = strings.TrimSpace(serverAddr)
	apiKey = strings.TrimSpace(apiKey)
	tunnelType = strings.ToLower(strings.TrimSpace(tunnelType))
	preferRegion = strings.ToLower(strings.TrimSpace(preferRegion))
	if tunnelType != "http" && tunnelType != "tcp" && tunnelType != "udp" && tunnelType != "mcp" {
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

	// Stable subdomain: if user didn't specify a subdomain, derive one from project+port.
	if stableSubdomain && strings.TrimSpace(customSubdomain) == "" && (tunnelType == "http" || tunnelType == "mcp") {
		p := strings.TrimSpace(projectName)
		if p == "" {
			if wd, err := os.Getwd(); err == nil {
				p = filepath.Base(wd)
			}
		}
		customSubdomain = makeStableSubdomain(p, localPort)
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

			err := startTunnel(ctx, serverAddr, localPort, customDomain, tunnelType, stableSubdomain)
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
func startTunnel(ctx context.Context, serverAddr string, localPort int, domain string, tType string, ensureReservation bool) error {
	// Optional: ensure the requested subdomain is reserved & bound to this API key.
	if ensureReservation && strings.TrimSpace(customSubdomain) != "" && (tType == "http" || tType == "mcp") {
		if err := ensureReservedSubdomain(ctx, serverAddr, apiKey, strings.TrimSpace(customSubdomain)); err != nil {
			return err
		}
	}

	// 1. Server'a bağlan (WebSocket üzerinden)
	var conn net.Conn
	var err error

	if strings.HasPrefix(serverAddr, "ws://") || strings.HasPrefix(serverAddr, "wss://") {
		// WebSocket bağlantısı (Fly.io shared IP için)
		header := http.Header{}
		// Provide API key at handshake time so the server can reject unauthenticated WS early.
		header.Set("X-API-Key", strings.TrimSpace(apiKey))
		if strings.TrimSpace(preferRegion) != "" {
			header.Set("Fly-Prefer-Region", strings.TrimSpace(preferRegion))
		}
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
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			_ = tcpConn.SetNoDelay(true)
		}
	}
	defer conn.Close()

	log.Println("Server'a bağlanıldı")

	// 2. REGISTER mesajı gönder
	clientID := utils.GenerateClientID()

	// Prepare register request with type
	regReq := protocol.RegisterRequest{
		ClientID:        clientID,
		Version:         rootCmd.Version,
		APIKey:          apiKey,
		CustomSubdomain: strings.TrimSpace(customSubdomain),
		CustomDomain:    domain,
		TunnelType:      tType,
		LocalPort:       localPort,
		KeyAuthToken:    strings.TrimSpace(keyAuthToken),
		IPWhitelist:     ipWhitelist,
		CORSEnabled:     corsEnabled,
		Auth:            strings.TrimSpace(authCredentials),
		PublicAccess:    &publicAccess,
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
	yamuxConfig.MaxStreamWindowSize = 1024 * 1024

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
		printSuccessBanner(url, localPort, regResp.AccessToken)
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

func makeStableSubdomain(project string, port int) string {
	p := slugify(project)
	if p == "" {
		p = "app"
	}
	// Keep it short-ish: most DNS label limits are 63 chars. Reserve some room for "-<port>".
	maxProjectLen := 63 - (1 + len(fmt.Sprintf("%d", port)))
	if maxProjectLen < 3 {
		maxProjectLen = 3
	}
	if len(p) > maxProjectLen {
		p = p[:maxProjectLen]
		p = strings.Trim(p, "-")
		if p == "" {
			p = "app"
		}
	}
	return fmt.Sprintf("%s-%d", p, port)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	b.Grow(len(s))
	prevDash := false
	for _, r := range s {
		isAZ := r >= 'a' && r <= 'z'
		is09 := r >= '0' && r <= '9'
		if isAZ || is09 {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if r == '-' || r == '_' || r == ' ' || r == '.' || r == '/' || r == '\\' {
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
			continue
		}
		// drop other characters
	}
	out := strings.Trim(b.String(), "-")
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return out
}

func ensureReservedSubdomain(ctx context.Context, serverAddr, apiKey, subdomain string) error {
	base, err := reservationsBaseURLFromServerAddr(serverAddr)
	if err != nil {
		return err
	}
	u := base + "/api/reservations/ensure"

	body, _ := json.Marshal(map[string]string{"subdomain": subdomain})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", strings.TrimSpace(apiKey))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("reservation ensure request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(b))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("reservation ensure failed (%d): %s", resp.StatusCode, msg)
	}
	return nil
}

func reservationsBaseURLFromServerAddr(serverAddr string) (string, error) {
	serverAddr = strings.TrimSpace(serverAddr)
	if serverAddr == "" {
		return "", fmt.Errorf("server address is empty")
	}

	if strings.HasPrefix(serverAddr, "ws://") || strings.HasPrefix(serverAddr, "wss://") {
		pu, err := url.Parse(serverAddr)
		if err != nil {
			return "", fmt.Errorf("invalid server url: %w", err)
		}
		switch pu.Scheme {
		case "wss":
			pu.Scheme = "https"
		case "ws":
			pu.Scheme = "http"
		}
		pu.Path = ""
		pu.RawQuery = ""
		pu.Fragment = ""
		return strings.TrimRight(pu.String(), "/"), nil
	}

	// Legacy raw TCP form: host:port
	host := serverAddr
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	if !strings.Contains(host, ":") {
		// best-effort: assume https
		return "https://" + host, nil
	}
	// If a port is provided, assume http (local/dev)
	return "http://" + host, nil
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
// Server tarafındaki httputil.ReverseProxy zaten Host header'ını doğru set ediyor.
// CLI tarafında veriyi DEĞİŞTİRMEDEN (unmodified) geçiriyoruz.
func proxyToLocalhost(stream net.Conn, localAddr string) {
	defer stream.Close()

	targetAddr := localAddr
	if strings.HasPrefix(targetAddr, "localhost:") {
		targetAddr = strings.Replace(targetAddr, "localhost:", "127.0.0.1:", 1)
	}

	startTime := time.Now()
	var method, path string
	isWebSocket := false

	// Server'ın httputil.ReverseProxy'si zaten Host: 127.0.0.1:<port> set ediyor.
	// CLI tarafında Host rewrite'a gerek yok — veriyi olduğu gibi geçiriyoruz.
	var requestReader io.Reader = stream
	if tunnelType == "http" || tunnelType == "mcp" {
		br := bufio.NewReader(stream)
		var headersBuf bytes.Buffer

		// --- ÖĞRETİCİ ADIM 1: Non-Destructive HTTP Header Ayrıştırma ---
		// Akıştan (stream) HTTP istek başlıklarını (\r\n\r\n ile sonlanan bölümü) satır satır okuyoruz.
		// Bu işlemi yaparken, okuduğumuz tüm verileri headersBuf içerisine kaydediyoruz (tamponluyoruz).
		// Böylece local sunucuya (localhost) aktarırken verilerin kaybolmamasını (non-destructive) sağlıyoruz.
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				if line != "" {
					headersBuf.WriteString(line)
				}
				break
			}
			headersBuf.WriteString(line)

			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				// Boş satır, HTTP başlıklarının sonunu işaret eder (\r\n\r\n).
				break
			}

			// --- ÖĞRETİCİ ADIM 2: WebSocket Upgrade Header Tespiti ---
			// HTTP protokolünde WebSockets, standart bir HTTP/1.1 isteğinin "Upgrade: websocket" başlığı ile
			// WebSocket protokolüne yükseltilmesi (Handshake) ile başlar.
			// Bu başlığı case-insensitive olarak kontrol ediyoruz.
			lowerLine := strings.ToLower(trimmed)
			if strings.HasPrefix(lowerLine, "upgrade:") && strings.Contains(lowerLine, "websocket") {
				isWebSocket = true
			}
		}

		// Okunan tüm başlıkları (headersBuf) ve soketin geri kalan kısmını (br) birleştiriyoruz.
		// io.MultiReader, önce ilk argümandaki tamponlanmış veriyi okur, o bitince stream'in kendisinden okumaya devam eder.
		requestReader = io.MultiReader(bytes.NewReader(headersBuf.Bytes()), br)

		// Metot ve path'i loglama için ilk satırdan çıkarıyoruz.
		headerStr := headersBuf.String()
		lines := strings.SplitN(headerStr, "\n", 2)
		if len(lines) > 0 {
			firstLine := lines[0]
			parts := strings.SplitN(strings.TrimSpace(firstLine), " ", 3)
			if len(parts) >= 2 {
				method = parts[0]
				path = parts[1]
			}
		}
	}

	localConn, err := net.DialTimeout("tcp", targetAddr, 5*time.Second)
	if err != nil {
		if Verbose {
			log.Printf("❌ Yerel servis hatası (%s): %v", targetAddr, err)
		}
		return
	}
	defer localConn.Close()
	if tcpConn, ok := localConn.(*net.TCPConn); ok {
		_ = tcpConn.SetNoDelay(true)
	}

	done := make(chan struct{}, 2)

	// Stream (Browser/Yamux) -> Localhost (TCP)
	go func() {
		bufPtr := copyBufPool.Get().(*[]byte)
		defer copyBufPool.Put(bufPtr)
		_, _ = io.CopyBuffer(localConn, requestReader, *bufPtr)
		
		// --- ÖĞRETİCİ ADIM 3: Koşullu TCP Yarım Kapatma (Half-Close) Kontrolü ---
		// Standart HTTP'de istek tamamlandıktan sonra soketin yazma ucunu kapatmak (TCP Half-Close)
		// yerel sunucuya "istek verisi bitti, şimdi cevabını yazabilirsin" sinyalini iletir.
		// Ancak WebSocket tam çift-yönlü (full-duplex) çalışan uzun ömürlü bir protokoldür.
		// Eğer tarayıcıdan yerel sunucuya giden yazma kanalını kapatırsak (CloseWrite),
		// tarayıcı daha sonra WebSocket üzerinden sunucuya veri gönderemez ve bağlantı derhal sonlanır (veya kilitlenir).
		if isWebSocket {
			if Verbose {
				log.Printf("\033[35m🔄 [Gorenel Client] [WebSocket Algılandı] Çift yönlü akışı (HMR/WS) korumak için TCP Half-Close devredışı bırakıldı. Bağlantı açık tutuluyor: %s\033[0m", path)
			}
		} else {
			// Standart HTTP istekleri için yarım kapatma (half-close) yaparak akışı sonlandırıyoruz.
			if tc, ok := localConn.(*net.TCPConn); ok {
				tc.CloseWrite()
			}
		}
		done <- struct{}{}
	}()

	// Localhost (TCP) -> Stream (Browser/Yamux)
	go func() {
		var statusCode int
		var responseReader io.Reader = localConn
		var headerBufToPut *[]byte

		defer func() {
			if headerBufToPut != nil {
				headerBufPool.Put(headerBufToPut)
			}
		}()

		if tunnelType == "http" || tunnelType == "mcp" {
			bufPtr := headerBufPool.Get().(*[]byte)
			buf := *bufPtr
			n, _ := localConn.Read(buf)
			if n > 0 {
				// Status code çıkar (logging amaçlı)
				chunk := string(buf[:n])
				if idx := strings.Index(chunk, "\r\n"); idx > 0 {
					statusParts := strings.SplitN(chunk[:idx], " ", 3)
					if len(statusParts) >= 2 {
						fmt.Sscanf(statusParts[1], "%d", &statusCode)
					}
				}

				// Veriyi DEĞİŞTİRMEDEN geri koy
				headerBufToPut = bufPtr
				responseReader = io.MultiReader(bytes.NewReader(buf[:n]), localConn)
			} else {
				headerBufPool.Put(bufPtr)
			}
		}

		copyBufPtr := copyBufPool.Get().(*[]byte)
		defer copyBufPool.Put(copyBufPtr)

		n, _ := io.CopyBuffer(stream, responseReader, *copyBufPtr)
		atomic.AddInt64(&bytesSent, n)

		// Final log output
		if (tunnelType == "http" || tunnelType == "mcp") && method != "" {
			dur := time.Since(startTime)
			statusStr := fmt.Sprintf("%d", statusCode)
			if statusCode == 0 {
				statusStr = "???"
			}

			color := "\033[32m" // Green
			icon := "✓"
			if statusCode >= 400 {
				color = "\033[31m"
				icon = "✗"
			}
			if statusCode >= 300 && statusCode < 400 {
				color = "\033[33m"
				icon = "→"
			}
			if statusCode == 101 {
				color = "\033[36m" // Cyan for WebSocket / protocol switch
				icon = "⚡"
			}
			reset := "\033[0m"

			// Eğer bu bir WebSocket ise veya 101 Switching Protocols döndüyse loga belirgin bir işaret ekle:
			wsSuffix := ""
			if isWebSocket || statusCode == 101 {
				wsSuffix = " ⚡ [WS-Upgrade]"
			}

			fmt.Printf("%s %s %-6s %-30s %s%s%s%s (%v)%s\n",
				color, icon, method, path, color, statusStr, wsSuffix, reset,
				dur.Round(time.Millisecond), reset)
		}
		done <- struct{}{}
	}()

	// Her iki yön de bitene kadar bekle — response kesilmesin.
	<-done
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

func printSuccessBanner(url string, port int, token string) {
	cizgi := strings.Repeat("―", 60)
	fmt.Println("\n" + cizgi)
	fmt.Println(" ✨  TÜNELİNİZ AKTİF!")
	fmt.Println(cizgi)
	fmt.Printf(" 🌍  Public URL:  \033[1;32m%s\033[0m\n", url)
	fmt.Printf(" 🏠  Local Port:  localhost:%d\n", port)
	if token != "" {
		fmt.Printf(" 🔑  Access Token: \033[1;33m%s\033[0m (Use X-TOKEN header for requests)\n", token)
	}
	fmt.Println(cizgi)
	fmt.Println(" ⚡  powered by gorenel | Kapatmak için: Ctrl+C")
	fmt.Println(cizgi + "\n")
}

func printRawSuccessBanner(tType string, publicPort int, localPort int) {
	cizgi := strings.Repeat("―", 60)
	fmt.Println("\n" + cizgi)
	fmt.Printf(" 🚀  %s TÜNELİNİZ AKTİF!\n", strings.ToUpper(tType))
	fmt.Println(cizgi)
	fmt.Printf(" 🌍  Public Address: \033[1;32mgorenel.site:%d\033[0m\n", publicPort)
	fmt.Printf(" 🏠  Local Target:   localhost:%d\n", localPort)
	fmt.Println(cizgi)
	fmt.Println(" ⚡  powered by gorenel | Kapatmak için: Ctrl+C")
	fmt.Println(cizgi + "\n")
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
