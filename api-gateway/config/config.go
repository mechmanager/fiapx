// Package config carrega as configurações do serviço a partir de variáveis de ambiente.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config agrupa todas as configurações necessárias para o api-gateway.
type Config struct {
	APIPort            string // porta HTTP em que a API escuta
	DBDSN              string // string de conexão com o PostgreSQL
	RabbitMQURL        string // URL de conexão com o RabbitMQ
	QueueName          string // nome da fila onde os vídeos são publicados
	MinIOEndpoint      string // endereço do MinIO (host:porta)
	MinIOAccessKey     string // chave de acesso do MinIO
	MinIOSecretKey     string // chave secreta do MinIO
	MinIOBucket        string // bucket onde os objetos são armazenados
	MinIOUseSSL        bool   // indica se a conexão com o MinIO usa TLS
	JWTSecret          string // segredo usado para assinar os tokens JWT
	JWTExpirationHours int    // validade do token em horas
}

// Load lê as variáveis de ambiente e devolve a configuração preenchida.
// Retorna erro quando uma variável obrigatória está ausente.
func Load() (*Config, error) {
	// Carrega as variáveis obrigatórias, acumulando o primeiro erro encontrado.
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

	jwtSecret, err := requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	// Monta a configuração aplicando valores padrão às variáveis opcionais.
	cfg := &Config{
		APIPort:            getEnvOrDefault("API_PORT", "8080"),
		DBDSN:              dbDSN,
		RabbitMQURL:        rabbitURL,
		QueueName:          getEnvOrDefault("QUEUE_NAME", "video.process"),
		MinIOEndpoint:      minioEndpoint,
		MinIOAccessKey:     minioAccessKey,
		MinIOSecretKey:     minioSecretKey,
		MinIOBucket:        getEnvOrDefault("MINIO_BUCKET", "videos"),
		MinIOUseSSL:        getEnvOrDefault("MINIO_USE_SSL", "false") == "true",
		JWTSecret:          jwtSecret,
		JWTExpirationHours: parseIntOrDefault("JWT_EXPIRATION_HOURS", 24),
	}

	return cfg, nil
}

// requireEnv devolve o valor da variável ou um erro caso ela esteja vazia.
func requireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("variável de ambiente obrigatória ausente: %s", key)
	}
	return value, nil
}

// getEnvOrDefault devolve o valor da variável ou o padrão informado.
func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// parseIntOrDefault converte a variável para inteiro ou devolve o padrão.
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
