package config_test

import (
	"testing"

	"github.com/mechmanager/fiapx/upload-service/config"
)

func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("RABBITMQ_URL", "amqp://localhost")
	t.Setenv("MINIO_ENDPOINT", "minio:9000")
	t.Setenv("MINIO_ACCESS_KEY", "access")
	t.Setenv("MINIO_SECRET_KEY", "secret")
}

func TestLoad_Defaults(t *testing.T) {
	setRequired(t)
	t.Setenv("MINIO_BUCKET", "")
	t.Setenv("MINIO_USE_SSL", "")
	t.Setenv("QUEUE_NAME", "")
	t.Setenv("UPLOAD_PORT", "")

	cfg := config.Load()

	if cfg.MinioBucket != "videos" {
		t.Errorf("expected default bucket 'videos', got %q", cfg.MinioBucket)
	}
	if cfg.MinIOUseSSL {
		t.Error("expected MinIOUseSSL false by default")
	}
	if cfg.QueueName != "video.upload" {
		t.Errorf("expected default queue 'video.upload', got %q", cfg.QueueName)
	}
	if cfg.UploadPort != "8082" {
		t.Errorf("expected default port '8082', got %q", cfg.UploadPort)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	setRequired(t)
	t.Setenv("MINIO_BUCKET", "mybucket")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("QUEUE_NAME", "custom.queue")
	t.Setenv("UPLOAD_PORT", "9090")

	cfg := config.Load()

	if cfg.MinioBucket != "mybucket" {
		t.Errorf("unexpected bucket: %q", cfg.MinioBucket)
	}
	if !cfg.MinIOUseSSL {
		t.Error("expected MinIOUseSSL true")
	}
	if cfg.QueueName != "custom.queue" {
		t.Errorf("unexpected queue: %q", cfg.QueueName)
	}
	if cfg.UploadPort != "9090" {
		t.Errorf("unexpected port: %q", cfg.UploadPort)
	}
}

func TestLoad_RequiredFields(t *testing.T) {
	setRequired(t)

	cfg := config.Load()

	if cfg.PostgresDSN == "" {
		t.Error("expected non-empty PostgresDSN")
	}
	if cfg.RabbitMQURL == "" {
		t.Error("expected non-empty RabbitMQURL")
	}
	if cfg.MinIOEndpoint == "" {
		t.Error("expected non-empty MinIOEndpoint")
	}
}
