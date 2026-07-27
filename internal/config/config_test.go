package config

import (
	"os"
	"testing"
)

func TestLoadMissingJWTSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("DATABASE_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	os.Setenv("JWT_SECRET", "a-secret-that-is-long-enough-32chars")
	defer os.Unsetenv("JWT_SECRET")
	os.Unsetenv("DATABASE_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoadJWTSecretTooShort(t *testing.T) {
	os.Setenv("JWT_SECRET", "short")
	defer os.Unsetenv("JWT_SECRET")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	defer os.Unsetenv("DATABASE_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is too short")
	}
}

func TestLoadSuccess(t *testing.T) {
	os.Setenv("JWT_SECRET", "a-secret-that-is-long-enough-32chars")
	defer os.Unsetenv("JWT_SECRET")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	defer os.Unsetenv("DATABASE_URL")
	os.Setenv("PORT", "9000")
	defer os.Unsetenv("PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9000" {
		t.Fatalf("expected port 9000, got %s", cfg.Port)
	}
	if cfg.JWTSecret != "a-secret-that-is-long-enough-32chars" {
		t.Fatal("JWT secret not loaded correctly")
	}
}

func TestLoadDefaultPort(t *testing.T) {
	os.Setenv("JWT_SECRET", "a-secret-that-is-long-enough-32chars")
	defer os.Unsetenv("JWT_SECRET")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	defer os.Unsetenv("DATABASE_URL")
	os.Unsetenv("PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
}
