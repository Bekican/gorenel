package handler

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Bekican/gorenel/pkg/auth"
	_ "github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		email VARCHAR(255) NOT NULL UNIQUE,
		name VARCHAR(255),
		password_hash TEXT,
		avatar_url TEXT,
		provider VARCHAR(50) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		last_login TIMESTAMP WITH TIME ZONE
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}

	// Schema Evolutions: Automatically add missing columns if upgrading an existing database
	evolutions := []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;`,
	}
	for _, sql := range evolutions {
		_, err = r.db.Exec(sql)
		if err != nil {
			// Sadece loglayabiliriz ama genelde postgres IF NOT EXISTS'i desteklediği için güvenlidir.
			return err
		}
	}

	return nil
}

func (r *PostgresUserRepository) GetByEmail(email string) (*auth.User, error) {
	query := `SELECT id, email, name, password_hash, avatar_url, provider, created_at, updated_at, last_login FROM users WHERE email = $1`
	return r.queryUser(query, email)
}

func (r *PostgresUserRepository) GetByID(id string) (*auth.User, error) {
	query := `SELECT id, email, name, password_hash, avatar_url, provider, created_at, updated_at, last_login FROM users WHERE id = $1`
	return r.queryUser(query, id)
}

func (r *PostgresUserRepository) queryUser(query string, arg interface{}) (*auth.User, error) {
	var user auth.User
	err := r.db.QueryRow(query, arg).Scan(
		&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.AvatarURL, &user.Provider,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) Create(user *auth.User) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	query := `
	INSERT INTO users (id, email, name, password_hash, avatar_url, provider, created_at, updated_at, last_login)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(query,
		user.ID, user.Email, user.Name, user.PasswordHash, user.AvatarURL, user.Provider,
		user.CreatedAt, user.UpdatedAt, user.LastLogin,
	)
	return err
}

func (r *PostgresUserRepository) UpdateLoginTime(id string) error {
	now := time.Now()
	query := `UPDATE users SET last_login = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, now, now, id)
	return err
}
