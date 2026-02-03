package logger

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
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

		//log detaylandırması
		wrapped := newResponseWriter(w)

		reqLogger := WithRequestID(requestID)

		//log request
		reqLogger.Debug("request-started",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)

		//işlem isteği
		next.ServeHTTP(wrapped, r)

		//geçen zamanı hesaplama
		duration := time.Since(start)

		reqLogger.Info("request_completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status_code", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.Int64("duration_ms", duration.Milliseconds()),
		)
	})
}

// hata öncesi önlem alınan fonksiyon
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := w.Header().Get("X-Request-ID")

				Error("panic_recovered",
					zap.Any("error", err),
					zap.String("request_id", requestID),
					zap.String("method", r.Method),
				)

				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// middlewareleri birbirine bağlar.
func ChainMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
