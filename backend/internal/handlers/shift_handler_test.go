package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type fakeShiftStore struct {
	createFunc            func(context.Context, *models.Shift) error
	getByIDFunc           func(context.Context, uuid.UUID) (*models.Shift, error)
	getCurrentOpenShiftFn func(context.Context) (*models.Shift, error)
	closeShiftFunc        func(context.Context, uuid.UUID, decimal.Decimal, *string, uuid.UUID) error
	listFunc              func(context.Context, *repository.ShiftFilter) ([]models.Shift, int, error)
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
func (f *fakeShiftStore) GetCurrentOpenShift(ctx context.Context) (*models.Shift, error) {
	if f.getCurrentOpenShiftFn != nil {
		return f.getCurrentOpenShiftFn(ctx)
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

func TestListShiftsRejectsInvalidOpenedByID(t *testing.T) {
	handler := NewShiftHandler(&fakeShiftStore{})
	app := fiber.New()
	app.Get("/shifts", handler.ListShifts)

	req := httptest.NewRequest(http.MethodGet, "/shifts?opened_by_id=bad-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestListShiftsUsesJakartaExclusiveDateBounds(t *testing.T) {
	var got *repository.ShiftFilter
	handler := NewShiftHandler(&fakeShiftStore{
		listFunc: func(_ context.Context, filter *repository.ShiftFilter) ([]models.Shift, int, error) {
			got = filter
			return []models.Shift{}, 0, nil
		},
	})
	app := fiber.New()
	app.Get("/shifts", handler.ListShifts)

	req := httptest.NewRequest(http.MethodGet, "/shifts?from=2026-05-29&to=2026-05-29", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	jakarta := time.FixedZone("WIB", 7*60*60)
	expectedStart := time.Date(2026, 5, 29, 0, 0, 0, 0, jakarta)
	expectedEnd := expectedStart.Add(24 * time.Hour)
	if got == nil || got.StartDate == nil || !got.StartDate.Equal(expectedStart) {
		t.Fatalf("expected startDate %v, got %+v", expectedStart, got)
	}
	if got.EndDate == nil || !got.EndDate.Equal(expectedEnd) {
		t.Fatalf("expected exclusive endDate %v, got %v", expectedEnd, got.EndDate)
	}
}

func TestGetCurrentShiftBlindMasksSensitiveFields(t *testing.T) {
	userID := uuid.New()
	shift := &models.Shift{
		ID:               uuid.New(),
		OpenedBy:         userID,
		OpeningCash:      decimal.RequireFromString("100"),
		TotalSales:       decimal.RequireFromString("55"),
		TotalVoided:      decimal.RequireFromString("5"),
		TransactionCount: 3,
		VoidCount:        1,
	}
	expectedCash := decimal.RequireFromString("150")
	shift.ExpectedCash = &expectedCash

	handler := NewShiftHandler(&fakeShiftStore{
		getCurrentOpenShiftFn: func(context.Context) (*models.Shift, error) {
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
