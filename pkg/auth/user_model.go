package auth

import (
	"time"
)

type User struct {
	ID           string     `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	Name          string     `json:"name" db:"name"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	AvatarURL     string     `json:"avatar_url,omitempty" db:"avatar_url"`
	Provider      string     `json:"provider" db:"provider"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin     *time.Time `json:"last_login,omitempty" db:"last_login"`
}

// user repository
type UserRepository interface {
	GetByEmail(email string) (*User, error)
	GetByID(id string) (*User, error)
	Create(user *User) error
	UpdateLoginTime(id string) error
}

/*
-- SQL Migration Example (PostgreSQL) --

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    avatar_url TEXT,
    provider VARCHAR(50) NOT NULL, -- 'google', 'github'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email);
*/
