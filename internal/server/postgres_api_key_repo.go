package server

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresAPIKeyRepository struct {
	db *sql.DB
}

func NewPostgresAPIKeyRepository(db *sql.DB) *PostgresAPIKeyRepository {
	return &PostgresAPIKeyRepository{
		db: db,
	}
}

func (r *PostgresAPIKeyRepository) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS api_keys (
		key_hash VARCHAR(64) PRIMARY KEY,
		key_value VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE,
		usage_count BIGINT DEFAULT 0,
		rate_limit INT DEFAULT 100
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *PostgresAPIKeyRepository) GetByHash(hash string) (*APIKey, error) {
	query := `SELECT key_value, user_id, created_at, expires_at, usage_count, rate_limit FROM api_keys WHERE key_hash = $1`
	var k APIKey
	var key_value string
	err := r.db.QueryRow(query, hash).Scan(
		&key_value, &k.UserID, &k.CreatedAt, &k.ExpiresAt, &k.UsageCount, &k.RateLimit,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("api key not found")
	}
	if err != nil {
		return nil, err
	}
	k.Key = key_value
	return &k, nil
}

func (r *PostgresAPIKeyRepository) Create(apiKey *APIKey) error {
	query := `
	INSERT INTO api_keys (key_hash, key_value, user_id, created_at, expires_at, usage_count, rate_limit)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	keyHash := hashKey(apiKey.Key)
	_, err := r.db.Exec(query,
		keyHash, apiKey.Key, apiKey.UserID, apiKey.CreatedAt, apiKey.ExpiresAt, apiKey.UsageCount, apiKey.RateLimit,
	)
	return err
}

func (r *PostgresAPIKeyRepository) IncrementUsage(hash string) error {
	query := `UPDATE api_keys SET usage_count = usage_count + 1 WHERE key_hash = $1`
	_, err := r.db.Exec(query, hash)
	return err
}

func (r *PostgresAPIKeyRepository) ListAll() ([]*APIKey, error) {
	query := `SELECT key_value, user_id, created_at, expires_at, usage_count, rate_limit FROM api_keys`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var k APIKey
		var key_value string
		if err := rows.Scan(&key_value, &k.UserID, &k.CreatedAt, &k.ExpiresAt, &k.UsageCount, &k.RateLimit); err != nil {
			return nil, err
		}
		k.Key = key_value
		keys = append(keys, &k)
	}
	return keys, nil
}

func (r *PostgresAPIKeyRepository) Delete(hash string) error {
	query := `DELETE FROM api_keys WHERE key_hash = $1`
	_, err := r.db.Exec(query, hash)
	return err
}
