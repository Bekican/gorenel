package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/Bekican/gorenel/internal/utils"
	"github.com/hashicorp/yamux"
)

func main() {
	serverAddr := flag.String("server", "localhost"+protocol.ControlPort, "Server adresi")
	localPort := flag.Int("port", 3000, "Yerel port (localhost:PORT)")
	flag.Parse()

	log.Println("Gorenel client başlatılıyor.")
	log.Printf("Yerel port : localhost:%d", *localPort)
	log.Printf("Server: %s", *serverAddr)

	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		log.Fatalf("Server'a bağlanılamadı:%v", err)
	}
	defer conn.Close()

	log.Println("Server'a bağlanıldı.")

	clientID := utils.GenerateClientID()
	registerMsg := protocol.NewRegisterMessage(clientID, "1.0.0")

	if err := protocol.WriteMessage(conn, registerMsg); err != nil {
		log.Fatalf("REGISTER mesajı gönderilemedi: %v", err)
	}

	log.Println("REGISTER mesajı gönderildi.")

	response, err := protocol.ReadMessage(conn)
	if err != nil {
		log.Fatalf("Cevap alınamadı : %v", err)
	}

	if response.Type == protocol.MsgTypeError {
		var errResp protocol.ErrorResponse
		json.Unmarshal([]byte(response.Payload), &errResp)
		log.Fatalf("Server hatası: %s", errResp.Message)
	}

	if response.Type != protocol.MsgTypeRegistered {
		log.Fatalf("Beklenmeyen cevap : %s", response.Type)
	}

	var regResp protocol.RegisterResponse
	if err := json.Unmarshal([]byte(response.Payload), &regResp); err != nil {
		log.Fatalf("Cevap parse edilemedi : %v", err)
	}

	//yamux(Yet Another Multiplexer) -> çoklayıcı session başlatıldı
	yamuxConfig := yamux.DefaultConfig()
	yamuxConfig.LogOutput = io.Discard
	yamuxConfig.EnableKeepAlive = false

	session, err := yamux.Client(conn, yamuxConfig)
	if err != nil {
		log.Fatalf("Yamux session başlatılamadı : %v", err)
	}
	defer session.Close()

	log.Println("Yamux session başlatıldı")

	cizgi := strings.Repeat("=", 60)
	fmt.Println("\n" + cizgi)
	fmt.Println("TÜNELİNİZ HAZIR!")
	fmt.Println(cizgi)
	fmt.Printf(" Public URL:  %s\n", regResp.FullURL)
	fmt.Printf(" Local Port:  localhost:%d\n", *localPort)
	fmt.Println(cizgi)
	fmt.Println("\n Tünel açık. Kapatmak için Ctrl+C basın..")

	// streamlari dinle ve localhosta gönder
	go handleStreams(session, *localPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n\nTünel kapatılıyor.")
	log.Println("Temiz kapanış gerçekleşti")
}

// serverdan gelen streamlari localhosta gönderir
func handleStreams(session *yamux.Session, localPort int) {

	localAddr := fmt.Sprintf("localhost : %d", localPort)

	for {
		stream, err := session.AcceptStream()
		if err != nil {
			log.Printf("Stream kabul edilemedi : %v", err)
			return
		}
		log.Printf("Yeni stream alındı (ID : %d)", stream.StreamID())

		//her stream için ayrı goroutine
		go proxyToLocalHost(stream, localAddr)
	}
}

// Streamdan gelen veriyi localhosta gönderir
func proxyToLocalHost(stream *yamux.Stream, localAddr string) {
	defer stream.Close()

	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Printf("Localhost'a bağlanılamadı : %v", err)
		return
	}
	defer localConn.Close()

	log.Printf("Localhost'a bağlanıldı : %s (stream : %d)", localAddr, stream.StreamID())

	//iki yönlü veri aktarımı
	//Stream ve localhost arasında
	done := make(chan struct{}, 2)

	//stream -> localhost
	go func() {
		io.Copy(localConn, stream)
		done <- struct{}{}
	}()

	//localhost -> stream
	go func() {
		io.Copy(stream, localConn)
		done <- struct{}{}
	}()

	<-done
	log.Printf("Stream kapandı (ID : %d)", stream.StreamID())
}
