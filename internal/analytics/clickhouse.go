package analytics

import (
	"database/sql"
	"fmt"

	"github.com/Bekican/gorenel/internal/protocol"
	_ "github.com/ClickHouse/clickhouse-go"
)

type ClickHouseRepo struct {
	db *sql.DB
}

func NewClickHouseRepo(addr, dbName, user, pass string) (*ClickHouseRepo, error) {
	// TCP default port for clickhouse-go is 9000
	connStr := fmt.Sprintf("tcp://%s?database=%s&username=%s&password=%s", addr, dbName, user, pass)
	db, err := sql.Open("clickhouse", connStr)
	if err != nil {
		return nil, err
	}
	return &ClickHouseRepo{db: db}, nil
}

func (r *ClickHouseRepo) InitSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS request_events (
		subdomain String,
		method String,
		path String,
		user_agent String,
		client_ip String,
		status_code Int32,
		response_time_ms Int64,
		bytes_received Int64,
		bytes_sent Int64,
		geo_country String,
		geo_city String,
		timestamp DateTime64(3, 'UTC')
	) ENGINE = MergeTree()
	ORDER BY (subdomain, timestamp)
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *ClickHouseRepo) BatchInsert(events []*protocol.RequestEvent) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO request_events (subdomain, method, path, user_agent, client_ip, status_code, response_time_ms, bytes_received, bytes_sent, geo_country, geo_city, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		_, err := stmt.Exec(
			e.Subdomain,
			e.Method,
			e.Path,
			e.UserAgent,
			e.ClientIP,
			int32(e.StatusCode),
			e.ResponseTime.Milliseconds(),
			e.BytesReceived,
			e.ByteSent,
			e.GeoCountry,
			e.GeoCity,
			e.Timestamp,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
