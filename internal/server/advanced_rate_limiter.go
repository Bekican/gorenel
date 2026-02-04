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
