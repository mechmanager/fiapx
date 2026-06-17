// Package config carrega as configurações do notification-service a partir de variáveis de ambiente.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config agrupa todas as configurações necessárias para o notification-service.
type Config struct {
	PostgresDSN       string
	RabbitMQURL       string
	NotificationQueue string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPass          string
	SMTPFrom          string
	MaxRetries        int
}

// Load lê as variáveis de ambiente e devolve a configuração preenchida.
func Load() (*Config, error) {
	postgresDSN, err := requireEnv("POSTGRES_DSN")
	if err != nil {
		return nil, err
	}

	rabbitURL, err := requireEnv("RABBITMQ_URL")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		PostgresDSN:       postgresDSN,
		RabbitMQURL:       rabbitURL,
		NotificationQueue: getEnvOrDefault("NOTIFICATION_QUEUE", "notification"),
		SMTPHost:          getEnvOrDefault("SMTP_HOST", ""),
		SMTPPort:          getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUser:          getEnvOrDefault("SMTP_USER", ""),
		SMTPPass:          getEnvOrDefault("SMTP_PASS", ""),
		SMTPFrom:          getEnvOrDefault("SMTP_FROM", "noreply@fiapx.com"),
		MaxRetries:        parseIntOrDefault("MAX_RETRIES", 3),
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

func parseIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
