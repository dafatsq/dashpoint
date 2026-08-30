package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func postShiftJSON(t *testing.T, parse func(c *fiber.Ctx) error, body string) int {
	t.Helper()
	app := fiber.New()
	app.Post("/", parse)
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	return resp.StatusCode
}

// parseStartShiftJSON adapts the three-value parse signature to a handler.
func parseStartShiftJSON(c *fiber.Ctx) error {
	_, _, err := parseStartShiftRequest(c)
	return err
}

// parseCloseShiftJSON adapts the three-value parse signature to a handler.
func parseCloseShiftJSON(c *fiber.Ctx) error {
	_, _, err := parseCloseShiftRequest(c)
	return err
}

func TestStartShiftRejectsNegativeOpeningCash(t *testing.T) {
	status := postShiftJSON(t, parseStartShiftJSON, `{"opening_cash":"-50.00"}`)
	if status != fiber.StatusBadRequest {
		t.Fatalf("expected negative opening cash to be rejected (400), got %d", status)
	}
}

func TestStartShiftAcceptsNonNegativeOpeningCash(t *testing.T) {
	if status := postShiftJSON(t, parseStartShiftJSON, `{"opening_cash":"100.00"}`); status != fiber.StatusOK {
		t.Fatalf("expected valid opening cash to pass (200), got %d", status)
	}
	if status := postShiftJSON(t, parseStartShiftJSON, `{"opening_cash":"0"}`); status != fiber.StatusOK {
		t.Fatalf("expected zero opening cash to pass (200), got %d", status)
	}
}

func TestCloseShiftRejectsNegativeClosingCash(t *testing.T) {
	status := postShiftJSON(t, parseCloseShiftJSON, `{"closing_cash":"-1.00"}`)
	if status != fiber.StatusBadRequest {
		t.Fatalf("expected negative closing cash to be rejected (400), got %d", status)
	}
}

func TestCloseShiftAcceptsNonNegativeClosingCash(t *testing.T) {
	if status := postShiftJSON(t, parseCloseShiftJSON, `{"closing_cash":"250.75"}`); status != fiber.StatusOK {
		t.Fatalf("expected valid closing cash to pass (200), got %d", status)
	}
}
