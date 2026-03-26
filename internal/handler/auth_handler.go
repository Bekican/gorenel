package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Bekican/gorenel/internal/authmgr"
	"github.com/Bekican/gorenel/internal/middleware"
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
	// Handle Manual Login (JSON POST)
	if r.Method == http.MethodPost {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			return errors.BadRequest("Geçersiz istek formatı", nil)
		}

		// Input validation
		credentials.Email = strings.TrimSpace(strings.ToLower(credentials.Email))
		if credentials.Email == "" || !strings.Contains(credentials.Email, "@") || !strings.Contains(credentials.Email, ".") {
			return errors.BadRequest("Geçersiz e-posta adresi", nil)
		}
		if credentials.Password == "" {
			return errors.BadRequest("Şifre gereklidir", nil)
		}

		// 1. Check if user exists locally
		user, err := h.userRepo.GetByEmail(credentials.Email)
		if err != nil {
			logger.Warn("Failed login attempt: user not found", zap.String("email", credentials.Email))
			return errors.Unauthorized("Geçersiz e-posta veya şifre")
		}

		// 2. Verify Password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {
			logger.Warn("Failed login attempt: invalid password", zap.String("email", credentials.Email))
			return errors.Unauthorized("Geçersiz e-posta veya şifre")
		}

		logger.Info("Manual login successful", zap.String("email", user.Email))

		// Generate Token
		tokenString, err := h.tokenSvc.GenerateToken(user)
		if err != nil {
			return errors.Internal(err)
		}

		// Set Cookie using helper
		h.setAuthCookie(w, r, tokenString)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user": map[string]string{
				"email": user.Email,
				"name":  user.Name,
			},
		})
		return nil
	}

	// Handle Social Login Redirection (GET)
	// Only allow GET for initial social login trigger
	if r.Method != http.MethodGet {
		return errors.BadRequest("Unsupported method for social login", nil)
	}

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

	// Derive cookie domain dynamically (same logic as setAuthCookie)
	oauthCookieDomain := ""
	if h.isProd {
		host := r.Host
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		if host != "localhost" && host != "127.0.0.1" {
			parts := strings.Split(host, ".")
			if len(parts) >= 2 {
				oauthCookieDomain = "." + strings.Join(parts[len(parts)-2:], ".")
			} else {
				oauthCookieDomain = "." + host
			}
		}
	}

	// Store provider in state or cookie so callback knows which one to use
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_provider",
		Value:    provider,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   h.isProd,
		Path:     "/",
		Domain:   oauthCookieDomain,
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
		Domain:   oauthCookieDomain,
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
	if err != nil || cookie.Value != r.FormValue("state") {
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

	// 5. Set JWT as HttpOnly Cookie using helper
	h.setAuthCookie(w, r, tokenString)

	// 6. Redirect to Frontend Dashboard
	http.Redirect(w, r, "/dashboard?login=success", http.StatusSeeOther)
	return nil
}

// Logout handles user logout by clearing the auth cookie
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	// Clear cookie using the same domain logic
	cookieDomain := ""
	if h.isProd {
		host := r.Host
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		if host != "localhost" && host != "127.0.0.1" {
			parts := strings.Split(host, ".")
			if len(parts) >= 2 {
				cookieDomain = "." + strings.Join(parts[len(parts)-2:], ".")
			} else {
				cookieDomain = "." + host
			}
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-24 * time.Hour),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isProd,
		Path:     "/",
		Domain:   cookieDomain,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
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

	// Input validation
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if body.Email == "" || !strings.Contains(body.Email, "@") || !strings.Contains(body.Email, ".") {
		return errors.BadRequest("Geçersiz e-posta adresi", nil)
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		return errors.BadRequest("İsim gereklidir", nil)
	}
	if len(body.Password) < 8 {
		return errors.BadRequest("Şifre en az 8 karakter olmalıdır", nil)
	}

	// Check if a user with this email already exists
	existingUser, _ := h.userRepo.GetByEmail(body.Email)
	if existingUser != nil {
		return errors.BadRequest("Bu e-posta adresi zaten kayıtlı", nil)
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

	// Automatic Login after Registration
	tokenString, err := h.tokenSvc.GenerateToken(user)
	if err == nil {
		h.setAuthCookie(w, r, tokenString)
	}

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

// setAuthCookie is a helper to set the authentication cookie consistently
func (h *AuthHandler) setAuthCookie(w http.ResponseWriter, r *http.Request, tokenString string) {
	cookieDomain := ""
	if h.isProd {
		// Try to extract base domain from host, but handle ports and localhost
		host := r.Host
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		if host != "localhost" && host != "127.0.0.1" {
			// If it's a subdomain like app.gorenel.site, set domain for .gorenel.site
			// This allows the cookie to be shared among subdomains
			parts := strings.Split(host, ".")
			if len(parts) >= 2 {
				cookieDomain = "." + strings.Join(parts[len(parts)-2:], ".")
			} else {
				cookieDomain = "." + host
			}
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   h.isProd,
		Path:     "/",
		Domain:   cookieDomain,
		SameSite: http.SameSiteLaxMode,
	})
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
