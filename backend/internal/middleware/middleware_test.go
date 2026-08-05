package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
	if got := resp.Header.Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentials header, got %q", got)
	}
}

func TestCORSAllowsWailsOriginWhenConfigured(t *testing.T) {
	app := fiber.New()
	app.Use(CORS([]string{"wails://wails", "http://wails.localhost"}))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for _, origin := range []string{"wails://wails", "http://wails.localhost"} {
		t.Run(origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Origin", origin)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test returned error: %v", err)
			}

			if got := resp.Header.Get("Access-Control-Allow-Origin"); got != origin {
				t.Fatalf("expected Wails allow origin header, got %q", got)
			}
			if got := resp.Header.Get("Access-Control-Allow-Credentials"); got != "true" {
				t.Fatalf("expected credentials header, got %q", got)
			}
		})
	}
}

func TestCORSPreflightAllowsSSEHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(CORS([]string{"http://localhost:3000"}))
	app.Options("/events/subscribe", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/events/subscribe", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set(
		"Access-Control-Request-Headers",
		"authorization,cache-control,last-event-id",
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected status 204 for preflight, got %d", resp.StatusCode)
	}

	allowHeaders := resp.Header.Get("Access-Control-Allow-Headers")
	for _, required := range []string{"Authorization", "Cache-Control", "Last-Event-ID"} {
		if !strings.Contains(allowHeaders, required) {
			t.Fatalf("expected allow headers to include %q, got %q", required, allowHeaders)
		}
	}
}

func TestJSONErrorIncludesRequestID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/", func(c *fiber.Ctx) error {
		return JSONError(c, fiber.StatusForbidden, "FORBIDDEN", "Denied")
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

func TestSecureHeadersSetsExpectedHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(SecureHeaders())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff header, got %q", got)
	}
	if got := resp.Header.Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected DENY frame option, got %q", got)
	}
	if got := resp.Header.Get("X-XSS-Protection"); got != "1; mode=block" {
		t.Fatalf("expected XSS protection header, got %q", got)
	}
	if got := resp.Header.Get("Referrer-Policy"); got != "no-referrer" {
		t.Fatalf("expected no-referrer policy, got %q", got)
	}
	if got := resp.Header.Get("Permissions-Policy"); !strings.Contains(got, "geolocation=()") {
		t.Fatalf("expected permissions policy, got %q", got)
	}
}

func TestAuthRateLimitBlocksAfterBurst(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Post("/auth/login", AuthRateLimit(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for attempt := 0; attempt < 10; attempt++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("attempt %d returned error: %v", attempt+1, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected attempt %d to pass, got %d", attempt+1, resp.StatusCode)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("limited request returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("expected 429 after limit, got %d", resp.StatusCode)
	}
}
