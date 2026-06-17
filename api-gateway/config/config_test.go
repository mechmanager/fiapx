package config_test

import (
	"testing"

	"github.com/mechmanager/fiapx/api-gateway/config"
)

func TestLoad_AllEnvsSet(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://localhost/test")
	t.Setenv("RABBITMQ_URL", "amqp://localhost/")
	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_ACCESS_KEY", "key")
	t.Setenv("MINIO_SECRET_KEY", "secret")
	t.Setenv("JWT_SECRET", "segredo")
	t.Setenv("API_PORT", "9090")
	t.Setenv("QUEUE_NAME", "minha.fila")
	t.Setenv("MINIO_BUCKET", "meu-bucket")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("JWT_EXPIRATION_HOURS", "48")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if cfg.APIPort != "9090" {
		t.Errorf("APIPort incorreto: %q", cfg.APIPort)
	}
	if cfg.JWTExpirationHours != 48 {
		t.Errorf("JWTExpirationHours incorreto: %d", cfg.JWTExpirationHours)
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
	t.Setenv("JWT_SECRET", "segredo")
	// Variáveis opcionais não definidas: devem usar default.
	t.Setenv("API_PORT", "")
	t.Setenv("QUEUE_NAME", "")
	t.Setenv("MINIO_BUCKET", "")
	t.Setenv("MINIO_USE_SSL", "")
	t.Setenv("JWT_EXPIRATION_HOURS", "invalido")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if cfg.APIPort != "8080" {
		t.Errorf("esperava porta padrão 8080, obteve %q", cfg.APIPort)
	}
	if cfg.QueueName != "video.process" {
		t.Errorf("QueueName padrão incorreto: %q", cfg.QueueName)
	}
	if cfg.JWTExpirationHours != 24 {
		t.Errorf("JWTExpirationHours padrão incorreto: %d", cfg.JWTExpirationHours)
	}
}

func TestLoad_MissingDBDSN(t *testing.T) {
	t.Setenv("DB_DSN", "")
	t.Setenv("RABBITMQ_URL", "amqp://localhost/")
	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_ACCESS_KEY", "key")
	t.Setenv("MINIO_SECRET_KEY", "secret")
	t.Setenv("JWT_SECRET", "segredo")

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

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://localhost/test")
	t.Setenv("RABBITMQ_URL", "amqp://localhost/")
	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_ACCESS_KEY", "key")
	t.Setenv("MINIO_SECRET_KEY", "secret")
	t.Setenv("JWT_SECRET", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("esperava erro para JWT_SECRET ausente")
	}
}
