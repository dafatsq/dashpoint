package audit

import (
	"context"
	"testing"

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
