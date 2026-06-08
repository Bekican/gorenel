package middleware

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// PanicRecovery wraps an HTTP handler with recover() so that any panic
// is caught, logged, and turned into a 500 response instead of crashing
// the entire server process.
//
// Why this matters:
// Without this middleware, a single nil pointer or index-out-of-range
// in any handler will kill the entire Go process, taking down all tunnels.
// With it, the crashing request returns 500 and the server stays alive.
func PanicRecovery(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// Capture full stack trace for debugging
					stack := string(debug.Stack())

					logger.Error("PANIC RECOVERED",
						zap.Any("panic", rec),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("remote_addr", r.RemoteAddr),
						zap.String("stack_trace", stack),
					)

					// Return 500 to the client
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// PanicRecoveryFunc is a convenience wrapper for http.HandlerFunc patterns.
func PanicRecoveryFunc(logger *zap.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := string(debug.Stack())

				logger.Error("PANIC RECOVERED",
					zap.Any("panic", rec),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("stack_trace", stack),
				)

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}
