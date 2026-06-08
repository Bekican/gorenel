package server

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ReservedSubdomain struct {
	Subdomain          string     `json:"subdomain"`
	UserID             string     `json:"user_id"`
	AssignedAPIKeyHash *string    `json:"assigned_api_key_hash,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
}

type PostgresReservationRepository struct {
	db *sql.DB
}

func NewPostgresReservationRepository(db *sql.DB) *PostgresReservationRepository {
	return &PostgresReservationRepository{db: db}
}

func (r *PostgresReservationRepository) Init() error {
	q := `
	CREATE TABLE IF NOT EXISTS reserved_subdomains (
		subdomain VARCHAR(63) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		assigned_api_key_hash VARCHAR(64),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		last_used_at TIMESTAMP WITH TIME ZONE
	);
	CREATE INDEX IF NOT EXISTS idx_reserved_subdomains_user_id ON reserved_subdomains(user_id);
	`
	_, err := r.db.Exec(q)
	return err
}

func normalizeSubdomain(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimSuffix(s, ".gorenel.site")
	if s == "" {
		return "", fmt.Errorf("subdomain required")
	}
	if len(s) < 3 || len(s) > 63 {
		return "", fmt.Errorf("subdomain length must be 3-63")
	}
	for _, ch := range s {
		ok := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-'
		if !ok {
			return "", fmt.Errorf("invalid subdomain character: %q", ch)
		}
	}
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return "", fmt.Errorf("subdomain cannot start/end with '-'")
	}
	return s, nil
}

func (r *PostgresReservationRepository) ListByUser(userID string) ([]ReservedSubdomain, error) {
	rows, err := r.db.Query(`SELECT subdomain, user_id, assigned_api_key_hash, created_at, last_used_at FROM reserved_subdomains WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ReservedSubdomain{}
	for rows.Next() {
		var rec ReservedSubdomain
		if err := rows.Scan(&rec.Subdomain, &rec.UserID, &rec.AssignedAPIKeyHash, &rec.CreatedAt, &rec.LastUsedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

func (r *PostgresReservationRepository) Get(subdomain string) (*ReservedSubdomain, error) {
	norm, err := normalizeSubdomain(subdomain)
	if err != nil {
		return nil, err
	}
	var rec ReservedSubdomain
	err = r.db.QueryRow(`SELECT subdomain, user_id, assigned_api_key_hash, created_at, last_used_at FROM reserved_subdomains WHERE subdomain=$1`, norm).
		Scan(&rec.Subdomain, &rec.UserID, &rec.AssignedAPIKeyHash, &rec.CreatedAt, &rec.LastUsedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *PostgresReservationRepository) Create(userID, subdomain string) (*ReservedSubdomain, error) {
	norm, err := normalizeSubdomain(subdomain)
	if err != nil {
		return nil, err
	}
	_, err = r.db.Exec(`INSERT INTO reserved_subdomains(subdomain,user_id) VALUES ($1,$2)`, norm, userID)
	if err != nil {
		return nil, err
	}
	return r.Get(norm)
}

func (r *PostgresReservationRepository) Delete(userID, subdomain string) error {
	norm, err := normalizeSubdomain(subdomain)
	if err != nil {
		return err
	}
	res, err := r.db.Exec(`DELETE FROM reserved_subdomains WHERE subdomain=$1 AND user_id=$2`, norm, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (r *PostgresReservationRepository) Assign(userID, subdomain string, apiKeyHash *string) error {
	norm, err := normalizeSubdomain(subdomain)
	if err != nil {
		return err
	}
	res, err := r.db.Exec(`UPDATE reserved_subdomains SET assigned_api_key_hash=$1 WHERE subdomain=$2 AND user_id=$3`, apiKeyHash, norm, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (r *PostgresReservationRepository) TouchLastUsed(subdomain string) {
	norm, err := normalizeSubdomain(subdomain)
	if err != nil {
		return
	}
	_, _ = r.db.Exec(`UPDATE reserved_subdomains SET last_used_at=NOW() WHERE subdomain=$1`, norm)
}
