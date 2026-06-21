// Package config loads service configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the runtime configuration of the Community service.
type Config struct {
	HTTPPort         int
	DatabaseURL      string
	JWTPublicKeyPath string
	LogLevel         string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:         getEnvInt("HTTP_PORT", 8002),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://tinta:tinta_dev_pass@localhost:5432/tinta?sslmode=disable&search_path=community"),
		JWTPublicKeyPath: getEnv("JWT_PUBLIC_KEY_PATH", "keys/jwt_public.pem"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
