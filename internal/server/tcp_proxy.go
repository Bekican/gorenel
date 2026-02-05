package server

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/hashicorp/yamux"
)

// TCPProxy handles raw TCP traffic forwarding
type TCPProxy struct {
	// Any configuration if needed
}

func NewTCPProxy() *TCPProxy {
	return &TCPProxy{}
}

// Start listener for a specific port and link it to a yamux session
func (p *TCPProxy) ListenAndForward(publicPort int, session *yamux.Session) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", publicPort))
	if err != nil {
		return fmt.Errorf("TCP listener error on port %d: %w", publicPort, err)
	}

	log.Printf("TCP Proxy listening on :%d", publicPort)

	go func() {
		defer listener.Close()
		for {
			// Accept connection from public internet
			conn, err := listener.Accept()
			if err != nil {
				// If session is closed, accept will fail
				if session.IsClosed() {
					return
				}
				log.Printf("TCP Accept error: %v", err)
				continue
			}

			// Open a new Yamux stream to the client
			stream, err := session.OpenStream()
			if err != nil {
				log.Printf("Failed to open Yamux stream: %v", err)
				conn.Close()
				continue
			}

			// Bi-directional pipe
			go p.proxyConn(conn, stream)
		}
	}()

	return nil
}

func (p *TCPProxy) proxyConn(conn net.Conn, stream net.Conn) {
	defer conn.Close()
	defer stream.Close()

	done := make(chan struct{}, 2)

	// Local internet -> Yamux stream -> Client
	go func() {
		io.Copy(stream, conn)
		done <- struct{}{}
	}()

	// Client -> Yamux stream -> Local internet
	go func() {
		io.Copy(conn, stream)
		done <- struct{}{}
	}()

	<-done
}
