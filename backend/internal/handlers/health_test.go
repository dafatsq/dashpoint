package handlers

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type fakeHealthChecker struct {
	healthy bool
	err     error
}

func (f *fakeHealthChecker) HealthCheck(context.Context) (bool, error) {
	return f.healthy, f.err
}

func TestHealthDoesNotLeakDatabaseErrorDetails(t *testing.T) {
	handler := &HealthHandler{db: &fakeHealthChecker{
		healthy: false,
		err:     errors.New(`pq: connection refused: 10.0.0.5:5432 SQLSTATE 08001`),
	}}

	app := fiber.New()
	app.Get("/health", handler.Check)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	raw, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		t.Fatalf("read body: %v", rerr)
	}
	text := string(raw)
	if strings.Contains(text, "database_error") {
		t.Fatalf("health response leaked database_error field: %s", raw)
	}
	if strings.Contains(text, "pq:") || strings.Contains(text, "SQLSTATE") || strings.Contains(text, "connection refused") {
		t.Fatalf("health response leaked driver internals: %s", raw)
	}
	if !strings.Contains(text, `"disconnected"`) {
		t.Fatalf("expected degraded database status, got %s", raw)
	}
}
