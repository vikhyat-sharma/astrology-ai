package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/vikhyat-sharma/astrology-ai/internal/constants"
)

// Config holds all configuration for the application.
type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	Environment string
	OllamaURL   string
	OllamaModel string
	AllowedOrigins []string
}

// Load loads configuration from environment variables.
// Returns an error if any required secret is missing or invalid.
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

	allowedOrigins := []string{}
	if o := os.Getenv("ALLOWED_ORIGINS"); o != "" {
		for _, origin := range splitTrim(o, ",") {
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
	}

	return &Config{
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		Port:           getEnvOrDefault("PORT", constants.DefaultPort),
		Environment:    getEnvOrDefault("ENVIRONMENT", constants.DefaultEnvironment),
		OllamaURL:      getEnvOrDefault("OLLAMA_URL", constants.DefaultOllamaURL),
		OllamaModel:    getEnvOrDefault("OLLAMA_MODEL", constants.DefaultOllamaModel),
		AllowedOrigins: allowedOrigins,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func splitTrim(s, sep string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i:i+len(sep)] == sep {
			part := trim(s[start:i])
			out = append(out, part)
			start = i + len(sep)
		}
	}
	return out
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
