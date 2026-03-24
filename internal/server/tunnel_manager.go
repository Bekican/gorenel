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
	UserID       string         `json:"userId,omitempty"`
	TunnelType   string         `json:"tunnelType,omitempty"`
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
				// Reserve immediately to prevent concurrent duplicate allocation.
				tm.tcpPorts[port] = "reserved"
				return port, nil
			}
		}
	}
	return 0, fmt.Errorf("boş port bulunamadı")
}

func (tm *TunnelManager) ReleasePort(port int, proto string) {
	tm.portMutex.Lock()
	defer tm.portMutex.Unlock()
	if proto == "udp" {
		delete(tm.udpPorts, port)
		return
	}
	delete(tm.tcpPorts, port)
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
func (tm *TunnelManager) RegisterTunnel(subdomain string, session *yamux.Session, customDomain string, localPort int, publicUrl, userID, tunnelType string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tunnels[subdomain] = &TunnelInfo{
		ID:           subdomain,
		Subdomain:    subdomain,
		UserID:       userID,
		TunnelType:   tunnelType,
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

// GetTunnelInfo returns a copy of tunnel metadata if present.
func (tm *TunnelManager) GetTunnelInfo(subdomain string) (*TunnelInfo, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	info, ok := tm.tunnels[subdomain]
	if !ok || info == nil {
		return nil, false
	}
	cp := *info
	cp.Session = nil
	return &cp, true
}

// GetTunnel attempts to find a session matching the provided host.
// It first checks if the host is a direct subdomain, then checks custom domain mappings.
func (tm *TunnelManager) GetTunnel(host string) (*yamux.Session, bool) {
	tm.mu.RLock()
	info, exists := tm.tunnels[host]
	if !exists {
		if sub, ok := tm.customDomains[host]; ok {
			info, exists = tm.tunnels[sub]
		}
	}
	tm.mu.RUnlock()

	if !exists || info == nil {
		return nil, false
	}

	tm.mu.Lock()
	if current, ok := tm.tunnels[info.Subdomain]; ok && current != nil {
		current.LastActivity = time.Now()
	}
	tm.mu.Unlock()

	return info.Session, true
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
