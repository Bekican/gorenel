package mcp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestHelperProcess is a mock stdio process used for unit testing.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "tools/list") {
			fmt.Println(`{"jsonrpc":"2.0","result":{"tools":[{"name":"get_weather"}]},"id":1}`)
		} else if strings.Contains(line, "tools/call") {
			fmt.Println(`{"jsonrpc":"2.0","result":{"content":[{"type":"text","text":"sunny"}]},"id":2}`)
		} else {
			fmt.Println(`{"jsonrpc":"2.0","method":"notifications/echo","params":{"text":"` + line + `"}}`)
		}
	}
}

func TestMcpBridge(t *testing.T) {
	// Setup bridge to run ourselves with helper flag
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bridge := NewBridge(os.Args[0], []string{"-test.run=TestHelperProcess", "--"})
	bridge.cmd = exec.CommandContext(ctx, os.Args[0], "-test.run=TestHelperProcess", "--")
	bridge.cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	// Manually hook up pipes to bypass b.Start() command building but test the logic
	stdin, err := bridge.cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	stdout, err := bridge.cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderr, err := bridge.cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	bridge.stdin = stdin
	bridge.stdout = stdout
	bridge.stderr = stderr

	if err := bridge.cmd.Start(); err != nil {
		t.Fatalf("cmd start: %v", err)
	}

	// Read stdout in background
	go func() {
		defer bridge.Stop()
		scanner := bufio.NewScanner(bridge.stdout)
		for scanner.Scan() {
			bridge.broadcast(scanner.Text())
		}
	}()

	// Read stderr
	go func() {
		scanner := bufio.NewScanner(bridge.stderr)
		for scanner.Scan() {
			// Consume
		}
	}()

	go func() {
		_ = bridge.cmd.Wait()
		bridge.Stop()
	}()

	// Test Server-Sent Events (SSE) Endpoint
	ts := httptest.NewServer(bridge)
	defer ts.Close()

	// Simulate client SSE subscription in a separate goroutine
	sseConnected := make(chan struct{})
	messageReceived := make(chan string, 5)

	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		req, err := http.NewRequest("GET", ts.URL+"/sse", nil)
		if err != nil {
			t.Errorf("NewRequest failed: %v", err)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("SSE get failed: %v", err)
			return
		}
		defer resp.Body.Close()

		close(sseConnected)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(strings.TrimSpace(line), "data: ")
				messageReceived <- data
			}
		}
	}()

	// Wait for client connection
	select {
	case <-sseConnected:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for SSE client to connect")
	}

	// Read the first SSE message, which is the endpoint declaration
	select {
	case endpointMsg := <-messageReceived:
		if !strings.Contains(endpointMsg, "/message?connectionId=") {
			t.Errorf("Expected endpoint message containing relative POST path, got: %s", endpointMsg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for initial endpoint declaration")
	}

	// Test Message Delivery via POST
	testPayload := `{"jsonrpc":"2.0","method":"tools/list","id":1}`
	postResp, err := http.Post(ts.URL+"/message?connectionId=test", "application/json", bytes.NewBufferString(testPayload))
	if err != nil {
		t.Fatalf("POST message failed: %v", err)
	}
	if postResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got: %s", postResp.Status)
	}
	postResp.Body.Close()

	// Verify that the response printed to stdout is sent back to the SSE client
	select {
	case sseMsg := <-messageReceived:
		if !strings.Contains(sseMsg, "tools/list") && !strings.Contains(sseMsg, "get_weather") {
			t.Errorf("Expected tools/list response, got: %s", sseMsg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for RPC response over SSE channel")
	}

	// Clean up
	bridge.Stop()
}
