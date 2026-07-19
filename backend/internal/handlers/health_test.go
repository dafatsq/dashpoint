package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type stubHealthChecker struct {
	healthy bool
	err     error
}

func (s stubHealthChecker) HealthCheck(context.Context) (bool, error) {
	return s.healthy, s.err
}

func TestHealthCheckReturnsOKWhenDatabaseIsHealthy(t *testing.T) {
	app := fiber.New()
	handler := NewHealthHandler(stubHealthChecker{healthy: true})
	app.Get("/health", handler.Check)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/health", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHealthCheckReturnsServiceUnavailableWhenDatabaseIsDown(t *testing.T) {
	app := fiber.New()
	handler := NewHealthHandler(stubHealthChecker{healthy: false})
	app.Get("/health", handler.Check)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/health", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", resp.StatusCode)
	}

	var payload HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.Status != "degraded" {
		t.Fatalf("expected degraded status, got %q", payload.Status)
	}
}
