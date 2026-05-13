package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/middleware"
)

func productJSONError(c *fiber.Ctx, status int, code, message string) error {
	return middleware.JSONError(c, status, code, message)
}

func productInternalError(c *fiber.Ctx, err error, logMessage, clientMessage string) error {
	log.Error().Err(err).Msg(logMessage)
	return productJSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", clientMessage)
}

func parseUUIDParam(c *fiber.Ctx, name, code, message string) (uuid.UUID, error) {
	value := c.Params(name)
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, productJSONError(c, fiber.StatusBadRequest, code, message)
	}
	return id, nil
}

func parseDecimalField(raw string, fieldName string, allowNegative bool) (decimal.Decimal, error) {
	value, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid %s", fieldName)
	}
	if !allowNegative && value.LessThan(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("invalid %s", fieldName)
	}
	return value, nil
}

func parseOptionalDecimalField(raw *string, fieldName string, allowNegative bool) (*decimal.Decimal, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := parseDecimalField(*raw, fieldName, allowNegative)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseOptionalUUIDField(raw *string, fieldName string) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	if *raw == "" {
		return nil, nil
	}
	value, err := uuid.Parse(*raw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s", fieldName)
	}
	return &value, nil
}
