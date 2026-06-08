package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

// MaxUDPFramePayload is the maximum UDP datagram size we tunnel (IPv6 jumbo-safe cap).
const MaxUDPFramePayload = 65507

// WriteUDPFrame writes one datagram as: [4-byte big-endian length][payload].
func WriteUDPFrame(w io.Writer, payload []byte) error {
	n := len(payload)
	if n > MaxUDPFramePayload {
		return fmt.Errorf("udp frame payload too large: %d", n)
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(n))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	_, err := w.Write(payload)
	return err
}

// ReadUDPFrame reads one length-prefixed datagram from r.
func ReadUDPFrame(r io.Reader) ([]byte, error) {
	var ln uint32
	if err := binary.Read(r, binary.BigEndian, &ln); err != nil {
		return nil, err
	}
	if ln > MaxUDPFramePayload {
		return nil, fmt.Errorf("udp frame length invalid: %d", ln)
	}
	if ln == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, ln)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
