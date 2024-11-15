package db

import (
	"database/sql"
	"log"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func ConnectDB() (*sql.DB, error) {
	connectionStr, err := config.LoadDBConnection()
	if err != nil {
		log.Fatalf("Failed to load connection str: %v", err)
		return nil, err
	}

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
		return nil, err
	}

	log.Println("Database connection established")
	return db, nil
}
