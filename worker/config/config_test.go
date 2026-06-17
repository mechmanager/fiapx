package config_test

import (
	"testing"

	"github.com/mechmanager/fiapx/worker/config"
)

func TestLoad_AllEnvs(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://localhost/test")
	t.Setenv("RABBITMQ_URL", "amqp://localhost/")
	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_ACCESS_KEY", "key")
	t.Setenv("MINIO_SECRET_KEY", "secret")
	t.Setenv("PREFETCH_COUNT", "10")
	t.Setenv("QUEUE_NAME", "minha.fila")
	t.Setenv("MINIO_BUCKET", "meu-bucket")
	t.Setenv("MINIO_USE_SSL", "true")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if cfg.PrefetchCount != 10 {
		t.Errorf("PrefetchCount incorreto: %d", cfg.PrefetchCount)
	}
	if !cfg.MinIOUseSSL {
		t.Error("MinIOUseSSL deveria ser true")
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://localhost/test")
	t.Setenv("RABBITMQ_URL", "amqp://localhost/")
	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_ACCESS_KEY", "key")
	t.Setenv("MINIO_SECRET_KEY", "secret")
	t.Setenv("PREFETCH_COUNT", "invalido")
	t.Setenv("QUEUE_NAME", "")
	t.Setenv("MINIO_BUCKET", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if cfg.PrefetchCount != 5 {
		t.Errorf("PrefetchCount padrão incorreto: %d", cfg.PrefetchCount)
	}
	if cfg.QueueName != "video.process" {
		t.Errorf("QueueName padrão incorreto: %q", cfg.QueueName)
	}
}

func TestLoad_MissingDBDSN(t *testing.T) {
	t.Setenv("DB_DSN", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("esperava erro para DB_DSN ausente")
	}
}

func TestLoad_MissingRabbitURL(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://localhost/test")
	t.Setenv("RABBITMQ_URL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("esperava erro para RABBITMQ_URL ausente")
	}
}

func TestLoad_MissingMinIOEndpoint(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://localhost/test")
	t.Setenv("RABBITMQ_URL", "amqp://localhost/")
	t.Setenv("MINIO_ENDPOINT", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("esperava erro para MINIO_ENDPOINT ausente")
	}
}
