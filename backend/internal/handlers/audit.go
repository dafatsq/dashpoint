package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type auditStore interface {
	List(context.Context, repository.AuditFilter) ([]models.AuditLog, int, error)
	GetByID(context.Context, uuid.UUID) (*models.AuditLog, error)
	GetEntityHistory(context.Context, string, string, int) ([]models.AuditLog, error)
	GetUserActivity(context.Context, uuid.UUID, int) ([]models.AuditLog, error)
	GetActionSummary(context.Context, time.Time, time.Time) ([]map[string]interface{}, error)
}

// AuditHandler handles audit log endpoints.
type AuditHandler struct {
	auditRepo auditStore
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(auditRepo auditStore) *AuditHandler {
	return &AuditHandler{auditRepo: auditRepo}
}
