package limiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Quota defines the limits for a specific tier
type Quota struct {
	Limit      int           // Max requests allowed
	WindowSize time.Duration // Time window (e.g., 1m, 1h, 24h)
}

type RateLimiter struct {
	rdb      *redis.Client
	tiers    map[string]Quota
	userTier map[string]string // UserID/IP -> PlanID
	mu       sync.RWMutex
	
	// Memory store fallback for testing (if rdb is nil)
	memStore map[string][]time.Time
}

func NewRateLimiter(rdb *redis.Client, defaultLimit int, defaultWindow time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rdb:      rdb,
		tiers:    make(map[string]Quota),
		userTier: make(map[string]string),
		memStore: make(map[string][]time.Time),
	}

	// Default Tiers
	rl.tiers["free"] = Quota{Limit: defaultLimit, WindowSize: defaultWindow}
	rl.tiers["premium"] = Quota{Limit: 1000, WindowSize: 1 * time.Minute}
	rl.tiers["enterprise"] = Quota{Limit: 10000, WindowSize: 24 * time.Hour}

	return rl
}

func (rl *RateLimiter) Allow(identifier string, n int) bool {
	rl.mu.RLock()
	tierID := rl.userTier[identifier]
	if tierID == "" {
		tierID = "free"
	}
	quota := rl.tiers[tierID]
	rl.mu.RUnlock()

	// If redis is not available, use memory store (fallback/testing)
	if rl.rdb == nil {
		rl.mu.Lock()
		defer rl.mu.Unlock()
		
		now := time.Now()
		expiry := now.Add(-quota.WindowSize)
		
		// Clean up old entries
		var valid []time.Time
		for _, t := range rl.memStore[identifier] {
			if t.After(expiry) {
				valid = append(valid, t)
			}
		}
		
		// Check limit
		if len(valid)+n > quota.Limit {
			rl.memStore[identifier] = valid
			return false
		}
		
		// Add new entries
		for i := 0; i < n; i++ {
			valid = append(valid, now)
		}
		rl.memStore[identifier] = valid
		return true
	}

	ctx := context.Background()
	now := time.Now()
	// Using a sliding window with Redis ZSETs
	windowKey := fmt.Sprintf("rate_limit:%s", identifier)

	// Remove older elements
	minScore := fmt.Sprintf("%d", now.Add(-quota.WindowSize).UnixNano())
	rl.rdb.ZRemRangeByScore(ctx, windowKey, "-inf", minScore)

	// Add new requests
	for i := 0; i < n; i++ {
		member := fmt.Sprintf("%d-%d", now.UnixNano(), i)
		rl.rdb.ZAdd(ctx, windowKey, redis.Z{
			Score:  float64(now.UnixNano()),
			Member: member,
		})
	}

	// Set expiry on the whole set to clean up automatically
	rl.rdb.Expire(ctx, windowKey, quota.WindowSize)

	// Count elements
	count, err := rl.rdb.ZCard(ctx, windowKey).Result()
	if err != nil {
		// Log error, fallback allow
		return true
	}

	return count <= int64(quota.Limit)
}

func (rl *RateLimiter) SetUserTier(identifier, tierID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.userTier[identifier] = tierID
}

func (rl *RateLimiter) Stats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"active_users": len(rl.memStore),
		"tiers_count":  len(rl.tiers),
	}
}
