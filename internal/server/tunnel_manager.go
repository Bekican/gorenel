package server

import (
	"log"
	"sync"

	"github.com/hashicorp/yamux"
)

type TunnelManager struct {
	tunnels map[string]*yamux.Session
	mu      sync.RWMutex
}

// yeni tünel oluşturur
func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		tunnels: make(map[string]*yamux.Session),
	}
}

func (tm *TunnelManager) RegisterTunnel(subdomain string, session *yamux.Session) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tunnels[subdomain] = session
	log.Printf("Tünel kaydedildi: %s (Toplam: %d)", subdomain, len(tm.tunnels))
}

func (tm *TunnelManager) GetTunnel(subdomain string) (*yamux.Session, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	session, exists := tm.tunnels[subdomain]
	return session, exists
}

// bağlantı kapandığında tüneli sil
func (tm *TunnelManager) RemoveTunnel(subdomain string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.tunnels, subdomain)
	log.Printf("[TunnelManager] Tünel silindi : %s", subdomain)
}

func (tm *TunnelManager) Count() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return len(tm.tunnels)
}
