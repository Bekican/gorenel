package server

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCache struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
	logger     *zap.Logger
}

func NewRedisCache(addr string, prefix string, ttl time.Duration, logger *zap.Logger) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisCache{
		client:     rdb,
		prefix:     prefix,
		defaultTTL: ttl,
		logger:     logger,
	}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		c.logger.Error("Redis cache get error", zap.Error(err), zap.String("key", key))
		return nil, err
	}
	return val, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	err := c.client.Set(ctx, c.prefix+key, value, ttl).Err()
	if err != nil {
		c.logger.Error("Redis cache set error", zap.Error(err), zap.String("key", key))
		return err
	}
	return nil
}
