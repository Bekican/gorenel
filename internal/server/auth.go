package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// api key yapısı
type APIKey struct {
	Key        string
	UserID     string
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	UsageCount int64
	RateLimit  int //1 dakikada gelen request
}

type AuthManager struct {
	keys map[string]*APIKey
	mu   sync.RWMutex
}

func NewAuthManager() *AuthManager {
	am := &AuthManager{
		keys: make(map[string]*APIKey),
	}
	am.AddKey("demo-key-12345", "demo-user", 100)
	am.AddKey("premium-key-67890", "premium-user", 1000)

	return am
}

func (am *AuthManager) AddKey(key, userID string, rateLimit int) {
	am.mu.Lock()
	defer am.mu.Unlock()

	keyHash := hashKey(key)
	am.keys[keyHash] = &APIKey{
		Key:        key,
		UserID:     userID,
		CreatedAt:  time.Now(),
		ExpiresAt:  nil,
		UsageCount: 0,
		RateLimit:  rateLimit,
	}
}

func (am *AuthManager) ValidateKey(key string) (*APIKey, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	keyHash := hashKey(key)
	apiKey, exists := am.keys[keyHash]

	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}
	return apiKey, nil
}

// Increment usage
func (am *AuthManager) IncrementUsage(key string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	keyHash := hashKey(key)
	if apiKey, exists := am.keys[keyHash]; exists {
		apiKey.UsageCount++
	}
}

// Get key
func (am *AuthManager) GetKeyInfo(key string) (*APIKey, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	keyHash := hashKey(key)
	apiKey, exists := am.keys[keyHash]
	return apiKey, exists
}

// api keyi iptal et
func (am *AuthManager) RevokeKey(key string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	keyHash := hashKey(key)
	delete(am.keys, keyHash)
}

// Tüm api keyleri listeleyecek
func (am *AuthManager) ListKeys() []*APIKey {
	am.mu.RLock()
	defer am.mu.RUnlock()

	keys := make([]*APIKey, 0, len(am.keys))
	for _, key := range am.keys {
		keys = append(keys, key)
	}
	return keys
}

// APİ keyi hashledik
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// Yeni api keyi ürettik
func GenerateAPIKey(prefix string) string {
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", prefix, timestamp)))
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(hash[:16]))
}
