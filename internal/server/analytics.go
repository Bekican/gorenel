package server

import (
	"sync"
	"time"
)

// real time analytics
type AnalyticsEngine struct {
	window     time.Duration
	dataPoints []DataPoint
	dataMu     sync.RWMutex

	topPaths      map[string]int64
	topCountries  map[string]int64
	topUserAgents map[string]int64
	statusCodes   map[int]int64
	statsMu       sync.RWMutex

	responseTimeSum   time.Duration
	responseTimeCount int64
	perfMu            sync.RWMutex
}
