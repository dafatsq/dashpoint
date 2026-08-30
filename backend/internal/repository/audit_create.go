package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

// Create creates a new audit log entry.
func (r *AuditRepository) Create(ctx context.Context, entry *models.AuditLogEntry) error {
	now := time.Now()
	id := uuid.New()

	oldValuesJSON, err := marshalAuditMap(entry.OldValues, "old_values")
	if err != nil {
		return err
	}
	newValuesJSON, err := marshalAuditMap(entry.NewValues, "new_values")
	if err != nil {
		return err
	}
	metadataJSON, err := marshalAuditMap(entry.Metadata, "metadata")
	if err != nil {
		return err
	}

	status := entry.Status
	if status == "" {
		status = models.AuditStatusSuccess
	}

	query := `
		INSERT INTO audit_logs (
			id, created_at, user_id, user_name, user_role, action, entity_type,
			entity_id, description, old_values, new_values, metadata, status,
			ip_address, user_agent, request_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err = r.pool.Exec(ctx, query,
		id, now, entry.UserID, optionalAuditString(entry.UserName), optionalAuditString(entry.UserRole),
		entry.Action, entry.EntityType, optionalAuditString(entry.EntityID), optionalAuditString(entry.Description),
		oldValuesJSON, newValuesJSON, metadataJSON, status,
		optionalAuditString(entry.IPAddress), optionalAuditString(entry.UserAgent), optionalAuditString(entry.RequestID),
	)

	return err
}

func marshalAuditMap(value map[string]interface{}, field string) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s: %w", field, err)
	}
	return data, nil
}

func optionalAuditString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
