package handler

import (
	"errors"
	"sync"

	"github.com/Bekican/gorenel/pkg/auth"
)

type InMemoryUserRepo struct {
	mu    sync.RWMutex
	users map[string]*auth.User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		users: make(map[string]*auth.User),
	}
}

func (r *InMemoryUserRepo) GetByEmail(email string) (*auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *InMemoryUserRepo) GetByID(id string) (*auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (r *InMemoryUserRepo) Create(user *auth.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepo) UpdateLoginTime(id string) error {
	return nil
}
