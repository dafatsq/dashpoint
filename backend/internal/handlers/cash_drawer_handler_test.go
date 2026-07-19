package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

type fakeCashDrawerStore struct {
	createFunc          func(context.Context, *models.CashDrawerOperation) error
	listByShiftFunc     func(context.Context, uuid.UUID) ([]models.CashDrawerOperation, error)
	getTotalsByShiftFun func(context.Context, uuid.UUID) (decimal.Decimal, decimal.Decimal, error)
}

func (f *fakeCashDrawerStore) Create(ctx context.Context, op *models.CashDrawerOperation) error {
	if f.createFunc != nil {
		return f.createFunc(ctx, op)
	}
	return nil
}
func (f *fakeCashDrawerStore) ListByShift(ctx context.Context, shiftID uuid.UUID) ([]models.CashDrawerOperation, error) {
	if f.listByShiftFunc != nil {
		return f.listByShiftFunc(ctx, shiftID)
	}
	return nil, nil
}
func (f *fakeCashDrawerStore) GetTotalsByShift(ctx context.Context, shiftID uuid.UUID) (decimal.Decimal, decimal.Decimal, error) {
	if f.getTotalsByShiftFun != nil {
		return f.getTotalsByShiftFun(ctx, shiftID)
	}
	return decimal.Zero, decimal.Zero, nil
}

func TestPayInRejectsMissingReason(t *testing.T) {
	userID := uuid.New()
	handler := NewCashDrawerHandler(&fakeCashDrawerStore{}, &fakeShiftLookup{
		shift: &models.Shift{ID: uuid.New(), OpenedBy: userID},
	})

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	})
	app.Post("/shifts/pay-in", handler.PayIn)

	req := httptest.NewRequest(http.MethodPost, "/shifts/pay-in", strings.NewReader(`{"amount":"10.00","reason":""}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
