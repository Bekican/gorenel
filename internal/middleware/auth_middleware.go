package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Bekican/gorenel/pkg/auth"
)

type userContextKey string

const UserKey userContextKey = "user"

func RequireAuth(jwtSvc *auth.JWTService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			// 1. Check Cookie
			cookie, err := r.Cookie("auth_token")
			if err == nil {
				tokenString = cookie.Value
			}

			// 2. Fallback to Authorization Header (Bearer)
			if tokenString == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenString == "" {
				http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
				return
			}

			claims, err := jwtSvc.ValidateToken(tokenString)
			if err != nil {
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
