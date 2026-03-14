package server

import (
	"fmt"
	"net"
	"sync"

	"time"

	"github.com/hashicorp/yamux"
	"go.uber.org/zap"
)

// TunnelInfo stores metadata and the session for an active tunnel.
type TunnelInfo struct {
	ID           string         `json:"id"`
	Subdomain    string         `json:"subdomain"`
	LocalPort    int            `json:"localPort"`
	PublicUrl    string         `json:"publicUrl"`
	Status       string         `json:"status"`
	RequestCount int64          `json:"requestCount"`
	Bandwidth    BandwidthInfo  `json:"bandwidth"`
	StartedAt    time.Time      `json:"startedAt"`
	LastActivity time.Time      `json:"lastActivity"`
	Session      *yamux.Session `json:"-"`
}

type BandwidthInfo struct {
	In  int64 `json:"in"`
	Out int64 `json:"out"`
}

// TunnelManager maintains the mapping between host names and active tunnel sessions.
// It supports both default subdomains (gorenel.site) and user-defined custom domains.
type TunnelManager struct {
	tunnels       map[string]*TunnelInfo
	customDomains map[string]string
	mu            sync.RWMutex
	tcpPorts      map[int]string
	udpPorts      map[int]string
	portMutex     sync.Mutex
	logger        *zap.Logger
}

func (tm *TunnelManager) AllocatePort() (int, error) {
	tm.portMutex.Lock()
	defer tm.portMutex.Unlock()

	// 10000-20000 arası boş bir port bul
	for port := 10000; port <= 20000; port++ {
		if _, used := tm.tcpPorts[port]; !used {
			// Portun sistemde gerçekten boş olup olmadığını kontrol et
			ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				_ = ln.Close()
				// Artık AllocatePort sadece portu rezerve eder
				// Kayıt işlemi Register aşamasında yapılacak veya burada geçici işaretlenebilir
				return port, nil
			}
		}
	}
	return 0, fmt.Errorf("boş port bulunamadı")
}

// NewTunnelManager creates a new instance of TunnelManager with initialized maps.
func NewTunnelManager() *TunnelManager {
	l, _ := zap.NewProduction()
	return &TunnelManager{
		tunnels:       make(map[string]*TunnelInfo),
		customDomains: make(map[string]string),
		tcpPorts:      make(map[int]string),
		udpPorts:      make(map[int]string),
		logger:        l,
	}
}

// RegisterTunnel adds a new tunnel session to the manager.
// If customDomain is provided, it links the domain to the subdomain.
func (tm *TunnelManager) RegisterTunnel(subdomain string, session *yamux.Session, customDomain string, localPort int, publicUrl string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tunnels[subdomain] = &TunnelInfo{
		ID:           subdomain,
		Subdomain:    subdomain,
		LocalPort:    localPort,
		PublicUrl:    publicUrl,
		Status:       "active",
		StartedAt:    time.Now(),
		LastActivity: time.Now(),
		Session:      session,
	}

	if customDomain != "" {
		tm.customDomains[customDomain] = subdomain
		tm.logger.Info("Tünel kaydedildi",
			zap.String("subdomain", subdomain),
			zap.String("custom_domain", customDomain),
			zap.Int("total", len(tm.tunnels)),
		)
	} else {
		tm.logger.Info("Tünel kaydedildi",
			zap.String("subdomain", subdomain),
			zap.Int("total", len(tm.tunnels)),
		)
	}
}

// GetTunnel attempts to find a session matching the provided host.
// It first checks if the host is a direct subdomain, then checks custom domain mappings.
func (tm *TunnelManager) GetTunnel(host string) (*yamux.Session, bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 1. Durum: Host direkt bir subdomain (örn: "xyz-123")
	if info, exists := tm.tunnels[host]; exists {
		info.LastActivity = time.Now()
		return info.Session, true
	}

	// 2. Durum: Host özel bir domain mi (örn: "api.bekircan.com")
	if sub, exists := tm.customDomains[host]; exists {
		if info, exists := tm.tunnels[sub]; exists {
			info.LastActivity = time.Now()
			return info.Session, true
		}
	}

	return nil, false
}

func (tm *TunnelManager) GetTunnels() []*TunnelInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	var list []*TunnelInfo
	for _, info := range tm.tunnels {
		list = append(list, info)
	}
	return list
}

// UpdateStats updates request count and bandwidth for a subdomain.
func (tm *TunnelManager) UpdateStats(subdomain string, bytesIn, bytesOut int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if info, exists := tm.tunnels[subdomain]; exists {
		info.RequestCount++
		info.Bandwidth.In += bytesIn
		info.Bandwidth.Out += bytesOut
		info.LastActivity = time.Now()
	}
}

// RemoveTunnel cleans up all mappings associated with a subdomain when it disconnects.
func (tm *TunnelManager) RemoveTunnel(subdomain string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Önce ana tünel oturumunu sil
	delete(tm.tunnels, subdomain)

	// Bu subdomain'e bağlı olan custom domain eşleşmesini bul ve temizle
	for domain, sub := range tm.customDomains {
		if sub == subdomain {
			delete(tm.customDomains, domain)
			tm.logger.Info("Custom domain eşleşmesi silindi", zap.String("domain", domain))
			break
		}
	}

	tm.logger.Info("Tünel silindi", zap.String("subdomain", subdomain), zap.Int("remaining", len(tm.tunnels)))
}

// Count returns the number of currently active tunnel sessions.
func (tm *TunnelManager) Count() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return len(tm.tunnels)
}
