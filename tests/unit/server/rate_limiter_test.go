package server_test

import (
	"sync"
	"testing"

	"github.com/Bekican/gorenel/internal/server"
	"github.com/stretchr/testify/assert"
)

// ===== FREE TIER TESTS =====

func TestRateLimiter_FreeTierAllows(t *testing.T) {
	rl := server.NewAdvancedRateLmiter()

	// First request should always pass
	assert.True(t, rl.Allow("user-1"), "First request should be allowed")
}

func TestRateLimiter_FreeTierLimit(t *testing.T) {
	rl := server.NewAdvancedRateLmiter()

	allowed := 0
	for i := 0; i < 150; i++ {
		if rl.Allow("free-user") {
			allowed++
		}
	}

	// Free tier: 100 req/hour — first request creates the log entry
	assert.LessOrEqual(t, allowed, 100, "Free tier should not exceed 100 requests")
	assert.Greater(t, allowed, 0, "At least some requests should be allowed")
}

// ===== PREMIUM TIER TESTS =====

func TestRateLimiter_PremiumTierHigherLimit(t *testing.T) {
	rl := server.NewAdvancedRateLmiter()
	rl.SetUserTier("premium-user", "premium")

	allowed := 0
	for i := 0; i < 200; i++ {
		if rl.Allow("premium-user") {
			allowed++
		}
	}

	// Premium users should have much more capacity than free tier
	assert.Equal(t, 200, allowed, "Premium user should allow all 200 requests (limit is 10000)")
}

// ===== TIER UPGRADE =====

func TestRateLimiter_TierUpgrade(t *testing.T) {
	rl := server.NewAdvancedRateLmiter()

	// Exhaust free tier
	for i := 0; i < 100; i++ {
		rl.Allow("upgrade-user")
	}

	// Free tier should be exhausted
	blocked := !rl.Allow("upgrade-user")
	assert.True(t, blocked, "Should be blocked after 100 free requests")

	// Upgrade to premium — requests should be allowed again
	rl.SetUserTier("upgrade-user", "premium")
	assert.True(t, rl.Allow("upgrade-user"), "After upgrade to premium, should be allowed")
}

// ===== INDEPENDENT USERS =====

func TestRateLimiter_IndependentUserLimits(t *testing.T) {
	rl := server.NewAdvancedRateLmiter()

	// Exhaust user-A
	for i := 0; i < 100; i++ {
		rl.Allow("user-A")
	}

	// user-B should still be allowed (independent)
	assert.True(t, rl.Allow("user-B"), "User B should not be affected by User A's limit")
}

// ===== CONCURRENCY =====

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := server.NewAdvancedRateLmiter()
	var wg sync.WaitGroup

	results := make([]bool, 200)
	wg.Add(200)
	for i := 0; i < 200; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = rl.Allow("concurrent-user")
		}(i)
	}
	wg.Wait()

	allowed := 0
	for _, r := range results {
		if r {
			allowed++
		}
	}

	// Under concurrent access, should still respect limits (100 for free tier)
	assert.LessOrEqual(t, allowed, 101, "Concurrent access should respect rate limits")
	assert.Greater(t, allowed, 0, "Some requests must be allowed")
}

// ===== BENCHMARKS =====

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := server.NewAdvancedRateLmiter()
	rl.SetUserTier("bench-user", "premium")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow("bench-user")
	}
}

func BenchmarkRateLimiter_Allow_MultiUser(b *testing.B) {
	rl := server.NewAdvancedRateLmiter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow("user-" + string(rune('A'+i%26)))
	}
}
