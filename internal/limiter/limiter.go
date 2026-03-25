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
		return rl.allowInMemory(identifier, n, quota)
	}

	ctx := context.Background()
	now := time.Now()
	// Using a sliding window with Redis ZSETs
	windowKey := fmt.Sprintf("rate_limit:%s", identifier)

	// Remove older elements
	minScore := fmt.Sprintf("%d", now.Add(-quota.WindowSize).UnixNano())
	if err := rl.rdb.ZRemRangeByScore(ctx, windowKey, "-inf", minScore).Err(); err != nil {
		return rl.allowInMemory(identifier, n, quota)
	}

	// Count elements
	count, err := rl.rdb.ZCard(ctx, windowKey).Result()
	if err != nil {
		return rl.allowInMemory(identifier, n, quota)
	}
	if count+int64(n) > int64(quota.Limit) {
		return false
	}

	// Add new requests only after quota check
	members := make([]redis.Z, 0, n)
	for i := 0; i < n; i++ {
		member := fmt.Sprintf("%d-%d", now.UnixNano(), i)
		members = append(members, redis.Z{
			Score:  float64(now.UnixNano()),
			Member: member,
		})
	}
	if err := rl.rdb.ZAdd(ctx, windowKey, members...).Err(); err != nil {
		return rl.allowInMemory(identifier, n, quota)
	}

	// Set expiry on the whole set to clean up automatically
	if err := rl.rdb.Expire(ctx, windowKey, quota.WindowSize).Err(); err != nil {
		return rl.allowInMemory(identifier, n, quota)
	}

	return true
}

func (rl *RateLimiter) AllowWithQuota(identifier string, n int, quota Quota) bool {
	// If redis is not available, use memory store (fallback/testing)
	if rl.rdb == nil {
		return rl.allowInMemory(identifier, n, quota)
	}

	ctx := context.Background()
	now := time.Now()
	windowKey := fmt.Sprintf("rate_limit:%s", identifier)

	minScore := fmt.Sprintf("%d", now.Add(-quota.WindowSize).UnixNano())
	if err := rl.rdb.ZRemRangeByScore(ctx, windowKey, "-inf", minScore).Err(); err != nil {
		return rl.allowInMemory(identifier, n, quota)
	}

	count, err := rl.rdb.ZCard(ctx, windowKey).Result()
	if err != nil {
		return rl.allowInMemory(identifier, n, quota)
	}
	if count+int64(n) > int64(quota.Limit) {
		return false
	}

	members := make([]redis.Z, 0, n)
	for i := 0; i < n; i++ {
		member := fmt.Sprintf("%d-%d", now.UnixNano(), i)
		members = append(members, redis.Z{
			Score:  float64(now.UnixNano()),
			Member: member,
		})
	}
	if err := rl.rdb.ZAdd(ctx, windowKey, members...).Err(); err != nil {
		return rl.allowInMemory(identifier, n, quota)
	}
	if err := rl.rdb.Expire(ctx, windowKey, quota.WindowSize).Err(); err != nil {
		return rl.allowInMemory(identifier, n, quota)
	}
	return true
}

func (rl *RateLimiter) allowInMemory(identifier string, n int, quota Quota) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	expiry := now.Add(-quota.WindowSize)

	var valid []time.Time
	for _, t := range rl.memStore[identifier] {
		if t.After(expiry) {
			valid = append(valid, t)
		}
	}

	if len(valid)+n > quota.Limit {
		rl.memStore[identifier] = valid
		return false
	}

	for i := 0; i < n; i++ {
		valid = append(valid, now)
	}
	rl.memStore[identifier] = valid
	return true
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
