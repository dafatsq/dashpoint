package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	defaultJWTExpiryMinutes       = 15
	maxProductionJWTExpiryMinutes = 15
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

	// Parse JWT expiry
	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_MINUTES", strconv.Itoa(defaultJWTExpiryMinutes)))
	if err != nil {
		jwtExpiry = defaultJWTExpiryMinutes
	}
	cfg.JWTExpiryMinutes = jwtExpiry

	refreshExpiry, err := strconv.Atoi(getEnv("REFRESH_EXPIRY_HOURS", "168"))
	if err != nil {
		refreshExpiry = 168 // 7 days
	}
	cfg.RefreshExpiryHours = refreshExpiry

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.JWTExpiryMinutes <= 0 {
		return nil, fmt.Errorf("JWT_EXPIRY_MINUTES must be greater than 0")
	}
	if !cfg.IsDevelopment() && !hasExplicitEnv("CORS_ORIGINS") {
		return nil, fmt.Errorf("CORS_ORIGINS is required outside development")
	}
	if cfg.IsProduction() && len(cfg.JWTSecret) < minProductionJWTSecretLength {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d characters in production", minProductionJWTSecretLength)
	}
	if cfg.IsProduction() && cfg.JWTExpiryMinutes > maxProductionJWTExpiryMinutes {
		return nil, fmt.Errorf("JWT_EXPIRY_MINUTES must be %d or less in production", maxProductionJWTExpiryMinutes)
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

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
