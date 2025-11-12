package main

import (
	"os"

	"poker-platform/backend/internal/db"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application
type Config struct {
	// Database configuration
	DBConfig db.Config

	// Server configuration
	ServerPort  string
	Environment string

	// Authentication
	JWTSecret string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() Config {
	// Load .env file if it exists
	godotenv.Load()

	return Config{
		DBConfig: db.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "poker_platform"),
		},
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		Environment: getEnv("ENV", "development"),
		JWTSecret:   getEnv("JWT_SECRET", "secret"),
	}
}

// getEnv retrieves an environment variable or returns a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
