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

// time-series data point
type DataPoint struct {
	Timestamp    time.Time
	RequestCount int64
	BytesIn      int64
	BytesOut     int64
	AvgLagency   time.Duration
}

// yeni analytics engine
func NewAnalyticsEngine(window time.Duration) *AnalyticsEngine {
	ae := &AnalyticsEngine{
		window:        window,
		dataPoints:    make([]DataPoint, 0),
		topPaths:      make(map[string]int64),
		topCountries:  make(map[string]int64),
		topUserAgents: make(map[string]int64),
		statusCodes:   make(map[int]int64),
	}
	go ae.Cleanup()

	return ae
}
