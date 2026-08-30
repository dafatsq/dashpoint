package middleware

import (
	"fmt"
	"net"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func probeIPAndProtocol(trustedProxies []string) string {
	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          trustedProxies,
		ProxyHeader:             "X-Forwarded-For",
	})

	var fctx fasthttp.RequestCtx
	var freq fasthttp.Request
	freq.Header.SetMethod("GET")
	freq.SetRequestURI("/")
	freq.Header.Set("X-Forwarded-For", "203.0.113.7")
	freq.Header.Set("X-Forwarded-Proto", "https")
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 52341}
	fctx.Init(&freq, remote, nil)

	c := app.AcquireCtx(&fctx)
	defer app.ReleaseCtx(c)
	return fmt.Sprintf("%s|%s", c.IP(), c.Protocol())
}

// Loopback connections count as a local reverse proxy when allowlisted; an
// empty allowlist simulates direct exposure where forwarding headers must be
// ignored entirely.
func TestForwardedHeadersHonoredOnlyFromTrustedProxy(t *testing.T) {
	if got := probeIPAndProtocol([]string{"127.0.0.1"}); got != "203.0.113.7|https" {
		t.Fatalf("trusted proxy: want forwarded client/proto, got %q", got)
	}
	if got := probeIPAndProtocol(nil); got != "127.0.0.1|http" {
		t.Fatalf("untrusted headers must be ignored: got %q", got)
	}
}

func TestSecureHeadersIgnoresSpoofedForwardedProto(t *testing.T) {
	app := fiber.New(fiber.Config{
		// Mirror the production server config: without the trusted-proxy
		// check, fiber's Protocol() trusts X-Forwarded-Proto unconditionally.
		EnableTrustedProxyCheck: true,
		ProxyHeader:             "X-Forwarded-Proto",
	})
	app.Use(SecureHeaders())
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.Header.Get("Strict-Transport-Security") != "" {
		t.Fatalf("spoofed X-Forwarded-Proto must not set HSTS on a plain-HTTP response")
	}
}
