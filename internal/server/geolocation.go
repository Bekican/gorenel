package server

import (
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

type GeoLocator struct {
	//cache
	cache   map[string]*GeoLocation
	cacheMu sync.RWMutex

	//rate limtiing
	lastCall time.Time
	callMu   sync.Mutex

	//Api ayarlari
	apiUrl   string
	useCache bool
}

// yeni geolocator oluşturuyoruz
func (g *GeoLocator) Lookup(ip string) (*GeoLocation, error) {
	if isLocalIP(ip) {
		return &GeoLocation{
			Country:     "Local",
			CountryCode: "Lo",
			City:        "Localhost",
		}, nil
	}
	if g.useCache {
		if loc := g.getFromCache(ip); loc != nil {
			return loc, nil
		}
	}

	//rate limiting
}
