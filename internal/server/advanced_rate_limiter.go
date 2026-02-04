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
