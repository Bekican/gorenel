package server

import (
	"bytes"
	"testing"
)

func TestHTTPProxy_PanicRecovery(t *testing.T) {
	// Stub test
	if t == nil {
		return
	}
}

func TestBoundedWriter(t *testing.T) {
	if t == nil {
		return
	}
	var buf bytes.Buffer
	bw := &BoundedWriter{W: &buf, Limit: 10}

	data := []byte("123456789012345") // 15 bytes
	n, err := bw.Write(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != 15 {
		t.Errorf("Expected n=15 (original length), got %d", n)
	}
	if buf.Len() != 10 {
		t.Errorf("Expected buffer length 10, got %d", buf.Len())
	}
	if buf.String() != "1234567890" {
		t.Errorf("Expected content '1234567890', got '%s'", buf.String())
	}
}
