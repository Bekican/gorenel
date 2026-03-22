package cmd

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ClientWSConn wraps a gorilla/websocket.Conn to implement net.Conn interface.
// Used by the client to establish yamux sessions over WebSocket.
type ClientWSConn struct {
	ws     *websocket.Conn
	reader io.Reader
	mu     sync.Mutex
}

// NewClientWSConn creates a new ClientWSConn adapter.
func NewClientWSConn(ws *websocket.Conn) *ClientWSConn {
	return &ClientWSConn{ws: ws}
}

func (c *ClientWSConn) Read(p []byte) (int, error) {
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
			c.reader = nil
			if n > 0 {
				return n, nil
			}
			continue
		}
		return n, err
	}
}

func (c *ClientWSConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.ws.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *ClientWSConn) Close() error {
	return c.ws.Close()
}

func (c *ClientWSConn) LocalAddr() net.Addr {
	return c.ws.LocalAddr()
}

func (c *ClientWSConn) RemoteAddr() net.Addr {
	return c.ws.RemoteAddr()
}

func (c *ClientWSConn) SetDeadline(t time.Time) error {
	if err := c.ws.SetReadDeadline(t); err != nil {
		return err
	}
	return c.ws.SetWriteDeadline(t)
}

func (c *ClientWSConn) SetReadDeadline(t time.Time) error {
	return c.ws.SetReadDeadline(t)
}

func (c *ClientWSConn) SetWriteDeadline(t time.Time) error {
	return c.ws.SetWriteDeadline(t)
}
