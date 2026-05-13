package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/middleware"
)

type apiError struct {
	status  int
	code    string
	message string
}

func (e *apiError) Error() string {
	return e.message
}

func respondAPIError(c *fiber.Ctx, err error) error {
	apiErr, ok := err.(*apiError)
	if !ok {
		return err
	}
	return middleware.JSONError(c, apiErr.status, apiErr.code, apiErr.message)
}

func saleInvalidRequest(c *fiber.Ctx) error {
	return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
}

func saleValidationError(c *fiber.Ctx, err error) error {
	return middleware.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
}

func saleInternalError(c *fiber.Ctx, message string) error {
	return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func saleParamUUID(c *fiber.Ctx, name, code, message string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(name))
	if err != nil {
		return uuid.Nil, &apiError{status: fiber.StatusBadRequest, code: code, message: message}
	}
	return id, nil
}

func saleRequiredReason(reason string) error {
	if reason == "" {
		return fmt.Errorf("reason is required")
	}
	return nil
}
