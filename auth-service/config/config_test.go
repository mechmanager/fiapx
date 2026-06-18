package config_test

import (
	"testing"

	"github.com/mechmanager/fiapx/auth-service/config"
)

func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_DSN", "postgres://test:test@localhost/test")
	t.Setenv("JWT_SECRET", "test-secret-that-is-long-enough-ok")
}

func TestLoad_Defaults(t *testing.T) {
	setRequired(t)
	t.Setenv("JWT_EXPIRATION_HOURS", "")
	t.Setenv("AUTH_PORT", "")

	cfg := config.Load()

	if cfg.JWTExpirationHours != 24 {
		t.Errorf("expected default JWTExpirationHours 24, got %d", cfg.JWTExpirationHours)
	}
	if cfg.AuthPort != "8081" {
		t.Errorf("expected default AuthPort '8081', got %q", cfg.AuthPort)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	setRequired(t)
	t.Setenv("JWT_EXPIRATION_HOURS", "48")
	t.Setenv("AUTH_PORT", "9090")

	cfg := config.Load()

	if cfg.JWTExpirationHours != 48 {
		t.Errorf("expected JWTExpirationHours 48, got %d", cfg.JWTExpirationHours)
	}
	if cfg.AuthPort != "9090" {
		t.Errorf("expected AuthPort '9090', got %q", cfg.AuthPort)
	}
}

func TestLoad_RequiredFields(t *testing.T) {
	setRequired(t)

	cfg := config.Load()

	if cfg.PostgresDSN == "" {
		t.Error("expected non-empty PostgresDSN")
	}
	if cfg.JWTSecret == "" {
		t.Error("expected non-empty JWTSecret")
	}
}
