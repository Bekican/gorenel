package server

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/hashicorp/yamux"
)

// TunnelManager maintains the mapping between host names and active tunnel sessions.
// It supports both default subdomains (gorenel.io) and user-defined custom domains.
type TunnelManager struct {
	tunnels       map[string]*yamux.Session // key: subdomain, value: yamux session
	customDomains map[string]string         // key: custom_domain, value: subdomain mapping
	mu            sync.RWMutex
	tcpPorts      map[int]string
	udpPorts      map[int]string
	portMutex     sync.Mutex
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
				ln.Close()
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
	return &TunnelManager{
		tunnels:       make(map[string]*yamux.Session),
		customDomains: make(map[string]string),
		tcpPorts:      make(map[int]string),
		udpPorts:      make(map[int]string),
	}
}

// RegisterTunnel adds a new tunnel session to the manager.
// If customDomain is provided, it links the domain to the subdomain.
func (tm *TunnelManager) RegisterTunnel(subdomain string, session *yamux.Session, customDomain string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tunnels[subdomain] = session

	if customDomain != "" {
		tm.customDomains[customDomain] = subdomain
		log.Printf("Tünel kaydedildi: %s (Özel Domain: %s) [Toplam: %d]", subdomain, customDomain, len(tm.tunnels))
	} else {
		log.Printf("Tünel kaydedildi: %s [Toplam: %d]", subdomain, len(tm.tunnels))
	}
}

// GetTunnel attempts to find a session matching the provided host.
// It first checks if the host is a direct subdomain, then checks custom domain mappings.
func (tm *TunnelManager) GetTunnel(host string) (*yamux.Session, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// 1. Durum: Host direkt bir subdomain (örn: "xyz-123")
	if session, exists := tm.tunnels[host]; exists {
		return session, true
	}

	// 2. Durum: Host özel bir domain mi (örn: "api.bekircan.com")
	if sub, exists := tm.customDomains[host]; exists {
		if session, exists := tm.tunnels[sub]; exists {
			return session, true
		}
	}

	return nil, false
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
			log.Printf("[TunnelManager] Custom domain eşleşmesi silindi: %s", domain)
			break
		}
	}

	log.Printf("[TunnelManager] Tünel silindi: %s (Kalan: %d)", subdomain, len(tm.tunnels))
}

// Count returns the number of currently active tunnel sessions.
func (tm *TunnelManager) Count() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return len(tm.tunnels)
}
