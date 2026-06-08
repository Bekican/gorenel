package server

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type MemoryAnalyticsStore struct {
	mu      sync.RWMutex
	metrics []*Analytics
	logger  *zap.Logger
}

func NewMemoryAnalyticsStore(logger *zap.Logger) *MemoryAnalyticsStore {
	return &MemoryAnalyticsStore{
		metrics: make([]*Analytics, 0),
		logger:  logger,
	}
}

func (s *MemoryAnalyticsStore) Store(ctx context.Context, m *Analytics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics = append(s.metrics, m)

	s.logger.Info("AI Analytics stored",
		zap.String("req_id", m.RequestID),
		zap.String("provider", m.Provider),
		zap.Int("tokens", m.TokensUsed),
		zap.Bool("cache_hit", m.CacheHit))

	return nil
}

func (s *MemoryAnalyticsStore) GetMetrics() []*Analytics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	dst := make([]*Analytics, len(s.metrics))
	copy(dst, s.metrics)
	return dst
}
