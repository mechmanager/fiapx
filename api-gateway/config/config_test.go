package config_test

import (
	"testing"

	"github.com/mechmanager/fiapx/api-gateway/config"
)

func TestLoad_AllEnvsSet(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo-valido-com-32-caracteres-ok")
	t.Setenv("GATEWAY_PORT", "9090")
	t.Setenv("AUTH_SERVICE_URL", "http://auth:8081")
	t.Setenv("UPLOAD_SERVICE_URL", "http://upload:8082")
	t.Setenv("STATUS_SERVICE_URL", "http://status:8083")
	t.Setenv("RATE_LIMIT_RPS", "50")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if cfg.GatewayPort != "9090" {
		t.Errorf("GatewayPort incorreto: %q", cfg.GatewayPort)
	}
	if cfg.JWTSecret != "segredo-valido-com-32-caracteres-ok" {
		t.Errorf("JWTSecret incorreto: %q", cfg.JWTSecret)
	}
	if cfg.AuthServiceURL != "http://auth:8081" {
		t.Errorf("AuthServiceURL incorreto: %q", cfg.AuthServiceURL)
	}
	if cfg.RateLimitRPS != 50 {
		t.Errorf("RateLimitRPS incorreto: %d", cfg.RateLimitRPS)
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo-valido-com-32-caracteres-ok")
	t.Setenv("GATEWAY_PORT", "")
	t.Setenv("AUTH_SERVICE_URL", "")
	t.Setenv("UPLOAD_SERVICE_URL", "")
	t.Setenv("STATUS_SERVICE_URL", "")
	t.Setenv("RATE_LIMIT_RPS", "invalido")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if cfg.GatewayPort != "8080" {
		t.Errorf("GatewayPort padrão incorreto: %q", cfg.GatewayPort)
	}
	if cfg.AuthServiceURL != "http://auth-service:8081" {
		t.Errorf("AuthServiceURL padrão incorreto: %q", cfg.AuthServiceURL)
	}
	if cfg.UploadServiceURL != "http://upload-service:8082" {
		t.Errorf("UploadServiceURL padrão incorreto: %q", cfg.UploadServiceURL)
	}
	if cfg.StatusServiceURL != "http://status-service:8083" {
		t.Errorf("StatusServiceURL padrão incorreto: %q", cfg.StatusServiceURL)
	}
	if cfg.RateLimitRPS != 100 {
		t.Errorf("RateLimitRPS padrão incorreto: %d", cfg.RateLimitRPS)
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("esperava erro para JWT_SECRET ausente")
	}
}
