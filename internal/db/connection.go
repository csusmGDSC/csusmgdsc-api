package db

import (
	"database/sql"
	"log"
	"sync"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	_ "github.com/lib/pq"
)

var (
	instance DatabaseConnection
	once     sync.Once
)

type PostgresConnection struct {
	db *sql.DB
}

func GetInstance() DatabaseConnection {
	once.Do(func() {
		instance = NewPostgresConnection()
		err := instance.Connect()
		if err != nil {
			log.Fatalf("Failed to initialize database connection: %v", err)
		}
	})
	return instance
}

func NewPostgresConnection() DatabaseConnection {
	return &PostgresConnection{}
}

func (p *PostgresConnection) Connect() error {
	cfg := config.LoadConfig()
	var err error
	p.db, err = sql.Open("postgres", cfg.DBConnectionUrl)
	if err != nil {
		return err
	}

	err = p.db.Ping()
	if err != nil {
		return err
	}

	log.Println("Database connection established")
	return nil
}

func (p *PostgresConnection) GetDB() *sql.DB {
	return p.db
}

func (p *PostgresConnection) Close() error {
	if p.db != nil {
		err := p.db.Close()
		if err != nil {
			return err
		}
		log.Println("Database connection closed")
	}
	return nil
}
