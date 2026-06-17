package config_test

import (
	"testing"

	"github.com/mechmanager/fiapx/notification-service/config"
)

func TestLoad_MissingPostgresDSN(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "")
	t.Setenv("RABBITMQ_URL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing POSTGRES_DSN")
	}
}

func TestLoad_MissingRabbitMQURL(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("RABBITMQ_URL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing RABBITMQ_URL")
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("RABBITMQ_URL", "amqp://localhost")
	t.Setenv("NOTIFICATION_QUEUE", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_FROM", "")
	t.Setenv("MAX_RETRIES", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.NotificationQueue != "notification" {
		t.Errorf("expected default queue 'notification', got %q", cfg.NotificationQueue)
	}
	if cfg.SMTPPort != "587" {
		t.Errorf("expected default port '587', got %q", cfg.SMTPPort)
	}
	if cfg.SMTPFrom != "noreply@fiapx.com" {
		t.Errorf("expected default from, got %q", cfg.SMTPFrom)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected default MaxRetries 3, got %d", cfg.MaxRetries)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://custom")
	t.Setenv("RABBITMQ_URL", "amqp://custom")
	t.Setenv("NOTIFICATION_QUEUE", "custom-queue")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USER", "user@example.com")
	t.Setenv("SMTP_PASS", "secret")
	t.Setenv("SMTP_FROM", "from@example.com")
	t.Setenv("MAX_RETRIES", "5")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PostgresDSN != "postgres://custom" {
		t.Errorf("unexpected PostgresDSN: %q", cfg.PostgresDSN)
	}
	if cfg.NotificationQueue != "custom-queue" {
		t.Errorf("unexpected NotificationQueue: %q", cfg.NotificationQueue)
	}
	if cfg.SMTPHost != "smtp.example.com" {
		t.Errorf("unexpected SMTPHost: %q", cfg.SMTPHost)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("unexpected MaxRetries: %d", cfg.MaxRetries)
	}
}

func TestLoad_InvalidMaxRetries(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://test")
	t.Setenv("RABBITMQ_URL", "amqp://localhost")
	t.Setenv("MAX_RETRIES", "notanumber")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected fallback MaxRetries 3, got %d", cfg.MaxRetries)
	}
}
