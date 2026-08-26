package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type healthChecker interface {
	HealthCheck(context.Context) (bool, error)
}

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db healthChecker
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db healthChecker) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// Check handles GET /api/v1/health
func (h *HealthHandler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]string),
	}

	// Check database connection. Driver error details stay in server logs
	// only — this endpoint is public and must not leak database internals.
	if healthy, err := h.db.HealthCheck(ctx); healthy {
		response.Services["database"] = "connected"
	} else {
		response.Status = "degraded"
		response.Services["database"] = "disconnected"
		if err != nil {
			log.Error().Err(err).Msg("Health check: database unavailable")
		}
	}

	statusCode := fiber.StatusOK
	if response.Status != "ok" {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(response)
}

// Ping handles GET /api/v1/ping - lightweight health check
func (h *HealthHandler) Ping(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "pong",
	})
}
