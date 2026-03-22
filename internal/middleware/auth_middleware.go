package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Bekican/gorenel/pkg/auth"
)

type userContextKey string

const UserKey userContextKey = "user"

func RequireAuth(jwtSvc *auth.JWTService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				// Log ALL cookies for debugging
				allCookies := r.Header.Get("Cookie")
				fmt.Printf("[DEBUG] Auth failed: Missing auth_token cookie. ALL cookies received: '%s'\n", allCookies)
				http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
				return
			}

			claims, err := jwtSvc.ValidateToken(cookie.Value)
			if err != nil {
				// Log for debugging
				fmt.Printf("[DEBUG] Auth failed: Invalid token: %v\n", err)
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, claims)
			next(w, r.WithContext(ctx))
		}
	}
}

func GetUserFromContext(ctx context.Context) *auth.Claims {
	user, ok := ctx.Value(UserKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return user
}
