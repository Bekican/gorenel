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
}

func NewRateLimiter(rdb *redis.Client, defaultLimit int, defaultWindow time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rdb:      rdb,
		tiers:    make(map[string]Quota),
		userTier: make(map[string]string),
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
		"active_users": -1, // Not tracked in Redis optimally
		"tiers_count":  len(rl.tiers),
	}
}
