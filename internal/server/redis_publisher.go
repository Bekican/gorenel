package server

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type TrafficData struct {
	Method       string `json:"method"`
	Path         string `json:"path"`
	StatusCode   int    `json:"status_code"`
	ResponseTime int64  `json:"response_time"`
	RequestSize  int64  `json:"request_size"`
	ResponseSize int64  `json:"response_size"`
	ClientIP     string `json:"client_ip"`
	Timestamp    string `json:"timestamp"`
}

type RedisPublisher struct {
	client     *redis.Client
	streamName string
}

func NewRedisPublisher(addr string) *RedisPublisher {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisPublisher{
		client:     rdb,
		streamName: "traffic_stream",
	}
}

func (r *RedisPublisher) Publish(data TrafficData) error {
	ctx := context.Background()
	jsonData, _ := json.Marshal(data)

	return r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: r.streamName,
		Values: map[string]interface{}{"data": string(jsonData)},
	}).Err()
}
