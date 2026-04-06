package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	Environment string
	OllamaURL   string
	OllamaModel string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/astrology_ai?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		OllamaURL:   getEnv("OLLAMA_URL", "http://127.0.0.1:11434"),
		OllamaModel: getEnv("OLLAMA_MODEL", "llama3"),
	}

	// Warn about default secrets in development
	if config.Environment == "development" && config.JWTSecret == "your-secret-key" {
		log.Println("WARNING: Using default JWT_SECRET. Set JWT_SECRET environment variable for security.")
	}
	if config.Environment == "development" && strings.Contains(config.DatabaseURL, "password") {
		log.Println("WARNING: Using default database password. Set DATABASE_URL environment variable for security.")
	}

	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
