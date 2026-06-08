package server

import (
	"database/sql"
	"sync"
	"time"
)

type TunnelSessionRecord struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Subdomain    string     `json:"subdomain"`
	TunnelType   string     `json:"tunnel_type"`
	LocalPort    int        `json:"local_port"`
	PublicURL    string     `json:"public_url"`
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	RequestCount int64      `json:"request_count"`
	BytesIn      int64      `json:"bytes_in"`
	BytesOut     int64      `json:"bytes_out"`
	AvgRPS       float64    `json:"avg_rps"`
}

type TunnelHistoryStore struct {
	db *sql.DB

	mu                sync.RWMutex
	activeBySubdomain map[string]string
}

func NewTunnelHistoryStore(db *sql.DB) *TunnelHistoryStore {
	return &TunnelHistoryStore{
		db:                db,
		activeBySubdomain: make(map[string]string),
	}
}

func (s *TunnelHistoryStore) Init() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS tunnel_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    subdomain TEXT NOT NULL,
    tunnel_type TEXT NOT NULL,
    local_port INT NOT NULL,
    public_url TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ NULL,
    request_count BIGINT NOT NULL DEFAULT 0,
    bytes_in BIGINT NOT NULL DEFAULT 0,
    bytes_out BIGINT NOT NULL DEFAULT 0,
    avg_rps DOUBLE PRECISION NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_tunnel_sessions_subdomain_started ON tunnel_sessions(subdomain, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_tunnel_sessions_user_started ON tunnel_sessions(user_id, started_at DESC);

CREATE TABLE IF NOT EXISTS anomaly_events (
    id TEXT PRIMARY KEY,
    session_id TEXT NULL REFERENCES tunnel_sessions(id) ON DELETE SET NULL,
    subdomain TEXT NOT NULL,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    client_ip TEXT NOT NULL,
    anomaly_score DOUBLE PRECISION NOT NULL,
    detected_by TEXT NOT NULL,
    if_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    ae_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    risk_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_anomaly_events_subdomain_created ON anomaly_events(subdomain, created_at DESC);
`)
	return err
}

func (s *TunnelHistoryStore) StartSession(rec TunnelSessionRecord) error {
	if rec.StartedAt.IsZero() {
		rec.StartedAt = time.Now()
	}

	_, err := s.db.Exec(
		`INSERT INTO tunnel_sessions (id, user_id, subdomain, tunnel_type, local_port, public_url, started_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		rec.ID, rec.UserID, rec.Subdomain, rec.TunnelType, rec.LocalPort, rec.PublicURL, rec.StartedAt,
	)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.activeBySubdomain[rec.Subdomain] = rec.ID
	s.mu.Unlock()
	return nil
}

func (s *TunnelHistoryStore) EndSession(subdomain string, requestCount, bytesIn, bytesOut int64, startedAt time.Time) error {
	endedAt := time.Now()
	durationSec := endedAt.Sub(startedAt).Seconds()
	avgRPS := 0.0
	if durationSec > 0 {
		avgRPS = float64(requestCount) / durationSec
	}

	s.mu.RLock()
	sessionID := s.activeBySubdomain[subdomain]
	s.mu.RUnlock()
	if sessionID == "" {
		return nil
	}

	_, err := s.db.Exec(
		`UPDATE tunnel_sessions
		 SET ended_at=$1, request_count=$2, bytes_in=$3, bytes_out=$4, avg_rps=$5
		 WHERE id=$6`,
		endedAt, requestCount, bytesIn, bytesOut, avgRPS, sessionID,
	)
	if err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.activeBySubdomain, subdomain)
	s.mu.Unlock()
	return nil
}

func (s *TunnelHistoryStore) PersistAnomaly(record AnomalyRecord) error {
	s.mu.RLock()
	sessionID := s.activeBySubdomain[record.Subdomain]
	s.mu.RUnlock()
	var sessionRef interface{}
	if sessionID != "" {
		sessionRef = sessionID
	}
	_, err := s.db.Exec(
		`INSERT INTO anomaly_events (id, session_id, subdomain, method, path, client_ip, anomaly_score, detected_by, if_score, ae_score, risk_reason, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		record.ID, sessionRef, record.Subdomain, record.Method, record.Path, record.ClientIP, record.AnomalyScore, record.DetectedBy, record.IFScore, record.AEScore, record.RiskReason, record.Timestamp,
	)
	return err
}

func (s *TunnelHistoryStore) ListRecentSessions(limit int) ([]TunnelSessionRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, subdomain, tunnel_type, local_port, public_url, started_at, ended_at, request_count, bytes_in, bytes_out, avg_rps
		 FROM tunnel_sessions
		 ORDER BY started_at DESC
		 LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]TunnelSessionRecord, 0, limit)
	for rows.Next() {
		var r TunnelSessionRecord
		if err := rows.Scan(&r.ID, &r.UserID, &r.Subdomain, &r.TunnelType, &r.LocalPort, &r.PublicURL, &r.StartedAt, &r.EndedAt, &r.RequestCount, &r.BytesIn, &r.BytesOut, &r.AvgRPS); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *TunnelHistoryStore) ListRecentSessionsByUser(userID string, limit int) ([]TunnelSessionRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, subdomain, tunnel_type, local_port, public_url, started_at, ended_at, request_count, bytes_in, bytes_out, avg_rps
		 FROM tunnel_sessions
		 WHERE user_id = $1
		 ORDER BY started_at DESC
		 LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]TunnelSessionRecord, 0, limit)
	for rows.Next() {
		var r TunnelSessionRecord
		if err := rows.Scan(&r.ID, &r.UserID, &r.Subdomain, &r.TunnelType, &r.LocalPort, &r.PublicURL, &r.StartedAt, &r.EndedAt, &r.RequestCount, &r.BytesIn, &r.BytesOut, &r.AvgRPS); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
