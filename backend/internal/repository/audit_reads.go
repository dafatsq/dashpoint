package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

const auditSelectColumns = `
	SELECT
		audit_logs.id, audit_logs.created_at, audit_logs.user_id, audit_logs.user_email,
		COALESCE(u.name, audit_logs.user_name) as user_name, audit_logs.user_role,
		audit_logs.action, audit_logs.entity_type, audit_logs.entity_id, audit_logs.description,
		audit_logs.old_values, audit_logs.new_values, audit_logs.metadata,
		audit_logs.ip_address, audit_logs.user_agent, audit_logs.request_id, audit_logs.status
	FROM audit_logs
	LEFT JOIN users u ON audit_logs.user_id = u.id
`

// List retrieves audit logs with filters.
func (r *AuditRepository) List(ctx context.Context, filter AuditFilter) ([]models.AuditLog, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}

	whereClause, args, argIndex := buildAuditWhereClause(filter)

	countQuery := "SELECT COUNT(*) FROM audit_logs LEFT JOIN users u ON audit_logs.user_id = u.id " + whereClause
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("%s %s ORDER BY audit_logs.created_at DESC LIMIT $%d OFFSET $%d", auditSelectColumns, whereClause, argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs, err := collectAuditLogs(rows)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID retrieves a single audit log by ID.
func (r *AuditRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	query := auditSelectColumns + " WHERE audit_logs.id = $1"

	row := r.pool.QueryRow(ctx, query, id)
	logEntry, err := scanAuditLog(row)
	if err != nil {
		return nil, err
	}
	return logEntry, nil
}

// GetEntityHistory retrieves audit history for a specific entity.
func (r *AuditRepository) GetEntityHistory(ctx context.Context, entityType string, entityID string, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := auditSelectColumns + `
		WHERE audit_logs.entity_type = $1 AND audit_logs.entity_id = $2
		ORDER BY audit_logs.created_at DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectAuditLogs(rows)
}

// GetUserActivity retrieves recent activity for a user.
func (r *AuditRepository) GetUserActivity(ctx context.Context, userID uuid.UUID, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := auditSelectColumns + `
		WHERE audit_logs.user_id = $1
		ORDER BY audit_logs.created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectAuditLogs(rows)
}
