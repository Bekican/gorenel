package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/Bekican/gorenel/internal/server"
	"github.com/Bekican/gorenel/internal/utils"
	"github.com/hashicorp/yamux"
)

func main() {
	log.Println("Gorenel server başlatılıyor.")

	tm := server.NewTunnelManager()

	proxy := server.NewHTTPProxy(tm)
	go func() {
		log.Println("HTTP proxy başlatılıyor.")
		if err := proxy.Start(); err != nil {
			log.Fatalf("HTTP proxy hatası : %v", err)
		}
	}()

	listener, err := net.Listen("tcp", protocol.ControlPort)
	if err != nil {
		log.Fatalf("Port dinlenemdi : %v", err)
	}
	defer listener.Close()

	log.Printf("Control port dinleniyor : %s", protocol.ControlPort)
	log.Printf("HTTP Proxy dinleniyor : %s", protocol.ProxyPort)
	log.Println("Client bağlantıları bekleniyor.")
	cizgi := strings.Repeat("=", 60)
	log.Println(cizgi)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Bağlantı hatası : %v", err)
			continue
		}

		log.Printf("Yeni bağlantı : %s", conn.RemoteAddr())
		go handleClient(conn, tm)
	}
}

func handleClient(conn net.Conn, tm *server.TunnelManager) {
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		log.Printf("Mesa okunamadı : %v", err)
		return
	}
	if msg.Type != protocol.MsgTypeRegister {
		log.Printf("Beklenmeyen mesj tipi %s", msg.Type)
		errMsg := protocol.NewErrorMessage(400, "İlk mesaj REGISTER olmalı.")

		protocol.WriteMessage(conn, errMsg)
		return
	}

	log.Printf("REGISTER mesajı alındı.")

	subdomain := utils.GenerateSubDomain(8)
	fullURL := fmt.Sprintf("http://%s.%s%s", subdomain, protocol.BaseDomain, protocol.ProxyPort)

	response := protocol.NewRegisteredMessage(subdomain, fullURL)
	if err := protocol.WriteMessage(conn, response); err != nil {
		log.Printf("Cevap gönderilemedi : %v", err)
		return
	}
	log.Printf("Subdomain atandı : %s", fullURL)
	yamuxConfig := yamux.DefaultConfig()
	yamuxConfig.LogOutput = io.Discard

	session, err := yamux.Server(conn, yamuxConfig)
	if err != nil {
		log.Printf("Yamux session başlatılamadı: %v", err)
		return
	}
	defer session.Close()

	log.Printf("Yamux session başlatıldı: %s", subdomain)

	tm.RegisterTunnel(subdomain, session)
	defer tm.RemoveTunnel(subdomain)

	log.Printf("Aktif tünel sayısı: %d", tm.Count())
	cizgi2 := strings.Repeat("=", 60)
	log.Println(cizgi2)

	for {
		if session.IsClosed() {
			log.Printf("Session kapandı: %s", subdomain)
			return
		}

		_, err := session.AcceptStream()
		if err != nil {
			log.Printf("Client bağlantısı kesildi: %s", subdomain)
			return
		}

	}
}
