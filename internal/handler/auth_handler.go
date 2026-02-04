package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Bekican/gorenel/pkg/auth"
	"github.com/google/uuid"
)

type AuthHandler struct {
	tokenSvc *auth.JWTService
	userRepo *InMemoryUserRepo
}

func NewAuthHandler(tokenSvc *auth.JWTService, repo *InMemoryUserRepo) *AuthHandler {
	return &AuthHandler{
		tokenSvc: tokenSvc,
		userRepo: repo,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  *auth.User `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Simple validation
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// In a real app, you would check password hash here
	// if !checkPasswordHash(req.Password, user.PasswordHash) { ... }

	tokenString, err := h.tokenSvc.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set as cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: tokenString,
		User:  user,
	})
}

// Register handler
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user auth.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()

	if err := h.userRepo.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// InMemoryUserRepo for quick demo
type InMemoryUserRepo struct {
	users map[string]*auth.User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	repo := &InMemoryUserRepo{users: make(map[string]*auth.User)}
	// Add a demo user
	repo.Create(&auth.User{
		ID:    "demo-1",
		Email: "demo@gorenel.io",
		Name:  "Demo User",
	})
	return repo
}

func (r *InMemoryUserRepo) GetByEmail(email string) (*auth.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, http.ErrNoCookie // Temporary use as "not found"
}

func (r *InMemoryUserRepo) Create(user *auth.User) error {
	r.users[user.ID] = user
	return nil
}
