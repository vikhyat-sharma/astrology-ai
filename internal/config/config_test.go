package config

import (
	"os"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("PORT")

	cfg := Load()

	if cfg == nil {
		t.Fatal("expected config to be non-nil")
	}

	if cfg.Port != "8080" {
		t.Fatalf("expected port 8080, got %s", cfg.Port)
	}
}

func TestLoadEnvironment(t *testing.T) {
	os.Setenv("PORT", "9000")
	defer os.Unsetenv("PORT")

	cfg := Load()

	if cfg.Port != "9000" {
		t.Fatalf("expected port 9000, got %s", cfg.Port)
	}
}
