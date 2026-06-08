package server

import (
	"container/list"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
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

// --- ÖĞRETİCİ ADIM 1: LRU Düğüm Yapısı (Node Structure) ---
// Bellekte tuttuğumuz her IP kaydını, çift yönlü bağlı listede (doubly linked list)
// yer alan bir düğüme (element) bağlıyoruz. Bu sayede düğüm üzerinden O(1) sürede
// silme ve güncelleme işlemleri yapabiliyoruz.
type geoEntry struct {
	ip        string
	loc       *GeoLocation
	expiresAt time.Time
	element   *list.Element // container/list içindeki düğümün pointer'ı
}

// --- ÖĞRETİCİ ADIM 2: Gerçek LRU (Least Recently Used) Önbelleği ---
// Rastgele silme (Random Eviction) yerine, gerçek bir LRU mimarisi tasarladık.
// En son sorgulanan IP'ler listenin başına (Front) taşınır.
// Önbellek dolduğunda ise en eski/en az kullanılan IP'ler listenin sonundan (Back) O(1) sürede çıkarılır.
type GeoLocator struct {
	// cache yapısı
	cache     map[string]*geoEntry
	evictList *list.List // LRU sırasını takip eden çift yönlü bağlı liste (Front = En yeni, Back = En eski)
	cacheMu   sync.Mutex // LRU güncellemelerinde listeyi korumak için standart Mutex (deadlock önler)
	maxCache  int

	// rate limiting
	lastCall time.Time
	callMu   sync.Mutex

	// Api ayarlari
	apiUrl   string
	useCache bool

	// Goroutine leak sızıntısını önlemek için yaşam döngüsü kanalı
	stopChan chan struct{}
}

// yeni geolocator oluşturuyoruz
func NewGeoLocator(useCache bool) *GeoLocator {
	gl := &GeoLocator{
		cache:     make(map[string]*geoEntry),
		evictList: list.New(),
		maxCache:  10000, // max 10k IP cache
		useCache:  useCache,
		apiUrl:    "http://ip-api.com/json/",
		stopChan:  make(chan struct{}),
	}
	
	// --- ÖĞRETİCİ ADIM 3: Goroutine Sızıntı (Leak) Koruması ---
	// Eski kodda arka plandaki goroutine sonsuza dek çalışıyor ve GeoLocator'ın
	// Garbage Collector tarafından temizlenmesini engelliyordu.
	// stopChan ekleyerek bu goroutine'in gerektiğinde temizce sonlanmasını sağladık.
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				gl.evictExpired()
			case <-gl.stopChan:
				return // Tünel durduğunda veya kapandığında goroutine sonlanır
			}
		}
	}()
	return gl
}

// Close stops the background eviction goroutine to prevent memory leaks
func (g *GeoLocator) Close() {
	if g.stopChan != nil {
		select {
		case <-g.stopChan:
			// Zaten kapatılmış
		default:
			close(g.stopChan)
		}
	}
}

// evictExpired - Süresi dolmuş (stale) verileri bellekten temizler
func (g *GeoLocator) evictExpired() {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()
	
	now := time.Now()
	// Listenin en arkasından (en eski erişilenlerden) başlayarak kontrol ediyoruz.
	// Böylece aktif kullanılan verileri gereksiz yere taramıyoruz.
	elem := g.evictList.Back()
	for elem != nil {
		next := elem.Prev()
		entry := elem.Value.(*geoEntry)
		if now.After(entry.expiresAt) {
			g.evictElement(elem)
		}
		elem = next
	}
}

// Lookup-ip konumu bulma
func (g *GeoLocator) Lookup(ip string) (*GeoLocation, error) {
	// Trim port if present
	if strings.Contains(ip, ":") {
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
	}
	ip = strings.TrimSpace(ip)

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

		loc := &GeoLocation{
			Country:     "Unknown",
			CountryCode: "UN",
			City:        "Unknown",
		}
		// Negative Cache: cache the temporary rate limit fallback for 5 minutes
		if g.useCache {
			g.saveToCacheWithTTL(ip, loc, 5*time.Minute)
		}
		return loc, nil
	}
	g.lastCall = time.Now()
	g.callMu.Unlock()

	loc, err := g.fetchFromAPI(ip)
	if err != nil {
		locErrorFallback := &GeoLocation{
			Country:     "Unknown",
			CountryCode: "UN",
			City:        "Unknown",
		}
		// Negative Cache: cache the API error fallback for 5 minutes
		if g.useCache {
			g.saveToCacheWithTTL(ip, locErrorFallback, 5*time.Minute)
		}
		return locErrorFallback, nil
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
		RegionName  string  `json:"regionName"` // regionName gives descriptive string instead of numeric region code
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
		Region:      apiResp.RegionName,
		Latitude:    apiResp.Lat,
		Longitude:   apiResp.Lon,
		Timezone:    apiResp.Timezone,
		ISP:         apiResp.ISP,
	}, nil
}

// get from cache (TTL and LRU access order aware)
func (g *GeoLocator) getFromCache(ip string) *GeoLocation {
	g.cacheMu.Lock() // LRU listesi güncelleneceği (MoveToFront) için Lock kullanıyoruz
	defer g.cacheMu.Unlock()

	entry, ok := g.cache[ip]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil
	}
	// --- ÖĞRETİCİ ADIM 4: Erişim Sırası Güncellemesi (LRU Hit) ---
	// Her başarılı önbellek isabetinde (cache hit), ilgili kaydı listenin en başına (Front) taşıyoruz.
	// Bu sayede sık kullanılan IP'lerin önbellekte kalmasını garanti altına alıyoruz.
	g.evictList.MoveToFront(entry.element)
	return entry.loc
}

// cache kaydet (with 24h TTL, evict oldest if at capacity)
func (g *GeoLocator) saveToCache(ip string, loc *GeoLocation) {
	g.saveToCacheWithTTL(ip, loc, 24*time.Hour)
}

// saveToCacheWithTTL kaydet (with custom TTL, evict oldest if at capacity)
func (g *GeoLocator) saveToCacheWithTTL(ip string, loc *GeoLocation, ttl time.Duration) {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()

	// Eğer IP zaten varsa, verisini güncelle ve en başa taşı
	if entry, ok := g.cache[ip]; ok {
		entry.loc = loc
		entry.expiresAt = time.Now().Add(ttl)
		g.evictList.MoveToFront(entry.element)
		return
	}

	// --- ÖĞRETİCİ ADIM 5: O(1) Sürede En Eski Kaydın Silinmesi (Eviction) ---
	// Önbellek maksimum kapasiteye (maxCache) ulaştığında, listenin en arkasındaki (Back)
	// yani en uzun süredir erişilmeyen/en eski kaydı O(1) sürede tespit edip siliyoruz.
	// Eski kodun yaptığı rastgele %10 silme hatası yerine, deterministik ve adil bir temizlik yapılıyor!
	if len(g.cache) >= g.maxCache {
		g.evictOldest()
	}

	entry := &geoEntry{
		ip:        ip,
		loc:       loc,
		expiresAt: time.Now().Add(ttl),
	}
	// Listenin başına ekle ve düğüm referansını kaydet
	elem := g.evictList.PushFront(entry)
	entry.element = elem
	g.cache[ip] = entry
}

func (g *GeoLocator) evictOldest() {
	elem := g.evictList.Back()
	if elem != nil {
		g.evictElement(elem)
	}
}

func (g *GeoLocator) evictElement(elem *list.Element) {
	g.evictList.Remove(elem)
	entry := elem.Value.(*geoEntry)
	delete(g.cache, entry.ip)
}

// istatistik
func (g *GeoLocator) CacheStats() map[string]interface{} {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()

	return map[string]interface{}{
		"cache_size": len(g.cache),
		"enabled":    g.useCache,
	}
}

// local ip kontrolü
func isLocalIP(ipStr string) bool {
	if strings.Contains(ipStr, ":") {
		if host, _, err := net.SplitHostPort(ipStr); err == nil {
			ipStr = host
		}
	}
	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "localhost" || ipStr == "" {
		return true
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		// Treat invalid IPs as local to avoid querying public APIs
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast()
}
