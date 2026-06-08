package middleware

import (
	"net"
	"net/http"

	"github.com/Bekican/gorenel/internal/limiter"
)

// RateLimitMiddleware protects routes using the Advanced Rate Limiter
func RateLimitMiddleware(rl *limiter.RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 1. Determine Identifier (UserID or IP)
			var identifier string
			user := GetUserFromContext(r.Context())

			if user != nil {
				identifier = user.UserID
			} else {
				// Fallback to IP address
				ip, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					identifier = r.RemoteAddr
				} else {
					identifier = ip
				}
			}

			// 2. Check Rate Limit (1 request)
			if !rl.Allow(identifier, 1) {
				w.Header().Set("X-RateLimit-Status", "Exceeded")
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			// 3. Continue to next handler
			next(w, r)
		}
	}
}
