package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
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
			Str("user_agent", c.Get("User-Agent")).
			Msg("HTTP request")

		return err
	}
}

// RequestID adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID already exists
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = c.GetRespHeader("X-Request-ID")
		}

		// Set request ID in response header
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

// CORS returns a middleware that handles CORS
func CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
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
