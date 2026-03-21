package server

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSConn wraps a gorilla/websocket.Conn to implement net.Conn interface.
// This allows yamux and protocol code to work transparently over WebSocket.
type WSConn struct {
	ws     *websocket.Conn
	reader io.Reader
	mu     sync.Mutex
}

// NewWSConn creates a new WSConn adapter from a gorilla websocket connection.
func NewWSConn(ws *websocket.Conn) *WSConn {
	return &WSConn{ws: ws}
}

func (c *WSConn) Read(p []byte) (int, error) {
	for {
		if c.reader == nil {
			_, reader, err := c.ws.NextReader()
			if err != nil {
				return 0, err
			}
			c.reader = reader
		}

		n, err := c.reader.Read(p)
		if err == io.EOF {
			// Current message consumed, get next one
			c.reader = nil
			if n > 0 {
				return n, nil
			}
			continue
		}
		return n, err
	}
}

func (c *WSConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.ws.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *WSConn) Close() error {
	return c.ws.Close()
}

func (c *WSConn) LocalAddr() net.Addr {
	return c.ws.LocalAddr()
}

func (c *WSConn) RemoteAddr() net.Addr {
	return c.ws.RemoteAddr()
}

func (c *WSConn) SetDeadline(t time.Time) error {
	if err := c.ws.SetReadDeadline(t); err != nil {
		return err
	}
	return c.ws.SetWriteDeadline(t)
}

func (c *WSConn) SetReadDeadline(t time.Time) error {
	return c.ws.SetReadDeadline(t)
}

func (c *WSConn) SetWriteDeadline(t time.Time) error {
	return c.ws.SetWriteDeadline(t)
}
