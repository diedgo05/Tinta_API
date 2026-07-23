package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort           int
	DatabaseURL        string
	RedisURL           string
	JWTPrivateKeyPath  string
	JWTPublicKeyPath   string
	JWTAccessTTL       time.Duration
	JWTRefreshTTL      time.Duration
	LogLevel           string

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:          getEnvInt("HTTP_PORT", 8001),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://tinta:tinta_dev_pass@localhost:5432/tinta?sslmode=disable&search_path=identity"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "keys/jwt_private.pem"),
		JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "keys/jwt_public.pem"),
		JWTAccessTTL:      time.Duration(getEnvInt("JWT_ACCESS_TTL_MIN", 15)) * time.Minute,
		JWTRefreshTTL:     time.Duration(getEnvInt("JWT_REFRESH_TTL_HOURS", 168)) * time.Hour,
		LogLevel:          getEnv("LOG_LEVEL", "info"),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", "tintaappmovil@gmail.com"),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "tintaappmovil@gmail.com"),
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