package server

import (
	"bytes"
	"io"
	"net/http"
)

type ResponseCaptureWriter struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
	maxBodyBytes int64
	capturedBytes int64
}

func NewResponseCaptureWriter(w http.ResponseWriter, maxBodyBytes int64) *ResponseCaptureWriter {
	return &ResponseCaptureWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Body:           &bytes.Buffer{},
		maxBodyBytes:   maxBodyBytes,
		capturedBytes:  0,
	}
}

func (rw *ResponseCaptureWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseCaptureWriter) Write(b []byte) (int, error) {
	// Always write the full response to the client, but only capture up to maxBodyBytes
	// for inspector/ML to avoid memory blowups.
	if rw.maxBodyBytes > 0 && rw.capturedBytes < rw.maxBodyBytes {
		remaining := rw.maxBodyBytes - rw.capturedBytes
		toCapture := int64(len(b))
		if toCapture > remaining {
			toCapture = remaining
		}
		if toCapture > 0 {
			n, _ := rw.Body.Write(b[:toCapture])
			rw.capturedBytes += int64(n)
		}
	}

	n, err := rw.ResponseWriter.Write(b)
	return n, err
}

func InterceptBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}
