package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	"go.uber.org/zap"
)

var tcpProxyBufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 32*1024) // 32KB
		return &b
	},
}

// TCPProxy handles raw TCP traffic forwarding
type TCPProxy struct {
	logger *zap.Logger
}

func NewTCPProxy() *TCPProxy {
	l, _ := zap.NewProduction()
	return &TCPProxy{logger: l}
}

// Start listener for a specific port and link it to a yamux session
func (p *TCPProxy) ListenAndForward(publicPort int, session *yamux.Session) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", publicPort))
	if err != nil {
		return fmt.Errorf("TCP listener error on port %d: %w", publicPort, err)
	}

	p.logger.Info("TCP Proxy listening", zap.Int("port", publicPort))

	// Ensure listener is closed if yamux session dies without new incoming accepts.
	go func() {
		<-session.CloseChan()
		_ = listener.Close()
	}()

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
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				p.logger.Error("TCP Accept error", zap.Error(err))
				continue
			}

			if tcpConn, ok := conn.(*net.TCPConn); ok {
				_ = tcpConn.SetNoDelay(true)
			}

			// Open a new Yamux stream to the client
			stream, err := session.OpenStream()
			if err != nil {
				p.logger.Error("Failed to open Yamux stream", zap.Error(err))
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

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetNoDelay(true)
	}

	done := make(chan struct{}, 2)

	// Local internet -> Yamux stream -> Client
	go func() {
		bufPtr := tcpProxyBufPool.Get().(*[]byte)
		defer tcpProxyBufPool.Put(bufPtr)
		if _, err := io.CopyBuffer(stream, conn, *bufPtr); err != nil {
			p.logger.Debug("TCP copy conn->stream ended", zap.Error(err))
		}
		done <- struct{}{}
	}()

	// Client -> Yamux stream -> Local internet
	go func() {
		bufPtr := tcpProxyBufPool.Get().(*[]byte)
		defer tcpProxyBufPool.Put(bufPtr)
		if _, err := io.CopyBuffer(conn, stream, *bufPtr); err != nil {
			p.logger.Debug("TCP copy stream->conn ended", zap.Error(err))
		}
		done <- struct{}{}
	}()

	<-done
	<-done
}
