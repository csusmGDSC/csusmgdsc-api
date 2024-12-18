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

func LoadJWTAccessSecret() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}

	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if jwtAccessSecret == "" {
		return "", errors.New("JWT Secret Not Found")
	}

	return jwtAccessSecret, nil
}

func LoadJWTRefreshSecret() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}

	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		return "", errors.New("JWT Refresh Secret Not Found")
	}

	return jwtRefreshSecret, nil
}
