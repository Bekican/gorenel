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
	Policy       TunnelPolicy   `json:"policy,omitempty"`
}

type BandwidthInfo struct {
	In  int64 `json:"in"`
	Out int64 `json:"out"`
}

type TunnelPolicy struct {
	// KeyAuthToken is secret; never expose it via API.
	KeyAuthEnabled bool   `json:"key_auth_enabled,omitempty"`
	KeyAuthToken   string `json:"-"`

	IPAllowlistEnabled bool     `json:"ip_allowlist_enabled,omitempty"`
	IPAllowlist        []string `json:"-"`

	// Basic auth (secret password hash, never expose)
	BasicAuthEnabled      bool   `json:"basic_auth_enabled,omitempty"`
	BasicAuthUsername     string `json:"basic_auth_username,omitempty"`
	BasicAuthPasswordHash string `json:"-"`

	// HTTPS redirect (based on X-Forwarded-Proto)
	HttpsRedirectEnabled bool `json:"https_redirect_enabled,omitempty"`

	// Per-tunnel rate limit
	RateLimitEnabled  bool  `json:"rate_limit_enabled,omitempty"`
	RateLimitRequests int   `json:"rate_limit_requests,omitempty"`
	RateLimitWindowS  int64 `json:"rate_limit_window_s,omitempty"`

	// Header edits
	AddRequestHeaders     map[string]string `json:"add_request_headers,omitempty"`
	RemoveRequestHeaders  []string          `json:"remove_request_headers,omitempty"`
	AddResponseHeaders    map[string]string `json:"add_response_headers,omitempty"`
	RemoveResponseHeaders []string          `json:"remove_response_headers,omitempty"`

	// Path rewrite
	PathPrefix      string `json:"path_prefix,omitempty"`
	ReplacePathFrom string `json:"replace_path_from,omitempty"`
	ReplacePathTo   string `json:"replace_path_to,omitempty"`
	CORSEnabled     bool   `json:"cors_enabled,omitempty"`
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

// --- ÖĞRETİCİ ADIM 1: Protokol Uyumlu ve Çift Güvenlikli Port Tahsisi ---
// Eski kodda, UDP tünelleri için ayrılan portlar da TCP tablosuna kaydediliyor
// ve silinirken UDP tablosundan silinmeye çalışıldığı için bellek sızıntısına (leak) sebep oluyordu.
// Ayrıca bir portun işletim sisteminde hem TCP hem de UDP olarak tamamen boş olduğunu doğrulamak gerekir.
func (tm *TunnelManager) AllocatePort(proto string) (int, error) {
	tm.portMutex.Lock()
	defer tm.portMutex.Unlock()

	// 10000-20000 arası boş bir port bul
	for port := 10000; port <= 20000; port++ {
		_, tcpUsed := tm.tcpPorts[port]
		_, udpUsed := tm.udpPorts[port]

		// Eğer port belleğimizdeki hiçbir tünel tarafından kullanılmıyorsa:
		if !tcpUsed && !udpUsed {
			tcpFree := false
			udpFree := false

			// 1. TCP Portunun İşletim Sisteminde Boş Olduğunu Kontrol Et
			ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				_ = ln.Close()
				tcpFree = true
			}

			// 2. UDP Portunun İşletim Sisteminde Boş Olduğunu Kontrol Et
			addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
			if err == nil {
				lConn, err := net.ListenUDP("udp", addr)
				if err == nil {
					_ = lConn.Close()
					udpFree = true
				}
			}

			// Port hem TCP hem de UDP protokolleri için boş ise tahsis et
			if tcpFree && udpFree {
				if proto == "udp" {
					tm.udpPorts[port] = "reserved"
				} else {
					tm.tcpPorts[port] = "reserved"
				}
				return port, nil
			}
		}
	}
	return 0, fmt.Errorf("boş port bulunamadı")
}

// --- ÖĞRETİCİ ADIM 2: Eşleşen Tablodan Port Salınması ---
// Artık AllocatePort portu doğru protokole göre rezerve ettiği için,
// ReleasePort da ilgili portu doğru tablodan silecek ve bellek sızıntısı önlenecektir.
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
func (tm *TunnelManager) RegisterTunnel(subdomain string, session *yamux.Session, customDomain string, localPort int, publicUrl, userID, tunnelType string, policy TunnelPolicy) {
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
		Policy:       policy,
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

func (tm *TunnelManager) GetTunnelPolicy(host string) (TunnelPolicy, bool) {
	tm.mu.RLock()
	info, exists := tm.tunnels[host]
	if !exists {
		if sub, ok := tm.customDomains[host]; ok {
			info, exists = tm.tunnels[sub]
		}
	}
	tm.mu.RUnlock()
	if !exists || info == nil {
		return TunnelPolicy{}, false
	}
	return info.Policy, true
}

func (tm *TunnelManager) UpdateTunnelPolicy(subdomain string, update func(p *TunnelPolicy) error) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	info := tm.tunnels[subdomain]
	if info == nil {
		return fmt.Errorf("tunnel not found")
	}
	return update(&info.Policy)
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

func (tm *TunnelManager) GetTunnelsByUser(userID string) []*TunnelInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	var list []*TunnelInfo
	for _, info := range tm.tunnels {
		if info.UserID == userID {
			list = append(list, info)
		}
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

func (tm *TunnelManager) CountByUser(userID string) int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if userID == "" {
		return 0
	}

	count := 0
	for _, info := range tm.tunnels {
		if info.UserID == userID {
			count++
		}
	}
	return count
}

func (tm *TunnelManager) GetStatsByUser(userID string) (requests int64, bytesIn int64, bytesOut int64) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if userID == "" {
		return 0, 0, 0
	}

	for _, info := range tm.tunnels {
		if info.UserID == userID {
			requests += info.RequestCount
			bytesIn += info.Bandwidth.In
			bytesOut += info.Bandwidth.Out
		}
	}
	return
}


