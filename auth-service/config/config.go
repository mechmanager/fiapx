// Package config loads the auth-service configuration from environment variables.
package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the auth-service.
type Config struct {
	PostgresDSN        string // required: PostgreSQL connection string
	JWTSecret          string // required: HMAC secret for JWT signing
	JWTExpirationHours int    // default: 24
	AuthPort           string // default: "8081"
}

// Load reads environment variables and returns a populated Config.
// Calls log.Fatalf if any required variable is missing.
func Load() *Config {
	cfg := &Config{}

	cfg.PostgresDSN = requireEnv("POSTGRES_DSN")
	cfg.JWTSecret = requireEnv("JWT_SECRET")

	cfg.JWTExpirationHours = 24
	if raw := os.Getenv("JWT_EXPIRATION_HOURS"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			log.Fatalf("invalid JWT_EXPIRATION_HOURS: %v", err)
		}
		cfg.JWTExpirationHours = v
	}

	cfg.AuthPort = "8081"
	if p := os.Getenv("AUTH_PORT"); p != "" {
		cfg.AuthPort = p
	}

	return cfg
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}
