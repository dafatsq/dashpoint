package middleware

import (
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	requestIDLocalKey       = "request_id"
	legacyRequestIDLocalKey = "requestid"
	corsAllowMethods        = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowHeaders        = "Origin, Content-Type, Accept, Authorization, Cache-Control, Last-Event-ID, X-Request-ID"
)

// Logger returns a middleware that logs HTTP requests
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()

		requestLogger(status).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("duration", time.Since(start)).
			Str("ip", c.IP()).
			Str("request_id", GetRequestID(c)).
			Str("user_agent", c.Get("User-Agent")).
			Msg("HTTP request")

		return err
	}
}

// RequestID adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Locals(legacyRequestIDLocalKey, requestID)
		c.Locals(requestIDLocalKey, requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals(legacyRequestIDLocalKey).(string); ok && requestID != "" {
		return requestID
	}
	if requestID, ok := c.Locals(requestIDLocalKey).(string); ok && requestID != "" {
		return requestID
	}
	return ""
}

// CORS returns a middleware that handles CORS
func CORS(allowedOrigins []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		applyCORSHeaders(c, allowedOrigins)
		c.Set("Access-Control-Allow-Methods", corsAllowMethods)
		c.Set("Access-Control-Allow-Headers", corsAllowHeaders)
		c.Set("Access-Control-Max-Age", "86400")

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

func SecureHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		if c.Protocol() == "https" || c.Get("X-Forwarded-Proto") == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		return c.Next()
	}
}

func AuthRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,
		Expiration: 15 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return JSONError(
				c,
				fiber.StatusTooManyRequests,
				"RATE_LIMITED",
				"Too many authentication attempts. Please try again later.",
			)
		},
	})
}

func ApplyCORSHeaders(c *fiber.Ctx, allowedOrigins []string) {
	applyCORSHeaders(c, allowedOrigins)
}

func applyCORSHeaders(c *fiber.Ctx, allowedOrigins []string) {
	origin := c.Get("Origin")
	if origin == "" {
		return
	}

	if slices.Contains(allowedOrigins, origin) {
		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Vary", "Origin")
	}
}

func JSONError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"code":       code,
		"message":    message,
		"request_id": GetRequestID(c),
	})
}

// Recover returns a middleware that recovers from panics
func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Interface("panic", r).
					Str("path", c.Path()).
					Msg("Recovered from panic")

				_ = JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
			}
		}()

		return c.Next()
	}
}

func requestLogger(status int) *zerolog.Event {
	switch {
	case status >= 500:
		return log.Error()
	case status >= 400:
		return log.Warn()
	default:
		return log.Info()
	}
}
