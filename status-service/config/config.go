// Package config carrega as configurações do status-service a partir de variáveis de ambiente.
package config

import (
	"fmt"
	"os"
)

// Config agrupa todas as configurações necessárias para o status-service.
type Config struct {
	PostgresDSN    string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool
	RedisURL       string
	StatusPort     string
}

// Load lê as variáveis de ambiente e devolve a configuração preenchida.
func Load() (*Config, error) {
	postgresDSN, err := requireEnv("POSTGRES_DSN")
	if err != nil {
		return nil, err
	}

	minioEndpoint, err := requireEnv("MINIO_ENDPOINT")
	if err != nil {
		return nil, err
	}

	minioAccessKey, err := requireEnv("MINIO_ACCESS_KEY")
	if err != nil {
		return nil, err
	}

	minioSecretKey, err := requireEnv("MINIO_SECRET_KEY")
	if err != nil {
		return nil, err
	}

	redisURL, err := requireEnv("REDIS_URL")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		PostgresDSN:    postgresDSN,
		MinIOEndpoint:  minioEndpoint,
		MinIOAccessKey: minioAccessKey,
		MinIOSecretKey: minioSecretKey,
		MinIOBucket:    getEnvOrDefault("MINIO_BUCKET", "videos"),
		MinIOUseSSL:    getEnvOrDefault("MINIO_USE_SSL", "false") == "true",
		RedisURL:       redisURL,
		StatusPort:     getEnvOrDefault("STATUS_PORT", "8083"),
	}

	return cfg, nil
}

func requireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("variável de ambiente obrigatória ausente: %s", key)
	}
	return value, nil
}

func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
