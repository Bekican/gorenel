package authmgr

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/Bekican/gorenel/pkg/auth"
)

type AuthManager struct {
	keys map[string]*auth.APIKey
	repo auth.APIKeyRepository
	mu   sync.RWMutex
}

func NewAuthManager(repo auth.APIKeyRepository) *AuthManager {
	am := &AuthManager{
		keys: make(map[string]*auth.APIKey),
		repo: repo,
	}

	// Load existing keys from DB
	if repo != nil {
		keys, err := repo.ListAll()
		if err == nil {
			for _, k := range keys {
				keyHash := HashKey(k.Key)
				am.keys[keyHash] = k
			}
		}
	} else {
		// Fallback for dev mode
		am.AddKey("demo-key-12345", "demo-user", 100)
		am.AddKey("premium-key-67890", "premium-user", 1000)
	}

	return am
}

func (am *AuthManager) AddKey(key, userID string, rateLimit int) {
	am.mu.Lock()
	defer am.mu.Unlock()

	apiKey := &auth.APIKey{
		Key:        key,
		UserID:     userID,
		CreatedAt:  time.Now(),
		ExpiresAt:  nil,
		UsageCount: 0,
		RateLimit:  rateLimit,
	}

	keyHash := HashKey(key)
	am.keys[keyHash] = apiKey

	if am.repo != nil {
		am.repo.Create(apiKey)
	}
}

func (am *AuthManager) ValidateKey(key string) (*auth.APIKey, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	keyHash := HashKey(key)
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

	keyHash := HashKey(key)
	if apiKey, exists := am.keys[keyHash]; exists {
		apiKey.UsageCount++
		if am.repo != nil {
			am.repo.IncrementUsage(keyHash)
		}
	}
}

// Get key
func (am *AuthManager) GetKeyInfo(key string) (*auth.APIKey, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	keyHash := HashKey(key)
	apiKey, exists := am.keys[keyHash]
	return apiKey, exists
}

// api keyi iptal et
func (am *AuthManager) RevokeKey(key string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	keyHash := HashKey(key)
	delete(am.keys, keyHash)
	if am.repo != nil {
		am.repo.Delete(keyHash)
	}
}

// Tüm api keyleri listeleyecek
func (am *AuthManager) ListKeys() []*auth.APIKey {
	am.mu.RLock()
	defer am.mu.RUnlock()

	keys := make([]*auth.APIKey, 0, len(am.keys))
	for _, key := range am.keys {
		keys = append(keys, key)
	}
	return keys
}

// APİ keyi hashledik
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// Yeni api keyi ürettik — kriptografik rastgelelik kullanarak
func GenerateAPIKey(prefix string) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback: should never happen on modern OS
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b[:16]))
}
