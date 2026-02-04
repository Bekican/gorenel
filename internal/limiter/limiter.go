package limiter

import (
	"sync"
	"time"
)

// Quota defines the limits for a specific tier
type Quota struct {
	Limit      int           // Max requests allowed
	WindowSize time.Duration // Time window (e.g., 1m, 1h, 24h)
}

// UsageLog tracks request timestamps for a user/identifier
type UsageLog struct {
	Timestamps []time.Time
	LastSeen   time.Time
}

type RateLimiter struct {
	mu       sync.RWMutex
	usage    map[string]*UsageLog
	tiers    map[string]Quota  // PlanID -> Quota settings
	userTier map[string]string // UserID/IP -> PlanID
}

func NewRateLimiter(defaultLimit int, defaultWindow time.Duration) *RateLimiter {
	rl := &RateLimiter{
		usage:    make(map[string]*UsageLog),
		tiers:    make(map[string]Quota),
		userTier: make(map[string]string),
	}

	// Default Tiers
	rl.tiers["free"] = Quota{Limit: defaultLimit, WindowSize: defaultWindow}
	rl.tiers["premium"] = Quota{Limit: 1000, WindowSize: 1 * time.Minute}
	rl.tiers["enterprise"] = Quota{Limit: 10000, WindowSize: 24 * time.Hour}

	// Start cleanup routine
	go rl.cleanup()

	return rl
}

// Allow checks if the request identified by 'identifier' should be permitted
func (rl *RateLimiter) Allow(identifier string, n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Get user tier (default to free)
	tierID := rl.userTier[identifier]
	if tierID == "" {
		tierID = "free"
	}
	quota := rl.tiers[tierID]

	now := time.Now()
	log, exists := rl.usage[identifier]
	if !exists {
		log = &UsageLog{
			Timestamps: make([]time.Time, 0, n),
			LastSeen:   now,
		}
		rl.usage[identifier] = log
	}

	// 1. Sliding Window Filtering
	cutoff := now.Add(-quota.WindowSize)
	validIndices := 0
	for i, t := range log.Timestamps {
		if t.After(cutoff) {
			validIndices = i
			break
		} else {
			validIndices = i + 1
		}
	}

	if validIndices > 0 {
		log.Timestamps = log.Timestamps[validIndices:]
	}

	// 2. Capacity Check
	if len(log.Timestamps)+n > quota.Limit {
		return false
	}

	// 3. Record new requests
	for i := 0; i < n; i++ {
		log.Timestamps = append(log.Timestamps, now)
	}
	log.LastSeen = now
	return true
}

func (rl *RateLimiter) SetUserTier(identifier, tierID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.userTier[identifier] = tierID
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for id, log := range rl.usage {
			if now.Sub(log.LastSeen) > 1*time.Hour {
				delete(rl.usage, id)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Stats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"active_users": len(rl.usage),
		"tiers_count":  len(rl.tiers),
	}
}
