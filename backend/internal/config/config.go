package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	defaultJWTExpiryMinutes       = 15
	maxProductionJWTExpiryMinutes = 15
	defaultRefreshExpiryHours     = 168
	maxRefreshExpiryHours         = 168
	minProductionJWTSecretLength  = 32
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port        string
	Environment string
	CORSOrigins []string

	// Database
	DatabaseURL string

	// JWT
	JWTSecret          string
	JWTExpiryMinutes   int
	RefreshExpiryHours int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		CORSOrigins: getEnvList("CORS_ORIGINS", []string{"http://localhost:3000"}),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
	}

	jwtExpiry, err := getPositiveIntEnv("JWT_EXPIRY_MINUTES", defaultJWTExpiryMinutes)
	if err != nil {
		return nil, err
	}
	cfg.JWTExpiryMinutes = jwtExpiry

	refreshExpiry, err := getPositiveIntEnv("REFRESH_EXPIRY_HOURS", defaultRefreshExpiryHours)
	if err != nil {
		return nil, err
	}
	cfg.RefreshExpiryHours = refreshExpiry

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.RefreshExpiryHours > maxRefreshExpiryHours {
		return nil, fmt.Errorf("REFRESH_EXPIRY_HOURS must be %d or less", maxRefreshExpiryHours)
	}
	if !cfg.IsDevelopment() && !hasExplicitEnv("CORS_ORIGINS") {
		return nil, fmt.Errorf("CORS_ORIGINS is required outside development")
	}
	if !cfg.IsDevelopment() && hasWildcardCORSOrigin(cfg.CORSOrigins) {
		return nil, fmt.Errorf("CORS_ORIGINS cannot contain wildcard origins outside development")
	}
	if cfg.IsProduction() && len(cfg.JWTSecret) < minProductionJWTSecretLength {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d characters in production", minProductionJWTSecretLength)
	}
	if cfg.IsProduction() && cfg.JWTExpiryMinutes > maxProductionJWTExpiryMinutes {
		return nil, fmt.Errorf("JWT_EXPIRY_MINUTES must be %d or less in production", maxProductionJWTExpiryMinutes)
	}
	if cfg.IsProduction() && !hasProductionDatabaseSSL(cfg.DatabaseURL) {
		return nil, fmt.Errorf("DATABASE_URL must use sslmode=require, verify-ca, or verify-full in production")
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func hasExplicitEnv(key string) bool {
	value, exists := os.LookupEnv(key)
	return exists && strings.TrimSpace(value) != ""
}

func getPositiveIntEnv(key string, defaultValue int) (int, error) {
	raw, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", key)
	}

	return value, nil
}

func getEnvList(key string, defaultValues []string) []string {
	value, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return append([]string(nil), defaultValues...)
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return append([]string(nil), defaultValues...)
	}

	return values
}

func hasWildcardCORSOrigin(origins []string) bool {
	for _, origin := range origins {
		if strings.TrimSpace(origin) == "*" {
			return true
		}
	}
	return false
}

func hasProductionDatabaseSSL(databaseURL string) bool {
	sslMode := databaseSSLMode(databaseURL)
	return sslMode == "require" || sslMode == "verify-ca" || sslMode == "verify-full"
}

func databaseSSLMode(databaseURL string) string {
	if parsed, err := url.Parse(databaseURL); err == nil && parsed.Query().Get("sslmode") != "" {
		return strings.ToLower(strings.TrimSpace(parsed.Query().Get("sslmode")))
	}

	for _, part := range strings.Fields(databaseURL) {
		key, value, found := strings.Cut(part, "=")
		if found && strings.EqualFold(strings.TrimSpace(key), "sslmode") {
			return strings.ToLower(strings.TrimSpace(value))
		}
	}

	return ""
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
