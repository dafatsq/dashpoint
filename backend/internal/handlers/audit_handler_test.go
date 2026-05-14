package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type fakeAuditStore struct {
	listFunc             func(context.Context, repository.AuditFilter) ([]models.AuditLog, int, error)
	getByIDFunc          func(context.Context, uuid.UUID) (*models.AuditLog, error)
	getEntityHistoryFunc func(context.Context, string, string, int) ([]models.AuditLog, error)
	getUserActivityFunc  func(context.Context, uuid.UUID, int) ([]models.AuditLog, error)
	getActionSummaryFunc func(context.Context, time.Time, time.Time) ([]map[string]interface{}, error)
}

func (f *fakeAuditStore) List(ctx context.Context, filter repository.AuditFilter) ([]models.AuditLog, int, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (f *fakeAuditStore) GetByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (f *fakeAuditStore) GetEntityHistory(ctx context.Context, entityType, entityID string, limit int) ([]models.AuditLog, error) {
	if f.getEntityHistoryFunc != nil {
		return f.getEntityHistoryFunc(ctx, entityType, entityID, limit)
	}
	return nil, nil
}

func (f *fakeAuditStore) GetUserActivity(ctx context.Context, userID uuid.UUID, limit int) ([]models.AuditLog, error) {
	if f.getUserActivityFunc != nil {
		return f.getUserActivityFunc(ctx, userID, limit)
	}
	return nil, nil
}

func (f *fakeAuditStore) GetActionSummary(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	if f.getActionSummaryFunc != nil {
		return f.getActionSummaryFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

func TestAuditListRejectsInvalidUserID(t *testing.T) {
	handler := NewAuditHandler(&fakeAuditStore{})
	app := fiber.New()
	app.Get("/logs", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/logs?user_id=bad-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAuditListRejectsInvalidStartDate(t *testing.T) {
	handler := NewAuditHandler(&fakeAuditStore{})
	app := fiber.New()
	app.Get("/logs", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/logs?start_date=05-14-2026", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
