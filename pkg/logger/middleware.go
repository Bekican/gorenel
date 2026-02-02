package logger

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// response writer http.ResponseWriterı kapsıyor -> statusCode'u öğrenmek için
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// http loglarını yazar -> zamanlama ve içerikleriyle
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := r.Header.Get("X-Request")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		//response headerına request id ekledik
		w.Header().Set("X-Request-ID", requestID)

	})
}
