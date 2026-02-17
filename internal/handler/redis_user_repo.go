package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Bekican/gorenel/pkg/auth"
	"github.com/redis/go-redis/v9"
)

const (
	userEmailPrefix = "user:email:"
	userIDPrefix    = "user:id:"
)

type RedisUserRepository struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisUserRepository(client *redis.Client) *RedisUserRepository {
	return &RedisUserRepository{
		client: client,
		ctx:    context.Background(),
	}
}

func (r *RedisUserRepository) GetByEmail(email string) (*auth.User, error) {
	id, err := r.client.Get(r.ctx, userEmailPrefix+email).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

func (r *RedisUserRepository) GetByID(id string) (*auth.User, error) {
	val, err := r.client.Get(r.ctx, userIDPrefix+id).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	var user auth.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *RedisUserRepository) Create(user *auth.User) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Use a transaction (pipeline) to ensure atomicity
	pipe := r.client.Pipeline()
	pipe.Set(r.ctx, userIDPrefix+user.ID, data, 0)
	pipe.Set(r.ctx, userEmailPrefix+user.Email, user.ID, 0)
	_, err = pipe.Exec(r.ctx)
	return err
}

func (r *RedisUserRepository) UpdateLoginTime(id string) error {
	user, err := r.GetByID(id)
	if err != nil {
		return err
	}

	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now

	return r.Create(user)
}
