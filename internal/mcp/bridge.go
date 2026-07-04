package mcp

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

// Bridge translates stdio MCP server communication to SSE/HTTP transport.
type Bridge struct {
	Command string
	Args    []string

	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdinMu  sync.Mutex
	stdout   io.ReadCloser
	stderr   io.ReadCloser
	once     sync.Once
	stopChan chan struct{}

	subscribers map[string]chan string
	subMu       sync.RWMutex
}

// NewBridge creates a new instance of Bridge.
func NewBridge(command string, args []string) *Bridge {
	return &Bridge{
		Command:     command,
		Args:        args,
		subscribers: make(map[string]chan string),
		stopChan:    make(chan struct{}),
	}
}

// Start spawns the stdio process and begins piping stdout/stderr.
func (b *Bridge) Start(ctx context.Context) error {
	var cmd *exec.Cmd
	// Support running commands through a shell on Windows if they are batch files or npm wrappers
	if len(b.Args) == 0 {
		// If command contains spaces, split it
		parts := strings.Fields(b.Command)
		if len(parts) > 1 {
			cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
		} else {
			cmd = exec.CommandContext(ctx, b.Command)
		}
	} else {
		cmd = exec.CommandContext(ctx, b.Command, b.Args...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	log.Printf("Starting MCP subprocess: %s %v", b.Command, b.Args)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	b.cmd = cmd
	b.stdin = stdin
	b.stdout = stdout
	b.stderr = stderr

	// Read stdout in a background goroutine
	go func() {
		defer b.Stop()
		scanner := bufio.NewScanner(b.stdout)
		// Set scanner buffer up to 1MB to handle large JSON-RPC messages (e.g. tools/list)
		const maxCapacity = 1024 * 1024
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			line := scanner.Text()
			b.broadcast(line)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("MCP stdout scanner error: %v", err)
		}
	}()

	// Read stderr in a background goroutine
	go func() {
		scanner := bufio.NewScanner(b.stderr)
		for scanner.Scan() {
			log.Printf("[MCP Stderr] %s", scanner.Text())
		}
	}()

	// Wait for process termination
	go func() {
		_ = cmd.Wait()
		log.Println("MCP subprocess terminated")
		b.Stop()
	}()

	return nil
}

// Stop terminates the subprocess and closes all subscriber channels.
func (b *Bridge) Stop() {
	b.once.Do(func() {
		close(b.stopChan)
		if b.stdin != nil {
			b.stdin.Close()
		}
		if b.cmd != nil && b.cmd.Process != nil {
			_ = b.cmd.Process.Kill()
		}

		b.subMu.Lock()
		for id, ch := range b.subscribers {
			close(ch)
			delete(b.subscribers, id)
		}
		b.subMu.Unlock()
	})
}

// Done returns a channel that is closed when the bridge terminates.
func (b *Bridge) Done() <-chan struct{} {
	return b.stopChan
}

// broadcast sends a line to all active SSE subscribers.
func (b *Bridge) broadcast(line string) {
	b.subMu.RLock()
	defer b.subMu.RUnlock()

	for _, ch := range b.subscribers {
		select {
		case ch <- line:
		default:
			// Buffer full, drop or log
		}
	}
}

// Generate a random connection ID
func generateConnID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "fallback-id"
	}
	return hex.EncodeToString(bytes)
}

// ServeHTTP implements the http.Handler interface for the bridge.
func (b *Bridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for client tools (e.g. Claude Desktop, Cursor)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Token")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.URL.Path {
	case "/sse":
		b.handleSSE(w, r)
	case "/message":
		b.handleMessage(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (b *Bridge) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	connID := generateConnID()
	ch := make(chan string, 128)

	b.subMu.Lock()
	b.subscribers[connID] = ch
	b.subMu.Unlock()

	defer func() {
		b.subMu.Lock()
		delete(b.subscribers, connID)
		b.subMu.Unlock()
		close(ch)
	}()

	// Send initial endpoint message per MCP specification
	// The client will use this endpoint to POST JSON-RPC payloads
	endpointURL := fmt.Sprintf("/message?connectionId=%s", connID)
	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpointURL)
	flusher.Flush()

	for {
		select {
		case <-b.stopChan:
			return
		case <-r.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (b *Bridge) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	b.stdinMu.Lock()
	defer b.stdinMu.Unlock()

	if b.stdin == nil {
		http.Error(w, "MCP process not running", http.StatusServiceUnavailable)
		return
	}

	// Write message as a single line to the stdio process
	var writeErr error
	if _, err := b.stdin.Write(body); err != nil {
		writeErr = err
	} else if _, err := b.stdin.Write([]byte("\n")); err != nil {
		writeErr = err
	}

	if writeErr != nil {
		log.Printf("Failed to write to MCP stdin: %v", writeErr)
		http.Error(w, "Failed to deliver message to MCP server", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
