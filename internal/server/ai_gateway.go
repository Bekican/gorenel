package server

import (
	"context"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type AIProvider interface {
	GetName() string
	GetURL() string
	GetAUthHeader() string
}

type RedisCache struct {
	Client     *redis.Client
	Ctx        context.Context
	Prefix     string
	DefaultTTL time.Duration
	Logger     *zap.Logger
}

type AIGateway struct {
	Providers map[string]AIProvider
	Logger    *zap.Logger
	AICache   *RedisCache
}

func (g *AIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pName := r.URL.Query().Get("provider")
	provider, ok := g.Providers[pName]

	if !ok {

	}
}
