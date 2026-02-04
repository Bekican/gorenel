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
}

func NewResponseCaptureWriter(w http.ResponseWriter) *ResponseCaptureWriter {
	return &ResponseCaptureWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Body:           &bytes.Buffer{},
	}
}

func (rw *ResponseCaptureWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseCaptureWriter) Write(b []byte) (int, error) {
	rw.Body.Write(b)
	return rw.ResponseWriter.Write(b)
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
