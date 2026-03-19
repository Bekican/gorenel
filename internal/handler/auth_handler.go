package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Bekican/gorenel/internal/middleware"
	"github.com/Bekican/gorenel/internal/authmgr"
	"github.com/Bekican/gorenel/pkg/auth"
	"github.com/Bekican/gorenel/pkg/errors"
	"github.com/Bekican/gorenel/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	oauthProviders map[string]auth.OAuthProvider
	tokenSvc       *auth.JWTService
	userRepo       auth.UserRepository
	authMgr        *authmgr.AuthManager
	isProd         bool
}

func NewAuthHandler(providers map[string]auth.OAuthProvider, tokenSvc *auth.JWTService, repo auth.UserRepository, authMgr *authmgr.AuthManager, isProd bool) *AuthHandler {
	return &AuthHandler{
		oauthProviders: providers,
		tokenSvc:       tokenSvc,
		userRepo:       repo,
		authMgr:        authMgr,
		isProd:         isProd,
	}
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	// 1. Check for Demo user bypass
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err == nil && credentials.Email != "" {
		// 1. Check if user exists locally
		user, err := h.userRepo.GetByEmail(credentials.Email)
		if err == nil {
			// 2. Verify Password
			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {
				logger.Warn("Failed login attempt: invalid password", zap.String("email", credentials.Email))
				return errors.Unauthorized("Invalid credentials")
			}

			logger.Info("Manual login successful", zap.String("email", user.Email))
			
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
					"name":  user.Name,
				},
			})
			return nil
		}
	}

	// 2. Standard OAuth Flow (Fallback or Explicit)
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "google" // Default fallback
	}

	oauth, exists := h.oauthProviders[provider]
	if !exists {
		logger.Error("OAuth provider not supported", zap.String("provider", provider))
		return errors.BadRequest("Unsupported OAuth provider", nil)
	}

	state := uuid.New().String()
	// Store provider in state or cookie so callback knows which one to use
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_provider",
		Value:    provider,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   h.isProd,
		Path:     "/",
		Domain:   ".gorenel.site", // Allow sharing between www and apex
		SameSite: http.SameSiteLaxMode,
	})

	// Store state in a cookie for validation later
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   h.isProd,
		Path:     "/",
		Domain:   ".gorenel.site", // Allow sharing between www and apex
		SameSite: http.SameSiteLaxMode,
	})

	url := oauth.GetAuthURL(state)

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
	if err != nil {
		logger.Warn("OAuth state cookie missing", 
			zap.Error(err), 
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("host", r.Host))
		return errors.Unauthorized("Invalid OAuth state")
	}

	if cookie.Value != r.FormValue("state") {
		logger.Warn("OAuth state mismatch", 
			zap.String("expected", cookie.Value), 
			zap.String("received", r.FormValue("state")))
		return errors.Unauthorized("Invalid OAuth state")
	}

	// 2. Get Provider from Cookie
	providerCookie, err := r.Cookie("oauth_provider")
	if err != nil {
		return errors.Unauthorized("OAuth provider context missing")
	}
	provider := providerCookie.Value

	oauth, exists := h.oauthProviders[provider]
	if !exists {
		return errors.Unauthorized("Invalid OAuth provider")
	}

	// 3. Get User Profile from Provider
	code := r.FormValue("code")
	profile, err := oauth.GetUserProfile(code)
	if err != nil {
		logger.Error("OAuth failed", zap.Error(err), zap.String("provider", provider))
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
		Domain:   ".gorenel.site", // Allow sharing between www and apex
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

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Internal(err)
	}

	user := &auth.User{
		ID:           uuid.New().String(),
		Email:        body.Email,
		Name:         body.Name,
		PasswordHash: string(hashedPassword),
		Provider:     "manual",
		CreatedAt:    time.Now(),
	}

	if err := h.userRepo.Create(user); err != nil {
		logger.Error("User registration failed in repository", zap.Error(err), zap.String("email", user.Email))
		return errors.Internal(err)
	}

	// Create an initial API key for the user
	apiKey := authmgr.GenerateAPIKey("gk")
	h.authMgr.AddKey(apiKey, user.ID, 100)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User initialization complete",
		"user": map[string]string{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
	return nil
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) error {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		return errors.Unauthorized("Unauthorized")
	}

	// Fetch user's API keys
	var userApiKey string
	keys := h.authMgr.ListKeys()
	for _, k := range keys {
		if k.UserID == claims.UserID { // We need to check if claims has UserID
			userApiKey = k.Key
			break
		}
	}

	user := map[string]string{
		"email":   claims.Email,
		"api_key": userApiKey,
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
	})
}

func (h *AuthHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) error {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		return errors.Unauthorized("Unauthorized")
	}

	allKeys := h.authMgr.ListKeys()
	userKeys := make([]*auth.APIKey, 0)
	for _, k := range allKeys {
		if k.UserID == claims.UserID {
			userKeys = append(userKeys, k)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(userKeys)
}

func (h *AuthHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) error {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		return errors.Unauthorized("Unauthorized")
	}

	apiKey := authmgr.GenerateAPIKey("gk")
	h.authMgr.AddKey(apiKey, claims.UserID, 100)

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]string{"key": apiKey})
}

func (h *AuthHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) error {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		return errors.Unauthorized("Unauthorized")
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		return errors.BadRequest("Missing key", nil)
	}

	// Double check ownership
	info, exists := h.authMgr.GetKeyInfo(key)
	if !exists || info.UserID != claims.UserID {
		return errors.Forbidden("Access denied")
	}

	h.authMgr.RevokeKey(key)
	w.WriteHeader(http.StatusNoContent)
	return nil
}
