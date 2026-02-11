package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/hashicorp/yamux"
	"go.uber.org/zap"
)

// UDPProxy handles UDP packet forwarding using a virtual stream mapping
type UDPProxy struct {
	mu       sync.Mutex
	sessions map[string]net.Conn
	logger   *zap.Logger
}

func NewUDPProxy() *UDPProxy {
	l, _ := zap.NewProduction()
	return &UDPProxy{
		sessions: make(map[string]net.Conn),
		logger:   l,
	}
}

func (p *UDPProxy) ListenAndForward(publicPort int, session *yamux.Session) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", publicPort))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	p.logger.Info("UDP Proxy listening", zap.Int("port", publicPort))

	go func() {
		defer conn.Close()
		buf := make([]byte, 65507)
		for {
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			data := buf[:n]

			// Get or create virtual stream for this remote address
			stream, err := p.getStream(remoteAddr.String(), session)
			if err != nil {
				continue
			}

			// Send data over stream (requires framing on client side if handled properly)
			// For basic UDP, we simple write the packet data
			stream.Write(data)

			// Note: This is an oversimplification.
			// UDP to TCP (Yamux) works for 1:1 sessions, but needs better management.
		}
	}()

	return nil
}

func (p *UDPProxy) getStream(addr string, session *yamux.Session) (net.Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if s, exists := p.sessions[addr]; exists {
		return s, nil
	}

	stream, err := session.OpenStream()
	if err != nil {
		return nil, err
	}

	p.sessions[addr] = stream

	// Background reader for stream -> UDP
	// go p.streamToUDP(stream, addr)

	return stream, nil
}
