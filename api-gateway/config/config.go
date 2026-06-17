package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GatewayPort      string
	JWTSecret        string
	AuthServiceURL   string
	UploadServiceURL string
	StatusServiceURL string
	RateLimitRPS     int
}

func Load() (*Config, error) {
	jwtSecret, err := requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}
	return &Config{
		GatewayPort:      getEnvOrDefault("GATEWAY_PORT", "8080"),
		JWTSecret:        jwtSecret,
		AuthServiceURL:   getEnvOrDefault("AUTH_SERVICE_URL", "http://auth-service:8081"),
		UploadServiceURL: getEnvOrDefault("UPLOAD_SERVICE_URL", "http://upload-service:8082"),
		StatusServiceURL: getEnvOrDefault("STATUS_SERVICE_URL", "http://status-service:8083"),
		RateLimitRPS:     parseIntOrDefault("RATE_LIMIT_RPS", 100),
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
