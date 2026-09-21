// Package config loads process configuration from the environment.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// ErrMissingEnv is returned when a required environment variable is empty.
var ErrMissingEnv = errors.New("config: missing required environment variable")

// Config holds every value the api process needs to boot.
type Config struct {
	Env                string
	Port               string
	DatabaseURL        string
	JWTSecret          string
	CORSAllowedOrigins []string
	LogLevel           slog.Level
	ShutdownTimeout    time.Duration
}

// Load reads the configuration from the environment. Cloud Run injects PORT, so
// it is read as a string and never rewritten.
func Load() (Config, error) {
	cfg := Config{
		Env:                lookup("APP_ENV", "development"),
		Port:               lookup("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		CORSAllowedOrigins: splitList(lookup("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		ShutdownTimeout:    15 * time.Second,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("%w: DATABASE_URL", ErrMissingEnv)
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("%w: JWT_SECRET", ErrMissingEnv)
	}

	level, err := parseLogLevel(lookup("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	cfg.LogLevel = level

	return cfg, nil
}

func lookup(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func parseLogLevel(raw string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(raw))); err != nil {
		return 0, fmt.Errorf("config: parse LOG_LEVEL %q: %w", raw, err)
	}
	return level, nil
}

// splitList tách chuỗi phân tách bằng dấu phẩy và bỏ phần tử rỗng.
func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}
