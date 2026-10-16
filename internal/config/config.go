package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
)

// Config holds all configuration for the application.
type Config struct {
	DatabaseURL    string
	JWTSecret      string
	Port           string
	Environment    string
	OllamaURL      string
	OllamaModel    string
	AllowedOrigins []string
}

// Load reads configuration from environment variables (and an optional .env file).
// Returns an error if any required value is missing or invalid.
func Load() (*Config, error) {
	// Best-effort: load .env file if present (ignored in production containers).
	_ = godotenv.Load()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters, got %d", len(jwtSecret))
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is required")
	}

	var allowedOrigins []string
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			if o = strings.TrimSpace(o); o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	}

	return &Config{
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		Port:           envOr("PORT", constants.DefaultPort),
		Environment:    envOr("ENVIRONMENT", constants.DefaultEnvironment),
		OllamaURL:      envOr("OLLAMA_URL", constants.DefaultOllamaURL),
		OllamaModel:    envOr("OLLAMA_MODEL", constants.DefaultOllamaModel),
		AllowedOrigins: allowedOrigins,
	}, nil
}

// envOr returns the environment variable value, or fallback if it is unset/empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
