package config

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConnectionUrl  string
	JWTAccessSecret  string
	JWTRefreshSecret string
}

var (
	once   sync.Once
	config *Config
)

func LoadConfig() *Config {
	once.Do(func() {
		// Determine the .env file to load
		envFile := ".env"
		if os.Getenv("GO_ENV") == "test" {
			envFile = ".env.test"
		}

		err := godotenv.Load(envFile)
		if err != nil {
			log.Printf("Warning: could not load %s file: %v", envFile, err)
		}

		config = &Config{
			DBConnectionUrl:  getEnv("DATABASE_URL"),
			JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET"),
			JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET"),
		}
	})
	return config
}

func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return ""
}

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
