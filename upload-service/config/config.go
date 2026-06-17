// Package config loads the upload-service configuration from environment variables.
package config

import (
	"log"
	"os"
)

// Config holds all runtime configuration for the upload-service.
type Config struct {
	PostgresDSN    string // required
	RabbitMQURL    string // required
	MinIOEndpoint  string // required
	MinIOAccessKey string // required
	MinIOSecretKey string // required
	MinioBucket    string // default: "videos"
	MinIOUseSSL    bool   // default: false
	QueueName      string // default: "video.upload"
	UploadPort     string // default: "8082"
}

// Load reads environment variables and returns a populated Config.
// Calls log.Fatalf if any required variable is missing.
func Load() *Config {
	cfg := &Config{}

	cfg.PostgresDSN = requireEnv("POSTGRES_DSN")
	cfg.RabbitMQURL = requireEnv("RABBITMQ_URL")
	cfg.MinIOEndpoint = requireEnv("MINIO_ENDPOINT")
	cfg.MinIOAccessKey = requireEnv("MINIO_ACCESS_KEY")
	cfg.MinIOSecretKey = requireEnv("MINIO_SECRET_KEY")

	cfg.MinioBucket = getEnvOrDefault("MINIO_BUCKET", "videos")
	cfg.MinIOUseSSL = getEnvOrDefault("MINIO_USE_SSL", "false") == "true"
	cfg.QueueName = getEnvOrDefault("QUEUE_NAME", "video.upload")
	cfg.UploadPort = getEnvOrDefault("UPLOAD_PORT", "8082")

	return cfg
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
