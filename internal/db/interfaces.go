package db

import "database/sql"

type DatabaseConnection interface {
	GetDB() *sql.DB
	Connect() error
	Close() error
}
