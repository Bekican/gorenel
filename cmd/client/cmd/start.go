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
	serverAddr      string
	localPort       int
	customSubdomain string

	requestCount  int64
	bytesReceived int64
	bytesSent     int64
)

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
	//flag tanımlamaları
	startCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server adresi (default : localhost:7000)")
	startCmd.Flags().IntVarP(&localPort, "port", "p", 3000, "Local port numarası")
	startCmd.Flags().StringVar(&customSubdomain, "subdomain", "", "Özel subdomain(mevcut değilse)")

	viper.BindPFlag("server", startCmd.Flags().Lookup("server"))
	viper.BindPFlag("port", startCmd.Flags().Lookup("port"))
}

func runStart(cmd *cobra.Command, args []string) {
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

	printBanner()

	log.Printf("Gorenel Client v%s", rootCmd.Version)
	log.Printf("Local port : localhost : %d", localPort)
	log.Printf("Server : %s", serverAddr)

	if Verbose {
		log.Printf("Verbose moddasınız.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- startTunnel(ctx, serverAddr, localPort)
	}()

	if Verbose {
		go printMetrics(ctx)
	}

	select {
	case <-sigChan:
		log.Println("\n Shutdown başlatılıyor.")
		cancel()
		time.Sleep(2 * time.Second)
	case err := <-errChan:
		if err != nil {
			log.Fatalf("Hata : %v", err)
		}
	}

	log.Println("Tunnel kapatıldı.")
	printFinalMetrics()
}

func startTunnel(ctx context.Context, serverAddr string, localport int) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("server'a bağlanılamadı : %w", err)
	}
	// conn.Close'u burada defer etmiyoruz, çünkü bağlantı session içinde yaşamalı.
	// Ancak session kapanınca conn da kapanır.

	log.Println("Servera bağlanıldı")

	clientID := utils.GenerateClientID()

	registerMsg := protocol.NewRegisterMessage(clientID, rootCmd.Version)

	if err := protocol.WriteMessage(conn, registerMsg); err != nil {
		return fmt.Errorf("REGISTER gönderilemedi: %w", err)
	}

	if Verbose {
		log.Println("REGISTER mesajı gönderildi")
	}

	response, err := protocol.ReadMessage(conn)
	if err != nil {
		return fmt.Errorf("cevap alınamadı :%w", err)
	}
	if response.Type == protocol.MsgTypeError {
		var errResp protocol.ErrorResponse
		json.Unmarshal([]byte(response.Payload), &errResp)
		return fmt.Errorf("server hatası : %s", errResp.Message)
	}

	var regResp protocol.RegisterResponse
	if err := json.Unmarshal([]byte(response.Payload), &regResp); err != nil {
		return fmt.Errorf("cevap parse edilemedi : %w", err)
	}

	yamuxConfig := yamux.DefaultConfig()
	yamuxConfig.LogOutput = io.Discard
	yamuxConfig.EnableKeepAlive = false

	session, err := yamux.Client(conn, yamuxConfig)
	if err != nil {
		return fmt.Errorf("yamux başlatılamadı : %w", err)
	}
	defer session.Close()

	if Verbose {
		log.Println("Yamux başlatıldı.")
	}
	printSuccessBanner(regResp.FullURL, localport)

	go handleStreams(ctx, session, localport)

	<-ctx.Done()
	return nil
}

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

// proxyToLocalhost - Streami localhosta bağla
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
	fmt.Printf(" Local Port:  localhost:%d\n", port)
	fmt.Println(cizgi)
	fmt.Println("\n Tunnel çalışıyor. Kapatmak için Ctrl+C basın...")
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

			log.Printf("Metrikler: %d istek | ⬇️  %s | ⬆️  %s",
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
