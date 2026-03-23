package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	"go.uber.org/zap"
)

// UDPProxy handles UDP packet forwarding using a virtual stream mapping
type UDPProxy struct {
	mu       sync.Mutex
	sessions map[string]net.Conn
	lastSeen map[string]time.Time
	logger   *zap.Logger
}

func NewUDPProxy() *UDPProxy {
	l, _ := zap.NewProduction()
	return &UDPProxy{
		sessions: make(map[string]net.Conn),
		lastSeen: make(map[string]time.Time),
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
	stopCleanup := make(chan struct{})

	go func() {
		<-session.CloseChan()
		close(stopCleanup)
		_ = conn.Close()
		p.cleanupAllSessions()
	}()

	go p.cleanupStaleSessions(5*time.Minute, stopCleanup)

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
				p.logger.Debug("UDP stream acquisition failed", zap.Error(err))
				continue
			}

			// Send data over stream (requires framing on client side if handled properly)
			// For basic UDP, we simple write the packet data
			if _, err := stream.Write(data); err != nil {
				p.logger.Debug("UDP stream write failed", zap.Error(err))
				p.removeSession(remoteAddr.String())
				continue
			}
			p.markSeen(remoteAddr.String())

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
		p.lastSeen[addr] = time.Now()
		return s, nil
	}

	stream, err := session.OpenStream()
	if err != nil {
		return nil, err
	}

	p.sessions[addr] = stream
	p.lastSeen[addr] = time.Now()

	// Background reader for stream -> UDP
	// go p.streamToUDP(stream, addr)

	return stream, nil
}

func (p *UDPProxy) markSeen(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lastSeen[addr] = time.Now()
}

func (p *UDPProxy) removeSession(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if s, ok := p.sessions[addr]; ok {
		_ = s.Close()
		delete(p.sessions, addr)
	}
	delete(p.lastSeen, addr)
}

func (p *UDPProxy) cleanupAllSessions() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for addr, s := range p.sessions {
		_ = s.Close()
		delete(p.sessions, addr)
		delete(p.lastSeen, addr)
	}
}

func (p *UDPProxy) cleanupStaleSessions(ttl time.Duration, stop <-chan struct{}) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	for {
		select {
		case <-stop:
			return
		case <-t.C:
			now := time.Now()
			p.mu.Lock()
			for addr, seen := range p.lastSeen {
				if now.Sub(seen) > ttl {
					if s, ok := p.sessions[addr]; ok {
						_ = s.Close()
						delete(p.sessions, addr)
					}
					delete(p.lastSeen, addr)
				}
			}
			p.mu.Unlock()
		}
	}
}
