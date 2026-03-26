package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type GeoLocation struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
}

type geoEntry struct {
	loc       *GeoLocation
	expiresAt time.Time
}

type GeoLocator struct {
	//cache
	cache    map[string]*geoEntry
	cacheMu  sync.RWMutex
	maxCache int

	//rate limiting
	lastCall time.Time
	callMu   sync.Mutex

	//Api ayarlari
	apiUrl   string
	useCache bool
}

// yeni geolocator oluşturuyoruz
func NewGeoLocator(useCache bool) *GeoLocator {
	gl := &GeoLocator{
		cache:    make(map[string]*geoEntry),
		maxCache: 10000, // max 10k IP cache
		useCache: useCache,
		apiUrl:   "http://ip-api.com/json/",
	}
	// Background cache eviction every 10 minutes
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			gl.evictExpired()
		}
	}()
	return gl
}

func (g *GeoLocator) evictExpired() {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()
	now := time.Now()
	for ip, entry := range g.cache {
		if now.After(entry.expiresAt) {
			delete(g.cache, ip)
		}
	}
}

// Lookup-ip konumu bulma
func (g *GeoLocator) Lookup(ip string) (*GeoLocation, error) {
	if isLocalIP(ip) {
		return &GeoLocation{
			Country:     "Local",
			CountryCode: "L0",
			City:        "Localhost",
		}, nil
	}
	if g.useCache {
		if loc := g.getFromCache(ip); loc != nil {
			return loc, nil
		}
	}
	g.callMu.Lock()
	if time.Since(g.lastCall) < 4*time.Second {
		g.callMu.Unlock()

		return &GeoLocation{Country: "Unknown"}, nil
	}
	g.lastCall = time.Now()
	g.callMu.Unlock()

	loc, err := g.fetchFromAPI(ip)

	if err != nil {
		return nil, err
	}
	if g.useCache {
		g.saveToCache(ip, loc)
	}
	return loc, nil
}

func (g *GeoLocator) fetchFromAPI(ip string) (*GeoLocation, error) {
	resp, err := http.Get(g.apiUrl + ip)

	if err != nil {
		return nil, fmt.Errorf("API isteği başarısız oldu: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		City        string  `json:"city"`
		Region      string  `json:"region"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
		ISP         string  `json:"isp"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("api cevabı parse edilemedi :%w", err)
	}

	if apiResp.Status != "success" {
		return nil, fmt.Errorf("api hatası : %s", apiResp.Status)
	}

	return &GeoLocation{
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		City:        apiResp.City,
		Region:      apiResp.Region,
		Latitude:    apiResp.Lat,
		Longitude:   apiResp.Lon,
		Timezone:    apiResp.Timezone,
		ISP:         apiResp.ISP,
	}, nil
}

// get from cache (TTL-aware)
func (g *GeoLocator) getFromCache(ip string) *GeoLocation {
	g.cacheMu.RLock()
	defer g.cacheMu.RUnlock()

	entry, ok := g.cache[ip]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil
	}
	return entry.loc
}

// cache kaydet (with 24h TTL, evict oldest if at capacity)
func (g *GeoLocator) saveToCache(ip string, loc *GeoLocation) {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()

	// If at capacity, evict ~10% oldest entries
	if len(g.cache) >= g.maxCache {
		count := 0
		for k := range g.cache {
			delete(g.cache, k)
			count++
			if count >= g.maxCache/10 {
				break
			}
		}
	}

	g.cache[ip] = &geoEntry{
		loc:       loc,
		expiresAt: time.Now().Add(24 * time.Hour),
	}
}

// istatistik
func (g *GeoLocator) CacheStats() map[string]interface{} {
	g.cacheMu.RLock()
	defer g.cacheMu.RUnlock()

	return map[string]interface{}{
		"cache_size": len(g.cache),
		"enabled":    g.useCache,
	}
}

// local ip kontrolü
func isLocalIP(ip string) bool {
	localIPs := []string{"127.0.0.1", "::1", "localhost"}
	for _, local := range localIPs {
		if ip == local {
			return true
		}
	}
	return false
}
