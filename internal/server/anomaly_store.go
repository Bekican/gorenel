package server

import (
	"sync"
	"time"
)

// AnomalyRecord bir anomali kaydını temsil eder
type AnomalyRecord struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Subdomain    string    `json:"subdomain"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	ClientIP     string    `json:"client_ip"`
	AnomalyScore float64   `json:"anomaly_score"`
	DetectedBy   string    `json:"detected_by"` // "isolation_forest", "autoencoder", "both"
	IFScore      float64   `json:"if_score"`    // Isolation Forest specific
	AEScore      float64   `json:"ae_score"`    // AutoEncoder specific
	RiskReason   string    `json:"risk_reason"` // AI Security risk explanation
}

// AnomalyStore thread-safe anomali deposu
type AnomalyStore struct {
	records []AnomalyRecord
	mu      sync.RWMutex
	maxSize int
}

// NewAnomalyStore yeni bir anomali deposu oluşturur
func NewAnomalyStore(maxSize int) *AnomalyStore {
	return &AnomalyStore{
		records: make([]AnomalyRecord, 0),
		maxSize: maxSize,
	}
}

// Add yeni bir anomali kaydı ekler (en yenisi başta)
func (s *AnomalyStore) Add(record AnomalyRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append([]AnomalyRecord{record}, s.records...)

	if len(s.records) > s.maxSize {
		s.records = s.records[:s.maxSize]
	}
}

// GetRecent son N anomali kaydını döndürür
func (s *AnomalyStore) GetRecent(limit int) []AnomalyRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit > len(s.records) {
		limit = len(s.records)
	}
	result := make([]AnomalyRecord, limit)
	copy(result, s.records[:limit])
	return result
}

// Count toplam anomali sayısını döndürür
func (s *AnomalyStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}
