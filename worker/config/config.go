// Package config carrega as configurações do worker a partir de variáveis de ambiente.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config agrupa todas as configurações necessárias para o worker.
type Config struct {
	DBDSN          string // string de conexão com o PostgreSQL
	RabbitMQURL    string // URL de conexão com o RabbitMQ
	QueueName      string // nome da fila a consumir
	PrefetchCount  int    // quantidade máxima de mensagens em processamento simultâneo
	MinIOEndpoint  string // endereço do MinIO (host:porta)
	MinIOAccessKey string // chave de acesso do MinIO
	MinIOSecretKey string // chave secreta do MinIO
	MinIOBucket    string // bucket onde os objetos estão armazenados
	MinIOUseSSL    bool   // indica se a conexão com o MinIO usa TLS
}

// Load lê as variáveis de ambiente e devolve a configuração preenchida.
func Load() (*Config, error) {
	dbDSN, err := requireEnv("DB_DSN")
	if err != nil {
		return nil, err
	}
	rabbitURL, err := requireEnv("RABBITMQ_URL")
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

	return &Config{
		DBDSN:          dbDSN,
		RabbitMQURL:    rabbitURL,
		QueueName:      getEnvOrDefault("QUEUE_NAME", "video.process"),
		PrefetchCount:  parseIntOrDefault("PREFETCH_COUNT", 5),
		MinIOEndpoint:  minioEndpoint,
		MinIOAccessKey: minioAccessKey,
		MinIOSecretKey: minioSecretKey,
		MinIOBucket:    getEnvOrDefault("MINIO_BUCKET", "videos"),
		MinIOUseSSL:    getEnvOrDefault("MINIO_USE_SSL", "false") == "true",
	}, nil
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("variável de ambiente obrigatória ausente: %s", key)
	}
	return v, nil
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
