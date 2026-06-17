package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBDSN             string
	RabbitMQURL       string
	QueueName         string
	NotificationQueue string
	PrefetchCount     int
	MaxRetries        int
	MinIOEndpoint     string
	MinIOAccessKey    string
	MinIOSecretKey    string
	MinIOBucket       string
	MinIOUseSSL       bool
}

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
		DBDSN:             dbDSN,
		RabbitMQURL:       rabbitURL,
		QueueName:         getEnvOrDefault("QUEUE_NAME", "video.upload"),
		NotificationQueue: getEnvOrDefault("NOTIFICATION_QUEUE", "notification"),
		PrefetchCount:     parseIntOrDefault("PREFETCH_COUNT", 5),
		MaxRetries:        parseIntOrDefault("MAX_RETRIES", 3),
		MinIOEndpoint:     minioEndpoint,
		MinIOAccessKey:    minioAccessKey,
		MinIOSecretKey:    minioSecretKey,
		MinIOBucket:       getEnvOrDefault("MINIO_BUCKET", "videos"),
		MinIOUseSSL:       getEnvOrDefault("MINIO_USE_SSL", "false") == "true",
	}, nil
}

func requireEnv(key string) (string, error) {
	if v := os.Getenv(key); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("variável de ambiente obrigatória ausente: %s", key)
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
