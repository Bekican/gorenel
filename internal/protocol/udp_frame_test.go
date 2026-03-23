package protocol

import (
	"bytes"
	"testing"
)

func TestUDPFrameRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	payload := []byte("hello-udp")
	if err := WriteUDPFrame(&buf, payload); err != nil {
		t.Fatal(err)
	}
	out, err := ReadUDPFrame(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(payload) {
		t.Fatalf("got %q want %q", out, payload)
	}
}

func TestUDPFrameEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteUDPFrame(&buf, nil); err != nil {
		t.Fatal(err)
	}
	out, err := ReadUDPFrame(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty, got %d bytes", len(out))
	}
}

func TestUDPFrameTooLarge(t *testing.T) {
	payload := make([]byte, MaxUDPFramePayload+1)
	var buf bytes.Buffer
	if err := WriteUDPFrame(&buf, payload); err == nil {
		t.Fatal("expected error for oversized payload")
	}
}
