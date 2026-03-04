package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Bekican/gorenel/internal/middleware"
	"github.com/Bekican/gorenel/pkg/auth"
	"github.com/Bekican/gorenel/pkg/errors"
	"github.com/Bekican/gorenel/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthHandler struct {
	oauth    auth.OAuthProvider
	tokenSvc *auth.JWTService
	userRepo auth.UserRepository
	isProd   bool
}

func NewAuthHandler(oauth auth.OAuthProvider, tokenSvc *auth.JWTService, repo auth.UserRepository, isProd bool) *AuthHandler {
	return &AuthHandler{
		oauth:    oauth,
		tokenSvc: tokenSvc,
		userRepo: repo,
		isProd:   isProd,
	}
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	// 1. Check for Demo user bypass
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err == nil {
		if credentials.Email == "demo@gorenel.io" && !h.isProd {
			// Find or create demo user
			user, err := h.userRepo.GetByEmail(credentials.Email)
			if err != nil {
				user = &auth.User{
					ID:        "demo-user-id",
					Email:     credentials.Email,
					Name:      "Demo User",
					Provider:  "demo",
					CreatedAt: time.Now(),
				}
				h.userRepo.Create(user)
			}

			// Generate Token
			tokenString, err := h.tokenSvc.GenerateToken(user)
			if err != nil {
				return errors.Internal(err)
			}

			// Set Cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    tokenString,
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
				Secure:   h.isProd,
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			})

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"user": map[string]string{
					"email": user.Email,
				},
			})
			return nil
		}
	}

	// 2. Standard OAuth Flow
	state := uuid.New().String()

	// Store state in a cookie for validation later
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   h.isProd,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	if h.oauth == nil {
		logger.Error("OAuth provider not initialized")
		http.Error(w, "OAuth provider not initialized", http.StatusInternalServerError)
		return nil
	}

	url := h.oauth.GetAuthURL(state)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"redirect_url": url,
	})
	return nil
}

// Callback handles the redirect from Google
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) error {
	// 1. Validate State (Anti-CSRF)
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != r.FormValue("state") {
		return errors.Unauthorized("Invalid OAuth state")
	}

	// 2. Get User Profile from Provider
	code := r.FormValue("code")
	profile, err := h.oauth.GetUserProfile(code)
	if err != nil {
		logger.Error("OAuth failed", zap.Error(err))
		return errors.Unauthorized("Authentication failed")
	}

	// 3. Find or Create User in DB
	user, err := h.userRepo.GetByEmail(profile.Email)
	if err != nil {
		// Assume "not found" means we create a new user
		user = &auth.User{
			ID:        uuid.New().String(),
			Email:     profile.Email,
			Name:      profile.Name,
			AvatarURL: profile.AvatarURL,
			Provider:  profile.Provider,
			CreatedAt: time.Now(),
		}
		if err := h.userRepo.Create(user); err != nil {
			return errors.Internal(err)
		}
	}

	// 4. Generate JWT
	tokenString, err := h.tokenSvc.GenerateToken(user)
	if err != nil {
		return errors.Internal(err)
	}

	// 5. Set JWT as HttpOnly Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   h.isProd,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// 6. Redirect to Frontend Dashboard
	http.Redirect(w, r, "/dashboard?login=success", http.StatusSeeOther)
	return nil
}

// Register handles manual user registration (simulation)
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return errors.BadRequest("Invalid request body", err)
	}

	user := &auth.User{
		ID:        uuid.New().String(),
		Email:     body.Email,
		Name:      body.Name,
		Provider:  "manual",
		CreatedAt: time.Now(),
	}

	if err := h.userRepo.Create(user); err != nil {
		return errors.Internal(err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created", "uid": user.ID})
	return nil
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) error {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		return errors.Unauthorized("Unauthorized")
	}

	user := map[string]string{
		"email": claims.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
	})
}
