package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	Env  string
	Port int

	DatabaseURL     string
	JWTSecret       string
	PhoneHashPepper string

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	ServerMode string
}

// Load reads configuration from the environment (and an optional .env file).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:             getEnv("APP_ENV", "development"),
		Port:            getEnvInt("PORT", 8080),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://sahayak:sahayak@localhost:5432/sahayak?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		PhoneHashPepper: getEnv("PHONE_HASH_PEPPER", ""),
		AccessTokenTTL:  time.Duration(getEnvInt("ACCESS_TOKEN_TTL_SEC", 900)) * time.Second,
		RefreshTokenTTL: time.Duration(getEnvInt("REFRESH_TOKEN_TTL_SEC", 2592000)) * time.Second,
		ServerMode:      strings.ToLower(getEnv("SERVER_MODE", "all")),
	}

	if cfg.JWTSecret == "" {
		if cfg.Env == "production" {
			return nil, fmt.Errorf("JWT_SECRET must be set in production")
		}
		cfg.JWTSecret = "dev-insecure-jwt-secret-change-me"
	}
	if cfg.PhoneHashPepper == "" {
		if cfg.Env == "production" {
			return nil, fmt.Errorf("PHONE_HASH_PEPPER must be set in production")
		}
		cfg.PhoneHashPepper = "dev-insecure-pepper-change-me"
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
