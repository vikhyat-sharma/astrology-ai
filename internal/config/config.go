package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
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
		DatabaseURL: getEnv("DATABASE_URL", constants.DefaultDatabaseURL),
		JWTSecret:   getEnv("JWT_SECRET", constants.DefaultJWTSecret),
		Port:        getEnv("PORT", constants.DefaultPort),
		Environment: getEnv("ENVIRONMENT", constants.DefaultEnvironment),
		OllamaURL:   getEnv("OLLAMA_URL", constants.DefaultOllamaURL),
		OllamaModel: getEnv("OLLAMA_MODEL", constants.DefaultOllamaModel),
	}

	// Warn about default secrets in development
	if config.Environment == constants.DefaultEnvironment && config.JWTSecret == constants.DefaultJWTSecret {
		log.Println("WARNING: Using default JWT_SECRET. Set JWT_SECRET environment variable for security.")
	}
	if config.Environment == constants.DefaultEnvironment && strings.Contains(config.DatabaseURL, "password") {
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
