package config

import (
	"testing"
)

func TestLoadParsesCORSOrigins(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("CORS_ORIGINS", "http://localhost:3000,https://dashpoint.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(cfg.CORSOrigins) != 2 {
		t.Fatalf("expected 2 CORS origins, got %d", len(cfg.CORSOrigins))
	}

	if cfg.CORSOrigins[0] != "http://localhost:3000" {
		t.Fatalf("expected first origin to be localhost, got %q", cfg.CORSOrigins[0])
	}
}

func TestLoadDefaultsCORSOriginsForDevelopment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("CORS_ORIGINS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(cfg.CORSOrigins) == 0 {
		t.Fatal("expected default CORS origins to be populated")
	}
}

func TestLoadRequiresCorsOriginsOutsideDevelopment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint?sslmode=require")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("CORS_ORIGINS", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to fail when CORS_ORIGINS is missing outside development")
	}
}

func TestLoadDefaultsJWTExpiryToFifteenMinutes(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint")
	t.Setenv("JWT_SECRET", "super-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.JWTExpiryMinutes != 15 {
		t.Fatalf("expected default JWT expiry 15, got %d", cfg.JWTExpiryMinutes)
	}
}

func TestLoadRejectsWeakProductionJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint?sslmode=require")
	t.Setenv("JWT_SECRET", "short-secret")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("CORS_ORIGINS", "https://dashpoint.example.com")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to reject weak production JWT secret")
	}
}

func TestLoadRejectsLongProductionJWTExpiry(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint?sslmode=require")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("CORS_ORIGINS", "https://dashpoint.example.com")
	t.Setenv("JWT_EXPIRY_MINUTES", "60")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to reject long production JWT expiry")
	}
}

func TestLoadRejectsInvalidJWTExpiry(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("JWT_EXPIRY_MINUTES", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to reject invalid JWT expiry")
	}
}

func TestLoadRejectsLongRefreshExpiry(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("REFRESH_EXPIRY_HOURS", "169")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to reject refresh expiry over seven days")
	}
}

func TestLoadRejectsWildcardCORSOutsideDevelopment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint?sslmode=require")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("CORS_ORIGINS", "*")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to reject wildcard CORS outside development")
	}
}

func TestLoadRejectsInsecureProductionDatabaseSSLMode(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint?sslmode=disable")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("CORS_ORIGINS", "https://dashpoint.example.com")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to reject insecure production database SSL mode")
	}
}

func TestLoadAcceptsProductionDatabaseVerifyFullSSLMode(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dashpoint?sslmode=verify-full")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("CORS_ORIGINS", "https://dashpoint.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.JWTExpiryMinutes != 15 {
		t.Fatalf("expected default JWT expiry 15, got %d", cfg.JWTExpiryMinutes)
	}
}
