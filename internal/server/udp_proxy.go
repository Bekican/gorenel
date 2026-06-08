package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/hashicorp/yamux"
	"go.uber.org/zap"
)

// UDPProxy handles UDP packet forwarding using a virtual stream mapping.
// Each internet peer address gets one yamux stream; datagrams are length-prefixed
// so packet boundaries are preserved over the byte stream.
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

			stream, err := p.getStream(remoteAddr.String(), remoteAddr, conn, session)
			if err != nil {
				p.logger.Debug("UDP stream acquisition failed", zap.Error(err))
				continue
			}

			if err := protocol.WriteUDPFrame(stream, data); err != nil {
				p.logger.Debug("UDP stream write failed", zap.Error(err))
				p.removeSession(remoteAddr.String())
				continue
			}
			p.markSeen(remoteAddr.String())
		}
	}()

	return nil
}

func (p *UDPProxy) getStream(addr string, remote *net.UDPAddr, udpConn *net.UDPConn, session *yamux.Session) (net.Conn, error) {
	p.mu.Lock()

	if s, exists := p.sessions[addr]; exists {
		p.lastSeen[addr] = time.Now()
		p.mu.Unlock()
		return s, nil
	}

	stream, err := session.OpenStream()
	if err != nil {
		p.mu.Unlock()
		return nil, err
	}

	p.sessions[addr] = stream
	p.lastSeen[addr] = time.Now()
	p.mu.Unlock()

	go p.streamToUDP(stream, udpConn, remote, addr)

	return stream, nil
}

func (p *UDPProxy) streamToUDP(stream net.Conn, udpConn *net.UDPConn, remote *net.UDPAddr, sessionKey string) {
	defer func() {
		_ = stream.Close()
		p.removeSession(sessionKey)
	}()

	for {
		payload, err := protocol.ReadUDPFrame(stream)
		if err != nil {
			return
		}
		if _, err := udpConn.WriteToUDP(payload, remote); err != nil {
			p.logger.Debug("UDP WriteToUDP failed", zap.Error(err))
			return
		}
	}
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
