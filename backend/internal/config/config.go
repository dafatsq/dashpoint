package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port        string
	Environment string

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
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
	}

	// Parse JWT expiry
	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_MINUTES", "60"))
	if err != nil {
		jwtExpiry = 60
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

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
