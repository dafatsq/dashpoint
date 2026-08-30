package handlers

import (
	"net"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func probeSecureScheme(trustedProxies []string) string {
	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          trustedProxies,
	})

	var fctx fasthttp.RequestCtx
	var freq fasthttp.Request
	freq.Header.SetMethod("GET")
	freq.SetRequestURI("/")
	freq.Header.Set("X-Forwarded-Proto", "https")
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 52341}
	fctx.Init(&freq, remote, nil)

	c := app.AcquireCtx(&fctx)
	defer app.ReleaseCtx(c)
	if isSecureRequest(c) {
		return "secure"
	}
	return "plain"
}

// A client-supplied X-Forwarded-Proto may only decide the cookie Secure flag
// when it comes from a configured (trusted) reverse proxy.
func TestIsSecureRequestIgnoresUntrustedForwardedProto(t *testing.T) {
	if got := probeSecureScheme([]string{"127.0.0.1"}); got != "secure" {
		t.Fatalf("trusted proxy https header should mark request secure, got %q", got)
	}
	if got := probeSecureScheme(nil); got != "plain" {
		t.Fatalf("untrusted https header must not mark request secure, got %q", got)
	}
}
