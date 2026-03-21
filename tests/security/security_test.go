package security_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bekican/gorenel/internal/authmgr"
	"github.com/Bekican/gorenel/internal/limiter"
	"github.com/Bekican/gorenel/internal/server"
	"github.com/Bekican/gorenel/tests"
	"github.com/stretchr/testify/assert"
)

// ===== API KEY VALIDATION =====

func TestSecurity_InvalidAPIKeyRejection(t *testing.T) {
	authManager := authmgr.NewAuthManager(nil)

	_, err := authManager.ValidateKey("completely-invalid-key")
	assert.Error(t, err, "Invalid API key should be rejected")
	assert.Contains(t, err.Error(), "invalid", "Error should mention invalid key")
}

func TestSecurity_EmptyKeyRejection(t *testing.T) {
	authManager := authmgr.NewAuthManager(nil)

	_, err := authManager.ValidateKey("")
	assert.Error(t, err, "Empty API key should be rejected")
}

func TestSecurity_ExpiredKeyRejection(t *testing.T) {
	authManager := authmgr.NewAuthManager(nil)
	testKey := tests.GenerateTestAPIKey()
	authManager.AddKey(testKey, "user-1", 100)

	// Manually expire the key
	pastTime := time.Now().Add(-24 * time.Hour)
	keyInfo, _ := authManager.GetKeyInfo(testKey)
	keyInfo.ExpiresAt = &pastTime

	_, err := authManager.ValidateKey(testKey)
	assert.Error(t, err, "Expired key should be rejected")
}

func TestSecurity_RevokedKeyRejection(t *testing.T) {
	authManager := authmgr.NewAuthManager(nil)
	testKey := tests.GenerateTestAPIKey()
	authManager.AddKey(testKey, "user-1", 100)

	// Verify key works first
	_, err := authManager.ValidateKey(testKey)
	assert.NoError(t, err, "Key should be valid before revocation")

	// Revoke it
	authManager.RevokeKey(testKey)

	// Should now fail
	_, err = authManager.ValidateKey(testKey)
	assert.Error(t, err, "Revoked key should be rejected")
}

// ===== RATE LIMIT BYPASS PROTECTION =====

func TestSecurity_RateLimitCannotBypass(t *testing.T) {
	rl := limiter.NewRateLimiter(nil, 100, time.Minute)

	// Exhaust limit for user
	for i := 0; i < 100; i++ {
		rl.Allow("attacker", 1)
	}

	// Verify blocked
	assert.False(t, rl.Allow("attacker", 1), "Should be blocked after 100 requests")

	// Try variations — same user ID should still be blocked
	assert.False(t, rl.Allow("attacker", 1), "Double attempt should be blocked")
	assert.False(t, rl.Allow("attacker", 1), "Triple attempt should be blocked")

	// Different user should NOT be affected
	assert.True(t, rl.Allow("legitimate-user", 1), "Different user should not be blocked")
}

func TestSecurity_RateLimitPerUserIsolation(t *testing.T) {
	rl := limiter.NewRateLimiter(nil, 100, time.Minute)

	// Exhaust one user
	for i := 0; i < 100; i++ {
		rl.Allow("attacker-1", 1)
	}

	// Other users should be completely independent
	for i := 0; i < 5; i++ {
		user := tests.GenerateTestAPIKey() // random user ID
		assert.True(t, rl.Allow(user, 1), "Unrelated user should not be rate limited")
	}
}

// ===== HEADER INJECTION PROTECTION =====

func TestSecurity_HeaderInjectionViaModifier(t *testing.T) {
	mod := server.NewTrafficModifier()

	// Attempt to inject malicious headers
	mod.AddRule(server.ModificationRule{
		ID:          "safe-rule",
		PathPattern: "/api/*",
		AddHeaders:  map[string]string{"X-Safe": "true"},
	})

	// Verify that only defined rules apply
	req := httptest.NewRequest("GET", "http://example.com/api/test", nil)
	req.Header.Set("X-Original", "preserved")
	mod.Apply(req)

	assert.Equal(t, "true", req.Header.Get("X-Safe"), "Defined header should be set")
	assert.Equal(t, "preserved", req.Header.Get("X-Original"), "Original headers should be preserved")
}

func TestSecurity_ModifierRemoveProtection(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{
		ID:            "strip",
		PathPattern:   "/api/*",
		RemoveHeaders: []string{"Authorization"},
	})

	req := httptest.NewRequest("GET", "http://example.com/api/admin", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("Content-Type", "application/json")

	mod.Apply(req)

	// Authorization should be stripped (e.g., for internal forwarding)
	assert.Empty(t, req.Header.Get("Authorization"), "Sensitive header should be removed")
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "Other headers preserved")
}

// ===== PATH TRAVERSAL PROTECTION =====

func TestSecurity_PathTraversalInModifier(t *testing.T) {
	mod := server.NewTrafficModifier()
	mod.AddRule(server.ModificationRule{
		ID:          "restricted",
		PathPattern: "/api/*",
		AddHeaders:  map[string]string{"X-Internal": "true"},
	})

	// Attempt path traversal
	maliciousPaths := []string{
		"/../etc/passwd",
		"/../../secret",
		"/%2e%2e/admin",
		"/admin/../../../etc/shadow",
	}

	for _, path := range maliciousPaths {
		req := httptest.NewRequest("GET", "http://example.com"+path, nil)
		mod.Apply(req)

		// Path traversal attempts should NOT match /api/* pattern
		assert.Empty(t, req.Header.Get("X-Internal"),
			"Path traversal '%s' should not match /api/* rule", path)
	}
}

// ===== SQL INJECTION IN PATHS =====

func TestSecurity_SQLInjectionInInspectorPaths(t *testing.T) {
	ti := server.NewTrafficInspector(100)

	// Simulate requests with SQL injection payloads
	injectionPayloads := []string{
		"/api/users?id=1 OR 1=1",
		"/api/users'; DROP TABLE users;--",
		"/api/users/<script>alert('xss')</script>",
		"/api/users?q=1 UNION SELECT * FROM passwords",
	}

	for _, payload := range injectionPayloads {
		ti.Record(&server.CapturedRequest{
			ID:     "sql-" + payload[:10],
			Method: "GET",
			Path:   payload,
		})
	}

	// Inspector should store them safely without executing
	history := ti.GetHistory()
	assert.Len(t, history, len(injectionPayloads), "All payloads should be stored safely")

	// Payloads should be stored as-is (no execution, just storage)
	for i, captured := range history {
		assert.Equal(t, injectionPayloads[i], captured.Path,
			"Payload should be stored verbatim without interpretation")
	}
}

// ===== API KEY BRUTE FORCE =====

func TestSecurity_BruteForceKeyAttempts(t *testing.T) {
	authManager := authmgr.NewAuthManager(nil)
	realKey := tests.GenerateTestAPIKey()
	authManager.AddKey(realKey, "real-user", 100)

	// Attempt 1000 invalid keys
	failCount := 0
	for i := 0; i < 1000; i++ {
		fakeKey := tests.GenerateTestAPIKey()
		_, err := authManager.ValidateKey(fakeKey)
		if err != nil {
			failCount++
		}
	}

	assert.Equal(t, 1000, failCount, "All brute force attempts should fail")

	// Real key should still work after brute force
	_, err := authManager.ValidateKey(realKey)
	assert.NoError(t, err, "Real key should still be valid after brute force attempts")
}

// ===== LARGE PAYLOAD HANDLING =====

func TestSecurity_LargePayloadInInspector(t *testing.T) {
	ti := server.NewTrafficInspector(10)

	// Create a very large body
	largeBody := []byte(strings.Repeat("A", 10*1024*1024)) // 10MB

	ti.Record(&server.CapturedRequest{
		ID:      "large-req",
		Method:  "POST",
		Path:    "/api/upload",
		ReqBody: largeBody,
	})

	found := ti.GetByID("large-req")
	assert.NotNil(t, found)
	assert.Len(t, found.ReqBody, 10*1024*1024, "Large payload should be stored completely")
}

// ===== AUTH CONCURRENT SAFETY =====

func TestSecurity_ConcurrentAuthValidation(t *testing.T) {
	authManager := authmgr.NewAuthManager(nil)
	validKey := tests.GenerateTestAPIKey()
	authManager.AddKey(validKey, "user-1", 100)

	done := make(chan bool, 200)

	// 100 valid attempts + 100 invalid attempts concurrently
	for i := 0; i < 100; i++ {
		go func() {
			_, err := authManager.ValidateKey(validKey)
			done <- (err == nil)
		}()
		go func() {
			_, err := authManager.ValidateKey(tests.GenerateTestAPIKey())
			done <- (err != nil)
		}()
	}

	correctResults := 0
	for i := 0; i < 200; i++ {
		if <-done {
			correctResults++
		}
	}

	assert.Equal(t, 200, correctResults, "All concurrent validations should return correct results")
}

// ===== BENCHMARK =====

func BenchmarkSecurity_AuthValidation(b *testing.B) {
	authManager := authmgr.NewAuthManager(nil)
	validKey := tests.GenerateTestAPIKey()
	authManager.AddKey(validKey, "bench-user", 100)

	b.Run("ValidKey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			authManager.ValidateKey(validKey)
		}
	})

	b.Run("InvalidKey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			authManager.ValidateKey("invalid-key-attempt")
		}
	})
}

func BenchmarkSecurity_RateLimiterDecision(b *testing.B) {
	rl := limiter.NewRateLimiter(nil, 100, time.Minute)

	b.Run("AllowedUser", func(b *testing.B) {
		rl.SetUserTier("premium", "premium")
		for i := 0; i < b.N; i++ {
			rl.Allow("premium", 1)
		}
	})

	b.Run("BlockedUser", func(b *testing.B) {
		// Pre-exhaust
		for i := 0; i < 100; i++ {
			rl.Allow("blocked-user", 1)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rl.Allow("blocked-user", 1)
		}
	})
}
