// package server

// import (
// 	"sync"
// 	"time"
// 	"sort"
// )

// // real time analytics
// type AnalyticsEngine struct {
// 	window     time.Duration
// 	dataPoints []DataPoint
// 	dataMu     sync.RWMutex

// 	topPaths      map[string]int64
// 	topCountries  map[string]int64
// 	topUserAgents map[string]int64
// 	statusCodes   map[int]int64
// 	statsMu       sync.RWMutex

// 	responseTimeSum   time.Duration
// 	responseTimeCount int64
// 	perfMu            sync.RWMutex
// }

// // time-series data point
// type DataPoint struct {
// 	Timestamp    time.Time
// 	RequestCount int64
// 	BytesIn      int64
// 	BytesOut     int64
// 	AvgLatency   time.Duration
// }

// // yeni analytics engine
// func NewAnalyticsEngine(window time.Duration) *AnalyticsEngine {
// 	ae := &AnalyticsEngine{
// 		window:        window,
// 		dataPoints:    make([]DataPoint, 0),
// 		topPaths:      make(map[string]int64),
// 		topCountries:  make(map[string]int64),
// 		topUserAgents: make(map[string]int64),
// 		statusCodes:   make(map[int]int64),
// 	}
// 	go ae.Cleanup()

// 	return ae
// }

// //event-consumer analizler için event işleme
// func(ae *AnalyticsEngine) Consume(event *RequestEvent) error{
// 	//ortalama istatistikler
// 	ae.statsMu.Lock()
// 	ae.topPaths[event.Path]++
// 	ae.topCountries[event.GeoCountry]++
// 	ae.topUserAgents[event.UserAgent]++
// 	ae.statusCodes[event.StatusCode]++
// 	ae.statsMu.Unlock()

// 	//performans metrikleri
// 	ae.perfMu.Lock()
// 	ae.responseTimeSum += event.ResponseTime
// 	ae.responseTimeCount++
// 	ae.perfMu.Unlock()

// 	//time series data point
// 	ae.dataMu.Lock()
// 	ae.dataPoints = append(ae.dataPoints, DataPoint{
// 		Timestamp: event.Timestamp,
// 		RequestCount: 1,
// 		BytesIn: event.BytesReceived,
// 		BytesOut: event.ByteSent,
// 		AvgLatency: event.ResponseTime,
// 	})
// 	ae.dataMu.Unlock()

// 	return nil
// }

// func(ae *AnalyticsEngine) Name() string{
// 	return "AnalyticsEngine"
// }

// //mevcut analizin snapshotu'u
// func(ae *AnalyticsEngine) GetSnapshot() *AnalyticsSnapshot{
// 	snapshot := &AnalyticsEngine{
// 		Timestamp : time.Now(),
// 	}

// 	ae.statsMu.RLock()
// 	snapshot.TopPaths = getTopN(ae.topPaths,10)
// 	snapshot.TopCountries = getTopN(ae.topCountries,10)
// 	snapshot.TopUserAgents = getTopN(ae.topUserAgents,5)
// 	snapshot.StatusCodeDist = copyMap(ae.statusCodes)
// 	ae.statsMu.RUnlock()

// 	ae.perfMu.RLock()
// 	if ae.responseTimeCount > 0{
// 		snapshot.AvgResponseTime = ae.responseTimeSum / time.Duration(ae.responseTimeCount)
// 	}
// 	snapshot.TotalRequests = ae.responseTimeCount
// 	ae.perfMu.RUnlock()

// 	ae.dataMu.RLock()
// 	snapshot.TimeSeries = ae.aggregateTimeSeries(time.Hour)
// 	ae.dataMu.RUnlock()

// 	return snapshot
// }

// func (ae *AnalyticsEngine) aggregateTimeSeries(duration time.Duration) []TimeSeriesBucket {
// 	now := time.Now()
// 	cutoff := now.Add(-duration)

// 	buckets := make(map[string]*TimeSeriesBucket)

// 	for _, dp := range ae.dataPoints {
// 		if dp.Timestamp.Before(cutoff) {
// 			continue
// 		}

// 		// 1 dakikalık bucket'lara böl
// 		bucketKey := dp.Timestamp.Truncate(time.Minute).Format(time.RFC3339)

// 		if _, exists := buckets[bucketKey]; !exists {
// 			buckets[bucketKey] = &TimeSeriesBucket{
// 				Timestamp: dp.Timestamp.Truncate(time.Minute),
// 			}
// 		}

// 		buckets[bucketKey].Requests++
// 		buckets[bucketKey].BytesIn += dp.BytesIn
// 		buckets[bucketKey].BytesOut += dp.BytesOut
// 		buckets[bucketKey].TotalLatency += dp.AvgLatency
// 	}

// 	// Map'i slice'a çevir ve sırala
// 	result := make([]TimeSeriesBucket, 0, len(buckets))
// 	for _, bucket := range buckets {
// 		if bucket.Requests > 0 {
// 			bucket.AvgLatency = bucket.TotalLatency / time.Duration(bucket.Requests)
// 		}
// 		result = append(result, *bucket)
// 	}

// 	sort.Slice(result, func(i, j int) bool {
// 		return result[i].Timestamp.Before(result[j].Timestamp)
// 	})

// 	return result
// }

// func (ae *AnalyticsEngine) cleanup() {
// 	ticker := time.NewTicker(5 * time.Minute)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		ae.dataMu.Lock()
// 		cutoff := time.Now().Add(-ae.window)

// 		// Window dışındaki point'leri sil
// 		newPoints := make([]DataPoint, 0, len(ae.dataPoints))
// 		for _, dp := range ae.dataPoints {
// 			if dp.Timestamp.After(cutoff) {
// 				newPoints = append(newPoints, dp)
// 			}
// 		}
// 		ae.dataPoints = newPoints
// 		ae.dataMu.Unlock()
// 	}
// }

// type AnalyticsSnapshot struct {
// 	Timestamp       time.Time                  `json:"timestamp"`
// 	TotalRequests   int64                      `json:"total_requests"`
// 	AvgResponseTime time.Duration              `json:"avg_response_time_ms"`
// 	TopPaths        []RankItem                 `json:"top_paths"`
// 	TopCountries    []RankItem                 `json:"top_countries"`
// 	TopUserAgents   []RankItem                 `json:"top_user_agents"`
// 	StatusCodeDist  map[int]int64              `json:"status_code_distribution"`
// 	TimeSeries      []TimeSeriesBucket         `json:"time_series"`
// }

// // RankItem - Top-N item
// type RankItem struct {
// 	Key   string `json:"key"`
// 	Count int64  `json:"count"`
// }

// // TimeSeriesBucket - Zaman serisi bucket (1 dakikalık)
// type TimeSeriesBucket struct {
// 	Timestamp    time.Time     `json:"timestamp"`
// 	Requests     int64         `json:"requests"`
// 	BytesIn      int64         `json:"bytes_in"`
// 	BytesOut     int64         `json:"bytes_out"`
// 	AvgLatency   time.Duration `json:"avg_latency_ms"`
// 	TotalLatency time.Duration `json:"-"` // Internal use only
// }

// // --- HELPER FUNCTIONS ---

// func getTopN(m map[string]int64, n int) []RankItem {
// 	items := make([]RankItem, 0, len(m))
// 	for k, v := range m {
// 		items = append(items, RankItem{Key: k, Count: v})
// 	}

// 	sort.Slice(items, func(i, j int) bool {
// 		return items[i].Count > items[j].Count
// 	})

// 	if len(items) > n {
// 		items = items[:n]
// 	}

// 	return items
// }

// func copyMap(m map[int]int64) map[int]int64 {
// 	result := make(map[int]int64, len(m))
// 	for k, v := range m {
// 		result[k] = v
// 	}
// 	return result
// }

package server

import (
	"sort"
	"sync"
	"time"
)

// AnalyticsEngine - Real-time traffic analytics
type AnalyticsEngine struct {
	// Time-series data (sliding window)
	window     time.Duration
	dataPoints []DataPoint
	dataMu     sync.RWMutex

	// Aggregated stats
	topPaths      map[string]int64
	topCountries  map[string]int64
	topUserAgents map[string]int64
	statusCodes   map[int]int64
	statsMu       sync.RWMutex

	// Performance metrics
	responseTimeSum   time.Duration
	responseTimeCount int64
	perfMu            sync.RWMutex
}

// DataPoint - Time-series veri noktası
type DataPoint struct {
	Timestamp    time.Time
	RequestCount int64
	BytesIn      int64
	BytesOut     int64
	AvgLatency   time.Duration
}

// NewAnalyticsEngine - Yeni analytics engine
func NewAnalyticsEngine(window time.Duration) *AnalyticsEngine {
	ae := &AnalyticsEngine{
		window:        window,
		dataPoints:    make([]DataPoint, 0),
		topPaths:      make(map[string]int64),
		topCountries:  make(map[string]int64),
		topUserAgents: make(map[string]int64),
		statusCodes:   make(map[int]int64),
	}

	// Cleanup goroutine (eski data point'leri sil)
	go ae.cleanup()

	return ae
}

// Consume - EventConsumer interface (analytics için event işle)
func (ae *AnalyticsEngine) Consume(event *RequestEvent) error {
	// Aggregated stats güncelle
	ae.statsMu.Lock()
	ae.topPaths[event.Path]++
	ae.topCountries[event.GeoCountry]++
	ae.topUserAgents[event.UserAgent]++
	ae.statusCodes[event.StatusCode]++
	ae.statsMu.Unlock()

	// Performance metrics
	ae.perfMu.Lock()
	ae.responseTimeSum += event.ResponseTime
	ae.responseTimeCount++
	ae.perfMu.Unlock()

	// Time-series data point ekle
	ae.dataMu.Lock()
	ae.dataPoints = append(ae.dataPoints, DataPoint{
		Timestamp:    event.Timestamp,
		RequestCount: 1,
		BytesIn:      event.BytesReceived,
		BytesOut:     event.BytesSent,
		AvgLatency:   event.ResponseTime,
	})
	ae.dataMu.Unlock()

	return nil
}

// Name - Consumer adı
func (ae *AnalyticsEngine) Name() string {
	return "AnalyticsEngine"
}

// GetSnapshot - Mevcut analytics snapshot'ı
func (ae *AnalyticsEngine) GetSnapshot() *AnalyticsSnapshot {
	snapshot := &AnalyticsSnapshot{
		Timestamp: time.Now(),
	}

	// Top paths
	ae.statsMu.RLock()
	snapshot.TopPaths = getTopN(ae.topPaths, 10)
	snapshot.TopCountries = getTopN(ae.topCountries, 10)
	snapshot.TopUserAgents = getTopN(ae.topUserAgents, 5)
	snapshot.StatusCodeDist = copyMap(ae.statusCodes)
	ae.statsMu.RUnlock()

	// Performance metrics
	ae.perfMu.RLock()
	if ae.responseTimeCount > 0 {
		snapshot.AvgResponseTime = ae.responseTimeSum / time.Duration(ae.responseTimeCount)
	}
	snapshot.TotalRequests = ae.responseTimeCount
	ae.perfMu.RUnlock()

	// Time-series (son 1 saat)
	ae.dataMu.RLock()
	snapshot.TimeSeries = ae.aggregateTimeSeries(time.Hour)
	ae.dataMu.RUnlock()

	return snapshot
}

// aggregateTimeSeries - Time-series data'yı aggregate et (1 dakikalık bucket'lar)
func (ae *AnalyticsEngine) aggregateTimeSeries(duration time.Duration) []TimeSeriesBucket {
	now := time.Now()
	cutoff := now.Add(-duration)

	buckets := make(map[string]*TimeSeriesBucket)

	for _, dp := range ae.dataPoints {
		if dp.Timestamp.Before(cutoff) {
			continue
		}

		// 1 dakikalık bucket'lara böl
		bucketKey := dp.Timestamp.Truncate(time.Minute).Format(time.RFC3339)

		if _, exists := buckets[bucketKey]; !exists {
			buckets[bucketKey] = &TimeSeriesBucket{
				Timestamp: dp.Timestamp.Truncate(time.Minute),
			}
		}

		buckets[bucketKey].Requests++
		buckets[bucketKey].BytesIn += dp.BytesIn
		buckets[bucketKey].BytesOut += dp.BytesOut
		buckets[bucketKey].TotalLatency += dp.AvgLatency
	}

	// Map'i slice'a çevir ve sırala
	result := make([]TimeSeriesBucket, 0, len(buckets))
	for _, bucket := range buckets {
		if bucket.Requests > 0 {
			bucket.AvgLatency = bucket.TotalLatency / time.Duration(bucket.Requests)
		}
		result = append(result, *bucket)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result
}

// cleanup - Eski data point'leri temizle
func (ae *AnalyticsEngine) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ae.dataMu.Lock()
		cutoff := time.Now().Add(-ae.window)

		// Window dışındaki point'leri sil
		newPoints := make([]DataPoint, 0, len(ae.dataPoints))
		for _, dp := range ae.dataPoints {
			if dp.Timestamp.After(cutoff) {
				newPoints = append(newPoints, dp)
			}
		}
		ae.dataPoints = newPoints
		ae.dataMu.Unlock()
	}
}

// --- SNAPSHOT TYPES ---

// AnalyticsSnapshot - Belirli bir andaki analytics durumu
type AnalyticsSnapshot struct {
	Timestamp       time.Time          `json:"timestamp"`
	TotalRequests   int64              `json:"total_requests"`
	AvgResponseTime time.Duration      `json:"avg_response_time_ms"`
	TopPaths        []RankItem         `json:"top_paths"`
	TopCountries    []RankItem         `json:"top_countries"`
	TopUserAgents   []RankItem         `json:"top_user_agents"`
	StatusCodeDist  map[int]int64      `json:"status_code_distribution"`
	TimeSeries      []TimeSeriesBucket `json:"time_series"`
}

// RankItem - Top-N item
type RankItem struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// TimeSeriesBucket - Zaman serisi bucket (1 dakikalık)
type TimeSeriesBucket struct {
	Timestamp    time.Time     `json:"timestamp"`
	Requests     int64         `json:"requests"`
	BytesIn      int64         `json:"bytes_in"`
	BytesOut     int64         `json:"bytes_out"`
	AvgLatency   time.Duration `json:"avg_latency_ms"`
	TotalLatency time.Duration `json:"-"` // Internal use only
}

// --- HELPER FUNCTIONS ---

func getTopN(m map[string]int64, n int) []RankItem {
	items := make([]RankItem, 0, len(m))
	for k, v := range m {
		items = append(items, RankItem{Key: k, Count: v})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})

	if len(items) > n {
		items = items[:n]
	}

	return items
}

func copyMap(m map[int]int64) map[int]int64 {
	result := make(map[int]int64, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
