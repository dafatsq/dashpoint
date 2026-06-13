package database

import "testing"

func TestNewPoolConfigSetsStatementTimeout(t *testing.T) {
	cfg, err := newPoolConfig("postgres://postgres:postgres@localhost:5432/dashpoint?sslmode=disable")
	if err != nil {
		t.Fatalf("newPoolConfig returned error: %v", err)
	}

	if got := cfg.ConnConfig.RuntimeParams["statement_timeout"]; got != defaultStatementTimeout {
		t.Fatalf("expected statement_timeout %q, got %q", defaultStatementTimeout, got)
	}
}

func TestNewPoolConfigPreservesExplicitStatementTimeout(t *testing.T) {
	cfg, err := newPoolConfig("postgres://postgres:postgres@localhost:5432/dashpoint?sslmode=disable&statement_timeout=45000")
	if err != nil {
		t.Fatalf("newPoolConfig returned error: %v", err)
	}

	if got := cfg.ConnConfig.RuntimeParams["statement_timeout"]; got != "45000" {
		t.Fatalf("expected statement_timeout 45000, got %q", got)
	}
}
