package audit

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

type stubAuditRepository struct {
	calls int
	last  *models.AuditLogEntry
}

func (s *stubAuditRepository) Create(_ context.Context, entry *models.AuditLogEntry) error {
	s.calls++
	s.last = entry
	return nil
}

func TestLogWritesThroughInitializedService(t *testing.T) {
	repo := &stubAuditRepository{}
	Init(repo)

	entry := &models.AuditLogEntry{
		Action:     models.AuditActionLogin,
		EntityType: models.AuditEntityAuth,
		Status:     models.AuditStatusSuccess,
	}

	Log(context.Background(), entry)

	if repo.calls != 1 {
		t.Fatalf("expected audit repository to be called once, got %d", repo.calls)
	}
}

func TestLogSyncReturnsNilWhenServiceIsNotInitialized(t *testing.T) {
	globalService = nil

	if err := LogSync(context.Background(), &models.AuditLogEntry{}); err != nil {
		t.Fatalf("expected nil error when audit service is not initialized, got %v", err)
	}
}

func TestFiberHelpersCaptureRequestMetadata(t *testing.T) {
	repo := &stubAuditRepository{}
	Init(repo)
	t.Cleanup(func() { globalService = nil })

	app := fiber.New()
	app.Use(middleware.RequestID())
	var captured *models.AuditLogEntry
	app.Get("/", func(c *fiber.Ctx) error {
		LogFromFiber(c, models.AuditActionLogin, models.AuditEntityAuth, "entity-1", "test entry")
		captured = repo.last
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "test-agent/1.0")
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if captured == nil {
		t.Fatalf("expected audit entry to be captured")
	}
	if captured.IPAddress == "" {
		t.Fatalf("expected IP address to be captured from the request")
	}
	if captured.UserAgent != "test-agent/1.0" {
		t.Fatalf("expected user agent to be captured, got %q", captured.UserAgent)
	}
	if captured.RequestID == "" {
		t.Fatalf("expected request ID to be captured from middleware")
	}
}
