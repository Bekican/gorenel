package tests

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/yamux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelper provides common test utilities
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// CreateTestServer creates a test HTTP server
func (h *TestHelper) CreateTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// CreateTestTCPServer creates a test TCP server
func (h *TestHelper) CreateTestTCPServer(port int, handler func(net.Conn)) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handler(conn)
		}
	}()

	return listener, nil
}

// CreateYamuxPair creates a client-server Yamux session pair
func (h *TestHelper) CreateYamuxPair() (*yamux.Session, *yamux.Session, error) {
	// Create in-memory connection
	serverConn, clientConn := net.Pipe()

	// Create Yamux sessions
	serverSession, err := yamux.Server(serverConn, yamux.DefaultConfig())
	if err != nil {
		return nil, nil, err
	}

	clientSession, err := yamux.Client(clientConn, yamux.DefaultConfig())
	if err != nil {
		serverSession.Close()
		return nil, nil, err
	}

	return serverSession, clientSession, nil
}

// WaitForCondition waits for a condition to be true
func (h *TestHelper) WaitForCondition(condition func() bool, timeout time.Duration, msg string) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return
		}
		if time.Now().After(deadline) {
			h.t.Fatalf("Timeout waiting for condition: %s", msg)
		}
		<-ticker.C
	}
}

// AssertHTTPStatus asserts HTTP response status
func (h *TestHelper) AssertHTTPStatus(resp *http.Response, expectedStatus int) {
	assert.Equal(h.t, expectedStatus, resp.StatusCode,
		"Expected status %d, got %d", expectedStatus, resp.StatusCode)
}

// RequireNoError fails test if error is not nil
func (h *TestHelper) RequireNoError(err error, msg string) {
	require.NoError(h.t, err, msg)
}

// MockHTTPResponse creates a mock HTTP response
func MockHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// GenerateTestAPIKey generates a test API key
func GenerateTestAPIKey() string {
	return "test-key-" + uuid.New().String()[:8]
}

// CleanupFunc is a function that cleans up test resources
type CleanupFunc func()

// Cleanup registers cleanup functions
func (h *TestHelper) Cleanup(fn CleanupFunc) {
	h.t.Cleanup(fn)
}
