package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

const probeEmail = "lockout-probe@dashpoint.local"

func newLockoutTestApp(validPassword string) *fiber.App {
	app := fiber.New()
	app.Post("/login", AccountLockout(), func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "bad body")
		}
		if body.Password == validPassword {
			return c.JSON(fiber.Map{"ok": true})
		}
		return fiber.NewError(fiber.StatusUnauthorized, "invalid")
	})
	return app
}

func postLogin(t *testing.T, app *fiber.App, email, password string) *http.Response {
	t.Helper()
	req := httptest.NewRequest("POST", "/login", strings.NewReader(
		`{"email":"`+email+`","password":"`+password+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	return resp
}

func TestAccountLockoutBlocksAfterThresholdEvenWithValidCredentials(t *testing.T) {
	app := newLockoutTestApp("right-password")

	for i := 0; i < AccountLockoutMaxFailures; i++ {
		resp := postLogin(t, app, probeEmail, "wrong-"+string(rune('a'+i)))
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, resp.StatusCode)
		}
	}

	resp := postLogin(t, app, probeEmail, "right-password")
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("expected locked account to return 429 even with valid credentials, got %d", resp.StatusCode)
	}
	var body struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	raw, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		t.Fatalf("failed to read body: %v", rerr)
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("failed to parse 429 body: %v", err)
	}
	if body.Code != "ACCOUNT_LOCKED" {
		t.Fatalf("expected code ACCOUNT_LOCKED, got %q", body.Code)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Fatalf("expected Retry-After header on lockout response")
	}
	if !strings.Contains(body.Message, "Too many failed attempts") {
		t.Fatalf("unexpected message: %q", body.Message)
	}
}

func TestAccountLockoutIsolatesOtherAccounts(t *testing.T) {
	app := newLockoutTestApp("right-password")

	for i := 0; i < AccountLockoutMaxFailures; i++ {
		postLogin(t, app, probeEmail, "wrong")
	}

	resp := postLogin(t, app, "other-account@dashpoint.local", "right-password")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected unrelated account to authenticate normally (200), got %d", resp.StatusCode)
	}
}

func TestAccountLockoutResetsOnSuccessfulLogin(t *testing.T) {
	app := newLockoutTestApp("right-password")

	for i := 0; i < AccountLockoutMaxFailures-1; i++ {
		postLogin(t, app, probeEmail, "wrong")
	}
	if resp := postLogin(t, app, probeEmail, "right-password"); resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected successful login before threshold, got %d", resp.StatusCode)
	}

	for i := 0; i < AccountLockoutMaxFailures-1; i++ {
		resp := postLogin(t, app, probeEmail, "wrong")
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("post-reset attempt %d should still be 401, got %d", i+1, resp.StatusCode)
		}
	}
	if resp := postLogin(t, app, probeEmail, "right-password"); resp.StatusCode != fiber.StatusOK {
		t.Fatalf("counter should have reset on success; expected 200 after 4 fresh failures, got %d", resp.StatusCode)
	}
}

func TestAccountLockoutKeysPinLoginByUserID(t *testing.T) {
	app := fiber.New()
	app.Post("/pin-login", AccountLockout(), func(c *fiber.Ctx) error {
		var body struct {
			UserID string `json:"user_id"`
			Pin    string `json:"pin"`
		}
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "bad body")
		}
		if body.UserID == "user-abc" && body.Pin == "1234" {
			return c.JSON(fiber.Map{"ok": true})
		}
		return fiber.NewError(fiber.StatusUnauthorized, "invalid")
	})

	post := func(userID, pin string) *http.Response {
		req := httptest.NewRequest("POST", "/pin-login", strings.NewReader(
			`{"user_id":"`+userID+`","pin":"`+pin+`"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test returned error: %v", err)
		}
		return resp
	}

	for i := 0; i < AccountLockoutMaxFailures; i++ {
		post("user-abc", "0000")
	}
	if resp := post("user-abc", "1234"); resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("expected pin-login lockout -> 429, got %d", resp.StatusCode)
	}
}

func TestAccountLockoutExpiresAfterDuration(t *testing.T) {
	store := newAccountLockoutStore()
	store.nowFunc = func() time.Time { return time.Unix(1_700_000_000, 0) }

	key := "em:expired@dashpoint.local"
	for i := 0; i < AccountLockoutMaxFailures; i++ {
		store.recordFailure(key)
	}
	if _, locked := store.lockedFor(key); !locked {
		t.Fatalf("expected account to be locked immediately after threshold")
	}

	store.nowFunc = func() time.Time { return time.Unix(1_700_000_000, 0).Add(AccountLockoutDuration + time.Minute) }
	if _, locked := store.lockedFor(key); locked {
		t.Fatalf("expected lock to expire after duration")
	}
}

func TestAccountLockoutPassesThroughUnidentifiableRequests(t *testing.T) {
	app := fiber.New()
	app.Post("/login", AccountLockout(), func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "no identifier")
	})
	req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"nope":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected handler's own 400 for unidentifiable body, got %d", resp.StatusCode)
	}
}
