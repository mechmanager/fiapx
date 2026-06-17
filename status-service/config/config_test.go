package config_test

import (
	"testing"

	"github.com/mechmanager/fiapx/status-service/config"
)

func setRequiredEnvs(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("MINIO_ENDPOINT", "minio:9000")
	t.Setenv("MINIO_ACCESS_KEY", "minioaccess")
	t.Setenv("MINIO_SECRET_KEY", "miniosecret")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
}

func TestLoad_MissingPostgresDSN(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing POSTGRES_DSN")
	}
}

func TestLoad_MissingMinIOEndpoint(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("MINIO_ENDPOINT", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing MINIO_ENDPOINT")
	}
}

func TestLoad_MissingMinIOAccessKey(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("MINIO_ENDPOINT", "minio:9000")
	t.Setenv("MINIO_ACCESS_KEY", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing MINIO_ACCESS_KEY")
	}
}

func TestLoad_MissingMinIOSecretKey(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("MINIO_ENDPOINT", "minio:9000")
	t.Setenv("MINIO_ACCESS_KEY", "key")
	t.Setenv("MINIO_SECRET_KEY", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing MINIO_SECRET_KEY")
	}
}

func TestLoad_MissingRedisURL(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("MINIO_ENDPOINT", "minio:9000")
	t.Setenv("MINIO_ACCESS_KEY", "key")
	t.Setenv("MINIO_SECRET_KEY", "secret")
	t.Setenv("REDIS_URL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing REDIS_URL")
	}
}

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnvs(t)
	t.Setenv("MINIO_BUCKET", "")
	t.Setenv("MINIO_USE_SSL", "")
	t.Setenv("STATUS_PORT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MinIOBucket != "videos" {
		t.Errorf("expected default bucket 'videos', got %q", cfg.MinIOBucket)
	}
	if cfg.MinIOUseSSL {
		t.Error("expected MinIOUseSSL false by default")
	}
	if cfg.StatusPort != "8083" {
		t.Errorf("expected default port '8083', got %q", cfg.StatusPort)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	setRequiredEnvs(t)
	t.Setenv("MINIO_BUCKET", "mybucket")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("STATUS_PORT", "9090")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MinIOBucket != "mybucket" {
		t.Errorf("unexpected bucket: %q", cfg.MinIOBucket)
	}
	if !cfg.MinIOUseSSL {
		t.Error("expected MinIOUseSSL true")
	}
	if cfg.StatusPort != "9090" {
		t.Errorf("unexpected port: %q", cfg.StatusPort)
	}
}
