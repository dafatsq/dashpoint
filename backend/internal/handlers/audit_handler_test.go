package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestAuditListRejectsInvalidFilterValues(t *testing.T) {
	handler := NewAuditHandler(&fakeAuditStore{})
	app := fiber.New()
	app.Get("/logs", handler.List)

	tests := []string{
		"/logs?limit=101",
		"/logs?offset=-1",
		"/logs?action=bad.action",
		"/logs?entity_type=bad",
		"/logs?entity_id=../../etc/passwd",
		"/logs?status=bad",
		"/logs?search=" + strings.Repeat("x", 201),
	}

	for _, path := range tests {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s returned error: %v", path, err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("%s expected 400, got %d", path, resp.StatusCode)
		}
	}
}

func TestAuditListUsesJakartaExclusiveEndDate(t *testing.T) {
	var got repository.AuditFilter
	handler := NewAuditHandler(&fakeAuditStore{
		listFunc: func(_ context.Context, filter repository.AuditFilter) ([]models.AuditLog, int, error) {
			got = filter
			return []models.AuditLog{}, 0, nil
		},
	})
	app := fiber.New()
	app.Get("/logs", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/logs?start_date=2026-05-29&end_date=2026-05-29", nil)
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
	if got.StartDate == nil || !got.StartDate.Equal(expectedStart) {
		t.Fatalf("expected startDate %v, got %v", expectedStart, got.StartDate)
	}
	if got.EndDate == nil || !got.EndDate.Equal(expectedEnd) {
		t.Fatalf("expected exclusive endDate %v, got %v", expectedEnd, got.EndDate)
	}
}

func TestAuditEntityHistoryRejectsInvalidParams(t *testing.T) {
	handler := NewAuditHandler(&fakeAuditStore{})
	app := fiber.New()
	app.Get("/logs/entity/:type/:id", handler.GetEntityHistory)

	tests := []string{
		"/logs/entity/bad/11111111-1111-1111-1111-111111111111",
		"/logs/entity/product/bad$value",
		"/logs/entity/product/11111111-1111-1111-1111-111111111111?limit=0",
		"/logs/entity/product/11111111-1111-1111-1111-111111111111?limit=101",
	}

	for _, path := range tests {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s returned error: %v", path, err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("%s expected 400, got %d", path, resp.StatusCode)
		}
	}
}

func TestAuditSummaryUsesJakartaExclusiveEndDate(t *testing.T) {
	var gotStart, gotEnd time.Time
	handler := NewAuditHandler(&fakeAuditStore{
		getActionSummaryFunc: func(_ context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
			gotStart = startDate
			gotEnd = endDate
			return []map[string]interface{}{}, nil
		},
	})
	app := fiber.New()
	app.Get("/logs/summary", handler.GetSummary)

	req := httptest.NewRequest(http.MethodGet, "/logs/summary?start_date=2026-05-29&end_date=2026-05-29", nil)
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
	if !gotStart.Equal(expectedStart) {
		t.Fatalf("expected startDate %v, got %v", expectedStart, gotStart)
	}
	if !gotEnd.Equal(expectedEnd) {
		t.Fatalf("expected exclusive endDate %v, got %v", expectedEnd, gotEnd)
	}
}
