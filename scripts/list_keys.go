package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Bekican/gorenel/internal/config"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT key_value, user_id FROM api_keys")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Existing API Keys:")
	for rows.Next() {
		var key, userID string
		if err := rows.Scan(&key, &userID); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Key: %s | User: %s\n", key, userID)
	}
}
