package db

import (
	"database/sql"
	"log"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	_ "github.com/lib/pq" // PostgreSQL driver
)

var db *sql.DB

func Connect() {
	cfg := config.LoadConfig()
	var err error
	db, err = sql.Open("postgres", cfg.DBConnectionUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	log.Println("Database connection established")
}

func GetDB() *sql.DB {
	return db
}

func Close() {
	if db != nil {
		db.Close()
		log.Println("Database connection closed")
	}
}
