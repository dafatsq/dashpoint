package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type fakeShiftStore struct {
	createFunc                func(context.Context, *models.Shift) error
	getByIDFunc               func(context.Context, uuid.UUID) (*models.Shift, error)
	getOpenShiftByEmployeeFun func(context.Context, uuid.UUID) (*models.Shift, error)
	closeShiftFunc            func(context.Context, uuid.UUID, decimal.Decimal, *string, uuid.UUID) error
	listFunc                  func(context.Context, *repository.ShiftFilter) ([]models.Shift, int, error)
}

func (f *fakeShiftStore) Create(ctx context.Context, shift *models.Shift) error {
	if f.createFunc != nil {
		return f.createFunc(ctx, shift)
	}
	return nil
}
func (f *fakeShiftStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Shift, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return nil, nil
}
func (f *fakeShiftStore) GetOpenShiftByEmployee(ctx context.Context, userID uuid.UUID) (*models.Shift, error) {
	if f.getOpenShiftByEmployeeFun != nil {
		return f.getOpenShiftByEmployeeFun(ctx, userID)
	}
	return nil, nil
}
func (f *fakeShiftStore) CloseShift(ctx context.Context, shiftID uuid.UUID, closingCash decimal.Decimal, notes *string, closedBy uuid.UUID) error {
	if f.closeShiftFunc != nil {
		return f.closeShiftFunc(ctx, shiftID, closingCash, notes, closedBy)
	}
	return nil
}
func (f *fakeShiftStore) List(ctx context.Context, filter *repository.ShiftFilter) ([]models.Shift, int, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, filter)
	}
	return nil, 0, nil
}

func TestListShiftsRejectsInvalidUserID(t *testing.T) {
	handler := NewShiftHandler(&fakeShiftStore{})
	app := fiber.New()
	app.Get("/shifts", handler.ListShifts)

	req := httptest.NewRequest(http.MethodGet, "/shifts?user_id=bad-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetCurrentShiftBlindMasksSensitiveFields(t *testing.T) {
	userID := uuid.New()
	shift := &models.Shift{
		ID:               uuid.New(),
		EmployeeID:       userID,
		OpeningCash:      decimal.RequireFromString("100"),
		TotalSales:       decimal.RequireFromString("55"),
		TotalRefunds:     decimal.RequireFromString("5"),
		TransactionCount: 3,
		RefundCount:      1,
	}
	expectedCash := decimal.RequireFromString("150")
	shift.ExpectedCash = &expectedCash

	handler := NewShiftHandler(&fakeShiftStore{
		getOpenShiftByEmployeeFun: func(context.Context, uuid.UUID) (*models.Shift, error) {
			return shift, nil
		},
	})

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	})
	app.Get("/shifts/current", handler.GetCurrentShift)

	req := httptest.NewRequest(http.MethodGet, "/shifts/current?blind=true", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload struct {
		Shift models.Shift `json:"shift"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !payload.Shift.TotalSales.Equal(decimal.Zero) {
		t.Fatalf("expected total_sales to be masked, got %s", payload.Shift.TotalSales)
	}
	if payload.Shift.ExpectedCash != nil {
		t.Fatal("expected expected_cash to be masked")
	}
	if payload.Shift.TransactionCount != 0 {
		t.Fatalf("expected transaction_count to be masked, got %d", payload.Shift.TransactionCount)
	}
}
