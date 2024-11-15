package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

func LoadDBConnection() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}

	connectionStr := os.Getenv("DATABASE_URL")
	if connectionStr == "" {
		return "", errors.New("Connection String Not Found")
	}

	return connectionStr, nil
}
