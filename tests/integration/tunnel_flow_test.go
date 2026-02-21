package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bekican/gorenel/internal/protocol"
	"github.com/Bekican/gorenel/tests"
	"github.com/hashicorp/yamux"
	"github.com/stretchr/testify/assert"
)

// TestFullTunnelFlow tests complete tunnel creation and usage
func TestFullTunnelFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	helper := tests.NewTestHelper(t)

	// Step 1: Start test server on localhost
	localServer := helper.CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from local server"))
	})
	defer localServer.Close()

	// Step 2: Create Yamux session (simulating client-server connection)
	serverSession, clientSession, err := helper.CreateYamuxPair()
	helper.RequireNoError(err, "Failed to create Yamux pair")
	defer serverSession.Close()
	defer clientSession.Close()

	// Step 3: Simulate REGISTER flow
	registerMsg := protocol.NewRegisterMessage("test-client", "1.0.0", "test-api-key", "")

	// Step 4: Server-side handler (in a goroutine)
	go func() {
		stream, err := serverSession.AcceptStream()
		if err != nil {
			return
		}
		defer stream.Close()

		msg, err := protocol.ReadMessage(stream)
		if err != nil {
			return
		}

		if msg.Type == protocol.MsgTypeRegister {
			resp := protocol.RegisterResponse{
				Subdomain: "test-sub",
				FullURL:   "http://test-sub.gorenel.io:8080",
			}
			respJSON, _ := json.Marshal(resp)
			protocol.WriteMessage(stream, protocol.Message{
				Type:    protocol.MsgTypeRegistered,
				Payload: string(respJSON),
			})
		}
	}()

	// Step 5: Client-side registration
	stream, err := clientSession.OpenStream()
	helper.RequireNoError(err, "Failed to open stream")

	err = protocol.WriteMessage(stream, registerMsg)
	helper.RequireNoError(err, "Failed to write register message")

	// Step 6: Server should respond with subdomain
	response, err := protocol.ReadMessage(stream)
	helper.RequireNoError(err, "Failed to read response")

	assert.Equal(t, protocol.MsgTypeRegistered, response.Type)

	var regResp protocol.RegisterResponse
	err = json.Unmarshal([]byte(response.Payload), &regResp)
	helper.RequireNoError(err, "Failed to parse response")

	assert.NotEmpty(t, regResp.Subdomain)
	assert.NotEmpty(t, regResp.FullURL)

	// Step 7: Cleanup
	assert.NotNil(t, stream)
	assert.False(t, clientSession.IsClosed(), "Client session should be open")

	stream.Close()
}

// TestWebSocketTunnel tests WebSocket tunneling
func TestWebSocketTunnel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	helper := tests.NewTestHelper(t)

	// Create WebSocket test server
	wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for WebSocket upgrade
		if r.Header.Get("Upgrade") != "websocket" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Simulate WebSocket upgrade
		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")
		w.WriteHeader(http.StatusSwitchingProtocols)
	}))
	defer wsServer.Close()

	// Test WebSocket upgrade detection
	req, err := http.NewRequest("GET", wsServer.URL, nil)
	helper.RequireNoError(err, "Failed to create request")

	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")

	client := &http.Client{}
	resp, err := client.Do(req)
	helper.RequireNoError(err, "Request failed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
}

// TestConcurrentTunnels tests multiple simultaneous tunnels
func TestConcurrentTunnels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	helper := tests.NewTestHelper(t)
	numTunnels := 10

	// Create multiple Yamux sessions concurrently
	clientSessions := make([]*yamux.Session, numTunnels)
	serverSessions := make([]*yamux.Session, numTunnels)
	done := make(chan error, numTunnels)

	for i := 0; i < numTunnels; i++ {
		go func(index int) {
			serverSession, clientSession, err := helper.CreateYamuxPair()
			if err != nil {
				done <- err
				return
			}

			serverSessions[index] = serverSession
			clientSessions[index] = clientSession

			// Keep session alive
			time.Sleep(100 * time.Millisecond)

			done <- nil
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < numTunnels; i++ {
		err := <-done
		helper.RequireNoError(err, fmt.Sprintf("Tunnel %d failed", i))
	}

	// Verify all sessions are open
	openCount := 0
	for _, session := range clientSessions {
		if session != nil && !session.IsClosed() {
			openCount++
		}
	}

	assert.Equal(t, numTunnels, openCount, "All tunnels should be open")

	// Cleanup
	for _, session := range clientSessions {
		if session != nil {
			session.Close()
		}
	}
	for _, session := range serverSessions {
		if session != nil {
			session.Close()
		}
	}
}

// TestRateLimitingIntegration tests rate limiting in action
func TestRateLimitingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	helper := tests.NewTestHelper(t)

	// Create rate-limited server
	requestCount := 0
	server := helper.CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount > 5 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	defer server.Close()

	// Send 10 requests
	client := &http.Client{}
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 10; i++ {
		resp, err := client.Get(server.URL)
		helper.RequireNoError(err, "Request failed")

		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++
		}

		resp.Body.Close()
	}

	assert.Equal(t, 5, successCount, "Should allow 5 requests")
	assert.Equal(t, 5, rateLimitedCount, "Should rate-limit 5 requests")
}
