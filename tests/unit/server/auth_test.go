package server_test

import (
	"testing"
	"time"

	"github.com/Bekican/gorenel/internal/server"
	"github.com/Bekican/gorenel/tests"
	"github.com/stretchr/testify/assert"
)

func TestAuthManager_AddKey(t *testing.T) {
	helper := tests.NewTestHelper(t)
	authManager := server.NewAuthManager(nil)

	// Test adding a new API key
	testKey := tests.GenerateTestAPIKey()
	authManager.AddKey(testKey, "test-user", 100)

	// Verify key was added
	apiKey, err := authManager.ValidateKey(testKey)
	helper.RequireNoError(err, "Failed to validate newly added key")
	assert.Equal(t, "test-user", apiKey.UserID)
	assert.Equal(t, 100, apiKey.RateLimit)
}

func TestAuthManager_ValidateKey_NotFound(t *testing.T) {
	authManager := server.NewAuthManager(nil)

	// Test with non-existent key
	_, err := authManager.ValidateKey("non-existent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API key")
}

func TestAuthManager_ValidateKey_Expired(t *testing.T) {
	authManager := server.NewAuthManager(nil)
	testKey := tests.GenerateTestAPIKey()

	// Add key with expiration in the past
	authManager.AddKey(testKey, "test-user", 100)

	// Manually set expiration to past (this would need an AddKeyWithExpiration method)
	pastTime := time.Now().Add(-1 * time.Hour)
	apiKey, _ := authManager.GetKeyInfo(testKey)
	apiKey.ExpiresAt = &pastTime

	// Validate should fail
	_, err := authManager.ValidateKey(testKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestAuthManager_IncrementUsage(t *testing.T) {
	authManager := server.NewAuthManager(nil)
	testKey := tests.GenerateTestAPIKey()
	authManager.AddKey(testKey, "test-user", 100)

	// Increment usage
	authManager.IncrementUsage(testKey)
	authManager.IncrementUsage(testKey)

	// Check usage count
	apiKey, _ := authManager.GetKeyInfo(testKey)
	assert.Equal(t, int64(2), apiKey.UsageCount)
}

func TestAuthManager_RevokeKey(t *testing.T) {
	authManager := server.NewAuthManager(nil)
	testKey := tests.GenerateTestAPIKey()
	authManager.AddKey(testKey, "test-user", 100)

	// Revoke key
	authManager.RevokeKey(testKey)

	// Key should no longer be valid
	_, err := authManager.ValidateKey(testKey)
	assert.Error(t, err)
}

func TestAuthManager_Concurrent(t *testing.T) {
	authManager := server.NewAuthManager(nil)
	testKey := tests.GenerateTestAPIKey()
	authManager.AddKey(testKey, "test-user", 100)

	// Concurrent usage increments
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			authManager.IncrementUsage(testKey)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify count
	apiKey, _ := authManager.GetKeyInfo(testKey)
	assert.Equal(t, int64(100), apiKey.UsageCount)
}

// Benchmark tests
func BenchmarkAuthManager_ValidateKey(b *testing.B) {
	authManager := server.NewAuthManager(nil)
	testKey := tests.GenerateTestAPIKey()
	authManager.AddKey(testKey, "test-user", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authManager.ValidateKey(testKey)
	}
}

func BenchmarkAuthManager_IncrementUsage(b *testing.B) {
	authManager := server.NewAuthManager(nil)
	testKey := tests.GenerateTestAPIKey()
	authManager.AddKey(testKey, "test-user", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authManager.IncrementUsage(testKey)
	}
}
