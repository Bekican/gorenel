package server

import (
	"sync"
	"time"
)

type Quota struct {
	Limit      int
	WindowSize time.Duration
}

type UsageLog struct {
	TimeStamps []time.Time
}

type AdvancedRateLimiter struct {
	mu       sync.RWMutex
	usage    map[string]*UsageLog
	tiers    map[string]Quota
	userTier map[string]string
}

func NewAdvancedRateLmiter() *AdvancedRateLimiter {
	rl := &AdvancedRateLimiter{
		usage:    make(map[string]*UsageLog),
		tiers:    make(map[string]Quota),
		userTier: make(map[string]string),
	}

	rl.tiers["free"] = Quota{Limit: 100, WindowSize: 1 * time.Hour}
	rl.tiers["premium"] = Quota{Limit: 10000, WindowSize: 24 * time.Hour}

	return rl
}

func (rl *AdvancedRateLimiter) Allow(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	tierID := rl.userTier[userID]
	if tierID == "" {
		tierID = "free"
	}
	quota := rl.tiers[tierID]

	log, exists := rl.usage[userID]

	if !exists {
		log = &UsageLog{TimeStamps: []time.Time{time.Now()}}
		rl.usage[userID] = log
		return true
	}

	now := time.Now()
	cutoff := now.Add(-quota.WindowSize)

	validRequests := []time.Time{}
	for _, t := range log.TimeStamps {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	log.TimeStamps = validRequests

	if len(log.TimeStamps) >= quota.Limit {
		return false
	}

	log.TimeStamps = append(log.TimeStamps, now)
	return true
}

func (rl *AdvancedRateLimiter) SetUserTier(userID, tierID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.userTier[userID] = tierID
}
