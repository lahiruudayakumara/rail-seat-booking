// Package config loads and validates runtime configuration at the process boundary.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment        string
	APIAddress         string
	DatabaseURL        string
	AllowedOrigins     []string
	ReadHeaderTimeout  time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
	DatabaseTimeout    time.Duration
	MaxDatabaseConns   int32
	OpenAPIPath        string
	ManagementSecret   string
	ManagementTTL      time.Duration
	SeatHoldTTL        time.Duration
	AdminAPIKey        string
	OutboxPollInterval time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Environment:        value("APP_ENV", "development"),
		APIAddress:         ":" + value("API_PORT", "8080"),
		DatabaseURL:        value("DATABASE_URL", "postgres://rail_app:local-development-only@localhost:5432/rail_booking?sslmode=disable"),
		AllowedOrigins:     split(value("ALLOWED_ORIGINS", "http://localhost:3000")),
		ReadHeaderTimeout:  duration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
		ReadTimeout:        duration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:       duration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:        duration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout:    duration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		DatabaseTimeout:    duration("DATABASE_TIMEOUT", 3*time.Second),
		MaxDatabaseConns:   int32Value("DATABASE_MAX_CONNECTIONS", 20),
		OpenAPIPath:        value("OPENAPI_PATH", "docs/openapi.yaml"),
		ManagementSecret:   value("MANAGEMENT_TOKEN_SECRET", "local-development-management-secret-change-me"),
		ManagementTTL:      duration("MANAGEMENT_TOKEN_TTL", 30*time.Minute),
		SeatHoldTTL:        duration("SEAT_HOLD_TTL", 10*time.Minute),
		AdminAPIKey:        value("ADMIN_API_KEY", "local-development-admin-key-change-me"),
		OutboxPollInterval: duration("OUTBOX_POLL_INTERVAL", 2*time.Second),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return Config{}, fmt.Errorf("ALLOWED_ORIGINS must contain at least one origin")
	}
	if cfg.Environment == "production" && len(cfg.ManagementSecret) < 32 {
		return Config{}, fmt.Errorf("MANAGEMENT_TOKEN_SECRET must contain at least 32 characters in production")
	}
	if cfg.Environment == "production" && len(cfg.AdminAPIKey) < 32 {
		return Config{}, fmt.Errorf("ADMIN_API_KEY must contain at least 32 characters in production")
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
func split(value string) []string {
	result := []string{}
	for _, v := range strings.Split(value, ",") {
		if v = strings.TrimSpace(v); v != "" {
			result = append(result, v)
		}
	}
	return result
}
func duration(key string, fallback time.Duration) time.Duration {
	raw := value(key, "")
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
func int32Value(key string, fallback int32) int32 {
	raw := value(key, "")
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || parsed < 1 {
		return fallback
	}
	return int32(parsed)
}
