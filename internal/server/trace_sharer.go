package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TraceSharer struct {
	redis *redis.Client
	ttl   time.Duration
}

func NewTraceSharer(redisAddr string) *TraceSharer {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &TraceSharer{
		redis: rdb,
		ttl:   24 * time.Hour,
	}
}

// Share stores a captured request in Redis and returns a unique ShareID
func (s *TraceSharer) Share(ctx context.Context, captured *CapturedRequest) (string, error) {
	shareID := uuid.New().String()

	data, err := json.Marshal(captured)
	if err != nil {
		return "", err
	}

	err = s.redis.Set(ctx, "share:"+shareID, data, s.ttl).Err()
	if err != nil {
		return "", err
	}

	return shareID, nil
}

// Get retrieves a shared trace from Redis
func (s *TraceSharer) Get(ctx context.Context, shareID string) (*CapturedRequest, error) {
	val, err := s.redis.Get(ctx, "share:"+shareID).Result()
	if err != nil {
		return nil, err
	}

	var captured CapturedRequest
	if err := json.Unmarshal([]byte(val), &captured); err != nil {
		return nil, err
	}

	return &captured, nil
}
