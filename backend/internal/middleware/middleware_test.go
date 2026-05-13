package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestIDGeneratesHeaderAndContextValue(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/", func(c *fiber.Ctx) error {
		requestID, _ := c.Locals("requestid").(string)
		if requestID == "" {
			t.Fatal("expected request ID in locals")
		}
		return c.SendString(requestID)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if got := resp.Header.Get("X-Request-ID"); got == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	app := fiber.New()
	app.Use(CORS([]string{"http://localhost:3000"}))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected allow origin header to echo configured origin, got %q", got)
	}
}
