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
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	LogLevel           slog.Level
	ShutdownTimeout    time.Duration
	// Để trống thì đăng nhập bằng Google tắt. Không có giá trị mặc định nào
	// dùng được, nên hai biến này không bắt buộc.
	GoogleClientID     string
	GoogleClientSecret string
	// Để trống thì hạn mức đếm trong bộ nhớ của từng tiến trình — chỉ đúng khi
	// chạy đúng một instance.
	RedisURL string
	// TrustedProxyHeader là header proxy phía trước tự đặt để nói IP người
	// dùng (ví dụ X-Forwarded-For). Rỗng: không tin header nào.
	TrustedProxyHeader string
	// BFFSharedSecret cho phép máy chủ web nói IP của người đang đăng nhập.
	// Rỗng thì tắt; đặt thì phải dài ít nhất 32 byte.
	BFFSharedSecret string
}

// minBFFSecretLen giống độ dài tối thiểu của JWT secret: đây cũng là một thứ
// cho phép tự khai danh tính (ở đây là IP), nên đoán được là mất lớp chặn.
const minBFFSecretLen = 32

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
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedisURL:           os.Getenv("REDIS_URL"),
		TrustedProxyHeader: os.Getenv("TRUSTED_PROXY_HEADER"),
		BFFSharedSecret:    os.Getenv("BFF_SHARED_SECRET"),
	}

	accessTTL, err := parseDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	cfg.AccessTokenTTL = accessTTL

	refreshTTL, err := parseDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	cfg.RefreshTokenTTL = refreshTTL

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("%w: DATABASE_URL", ErrMissingEnv)
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("%w: JWT_SECRET", ErrMissingEnv)
	}

	if cfg.BFFSharedSecret != "" && len(cfg.BFFSharedSecret) < minBFFSecretLen {
		return Config{}, fmt.Errorf("config: BFF_SHARED_SECRET must be at least %d bytes", minBFFSecretLen)
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

func parseDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("config: parse %s %q: %w", key, raw, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("config: %s phải lớn hơn 0", key)
	}
	return value, nil
}
