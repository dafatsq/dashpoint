package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const staleSubmitMessage = "This record changed after you opened it. Refresh and try again."
const staleShiftMessage = "The open shift changed after you opened this form. Refresh and try again."

func expectedUpdatedAtFromQuery(c *fiber.Ctx) *string {
	value := c.Query("expected_updated_at")
	if value == "" {
		return nil
	}
	return &value
}

func isStaleSubmit(expectedUpdatedAt *string, actualUpdatedAt time.Time) (bool, error) {
	if expectedUpdatedAt == nil || *expectedUpdatedAt == "" {
		return false, nil
	}

	expected, err := time.Parse(time.RFC3339Nano, *expectedUpdatedAt)
	if err != nil {
		return false, err
	}

	return !expected.UTC().Truncate(time.Second).Equal(actualUpdatedAt.UTC().Truncate(time.Second)), nil
}

func parseExpectedUUID(expectedID *string) (*uuid.UUID, error) {
	if expectedID == nil || *expectedID == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(*expectedID)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
