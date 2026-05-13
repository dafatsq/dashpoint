package middleware

import (
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Logger returns a middleware that logs HTTP requests
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		status := c.Response().StatusCode()

		// Log the request
		logEvent := log.Info()
		if status >= 400 {
			logEvent = log.Warn()
		}
		if status >= 500 {
			logEvent = log.Error()
		}

		logEvent.
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("duration", duration).
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

		c.Locals("requestid", requestID)
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

func GetRequestID(c *fiber.Ctx) string {
	if requestID, ok := c.Locals("requestid").(string); ok && requestID != "" {
		return requestID
	}
	if requestID, ok := c.Locals("request_id").(string); ok && requestID != "" {
		return requestID
	}
	return ""
}

// CORS returns a middleware that handles CORS
func CORS(allowedOrigins []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		applyCORSHeaders(c, allowedOrigins)
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
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
		c.Set("Vary", "Origin")
	}
}

func JSONError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"code":    code,
		"message": message,
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

				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "An unexpected error occurred",
				})
			}
		}()

		return c.Next()
	}
}
