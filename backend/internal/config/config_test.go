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
	t.Setenv("CORS_ORIGINS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(cfg.CORSOrigins) == 0 {
		t.Fatal("expected default CORS origins to be populated")
	}
}
