package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"dashpoint/backend/internal/models"
)

// AuditRepository handles audit log database operations
type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, entry *models.AuditLogEntry) error {
	now := time.Now()
	id := uuid.New()

	// Convert maps to JSON
	var oldValuesJSON, newValuesJSON, metadataJSON []byte
	var err error

	if entry.OldValues != nil {
		oldValuesJSON, err = json.Marshal(entry.OldValues)
		if err != nil {
			return fmt.Errorf("failed to marshal old_values: %w", err)
		}
	}

	if entry.NewValues != nil {
		newValuesJSON, err = json.Marshal(entry.NewValues)
		if err != nil {
			return fmt.Errorf("failed to marshal new_values: %w", err)
		}
	}

	if entry.Metadata != nil {
		metadataJSON, err = json.Marshal(entry.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Set default status
	status := entry.Status
	if status == "" {
		status = models.AuditStatusSuccess
	}

	query := `
		INSERT INTO audit_logs (
			id, created_at, user_id, user_email, user_name, user_role,
			action, entity_type, entity_id, description,
			old_values, new_values, metadata,
			ip_address, user_agent, request_id, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	var entityID, description, ipAddress, userAgent, requestID, userEmail, userName, userRole *string

	if entry.EntityID != "" {
		entityID = &entry.EntityID
	}
	if entry.Description != "" {
		description = &entry.Description
	}
	if entry.IPAddress != "" {
		ipAddress = &entry.IPAddress
	}
	if entry.UserAgent != "" {
		userAgent = &entry.UserAgent
	}
	if entry.RequestID != "" {
		requestID = &entry.RequestID
	}
	if entry.UserEmail != "" {
		userEmail = &entry.UserEmail
	}
	if entry.UserName != "" {
		userName = &entry.UserName
	}
	if entry.UserRole != "" {
		userRole = &entry.UserRole
	}

	_, err = r.pool.Exec(ctx, query,
		id, now, entry.UserID, userEmail, userName, userRole,
		entry.Action, entry.EntityType, entityID, description,
		oldValuesJSON, newValuesJSON, metadataJSON,
		ipAddress, userAgent, requestID, status,
	)

	return err
}

// AuditFilter represents filters for querying audit logs
type AuditFilter struct {
	UserID     *uuid.UUID
	Action     *string
	EntityType *string
	EntityID   *string
	Status     *string
	StartDate  *time.Time
	EndDate    *time.Time
	Search     *string
	Limit      int
	Offset     int
}

// List retrieves audit logs with filters
func (r *AuditRepository) List(ctx context.Context, filter AuditFilter) ([]models.AuditLog, int, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}

	// Build WHERE clause
	args := []interface{}{}
	argIndex := 1
	whereClause := "WHERE 1=1"

	if filter.UserID != nil {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.Action != nil {
		whereClause += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, *filter.Action)
		argIndex++
	}

	if filter.EntityType != nil {
		whereClause += fmt.Sprintf(" AND entity_type = $%d", argIndex)
		args = append(args, *filter.EntityType)
		argIndex++
	}

	if filter.EntityID != nil {
		whereClause += fmt.Sprintf(" AND entity_id = $%d", argIndex)
		args = append(args, *filter.EntityID)
		argIndex++
	}

	if filter.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.StartDate != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if filter.Search != nil && *filter.Search != "" {
		whereClause += fmt.Sprintf(" AND (description ILIKE $%d OR user_email ILIKE $%d OR user_name ILIKE $%d)", argIndex, argIndex, argIndex)
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM audit_logs " + whereClause
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get logs
	query := fmt.Sprintf(`
		SELECT 
			id, created_at, user_id, user_email, user_name, user_role,
			action, entity_type, entity_id, description,
			old_values, new_values, metadata,
			ip_address, user_agent, request_id, status
		FROM audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var oldValuesJSON, newValuesJSON, metadataJSON []byte

		err := rows.Scan(
			&log.ID, &log.CreatedAt, &log.UserID, &log.UserEmail, &log.UserName, &log.UserRole,
			&log.Action, &log.EntityType, &log.EntityID, &log.Description,
			&oldValuesJSON, &newValuesJSON, &metadataJSON,
			&log.IPAddress, &log.UserAgent, &log.RequestID, &log.Status,
		)
		if err != nil {
			return nil, 0, err
		}

		// Parse JSON fields
		if oldValuesJSON != nil {
			json.Unmarshal(oldValuesJSON, &log.OldValues)
		}
		if newValuesJSON != nil {
			json.Unmarshal(newValuesJSON, &log.NewValues)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetByID retrieves a single audit log by ID
func (r *AuditRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	query := `
		SELECT 
			id, created_at, user_id, user_email, user_name, user_role,
			action, entity_type, entity_id, description,
			old_values, new_values, metadata,
			ip_address, user_agent, request_id, status
		FROM audit_logs
		WHERE id = $1
	`

	var log models.AuditLog
	var oldValuesJSON, newValuesJSON, metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&log.ID, &log.CreatedAt, &log.UserID, &log.UserEmail, &log.UserName, &log.UserRole,
		&log.Action, &log.EntityType, &log.EntityID, &log.Description,
		&oldValuesJSON, &newValuesJSON, &metadataJSON,
		&log.IPAddress, &log.UserAgent, &log.RequestID, &log.Status,
	)
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if oldValuesJSON != nil {
		json.Unmarshal(oldValuesJSON, &log.OldValues)
	}
	if newValuesJSON != nil {
		json.Unmarshal(newValuesJSON, &log.NewValues)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &log.Metadata)
	}

	return &log, nil
}

// GetEntityHistory retrieves audit history for a specific entity
func (r *AuditRepository) GetEntityHistory(ctx context.Context, entityType string, entityID string, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT 
			id, created_at, user_id, user_email, user_name, user_role,
			action, entity_type, entity_id, description,
			old_values, new_values, metadata,
			ip_address, user_agent, request_id, status
		FROM audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var oldValuesJSON, newValuesJSON, metadataJSON []byte

		err := rows.Scan(
			&log.ID, &log.CreatedAt, &log.UserID, &log.UserEmail, &log.UserName, &log.UserRole,
			&log.Action, &log.EntityType, &log.EntityID, &log.Description,
			&oldValuesJSON, &newValuesJSON, &metadataJSON,
			&log.IPAddress, &log.UserAgent, &log.RequestID, &log.Status,
		)
		if err != nil {
			return nil, err
		}

		if oldValuesJSON != nil {
			json.Unmarshal(oldValuesJSON, &log.OldValues)
		}
		if newValuesJSON != nil {
			json.Unmarshal(newValuesJSON, &log.NewValues)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetUserActivity retrieves recent activity for a user
func (r *AuditRepository) GetUserActivity(ctx context.Context, userID uuid.UUID, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT 
			id, created_at, user_id, user_email, user_name, user_role,
			action, entity_type, entity_id, description,
			old_values, new_values, metadata,
			ip_address, user_agent, request_id, status
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var oldValuesJSON, newValuesJSON, metadataJSON []byte

		err := rows.Scan(
			&log.ID, &log.CreatedAt, &log.UserID, &log.UserEmail, &log.UserName, &log.UserRole,
			&log.Action, &log.EntityType, &log.EntityID, &log.Description,
			&oldValuesJSON, &newValuesJSON, &metadataJSON,
			&log.IPAddress, &log.UserAgent, &log.RequestID, &log.Status,
		)
		if err != nil {
			return nil, err
		}

		if oldValuesJSON != nil {
			json.Unmarshal(oldValuesJSON, &log.OldValues)
		}
		if newValuesJSON != nil {
			json.Unmarshal(newValuesJSON, &log.NewValues)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetActionSummary gets a summary of actions in a time period
func (r *AuditRepository) GetActionSummary(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			action,
			COUNT(*) as count,
			COUNT(DISTINCT user_id) as unique_users
		FROM audit_logs
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY action
		ORDER BY count DESC
	`

	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var action string
		var count, uniqueUsers int

		if err := rows.Scan(&action, &count, &uniqueUsers); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"action":       action,
			"count":        count,
			"unique_users": uniqueUsers,
		})
	}

	return results, nil
}

// Cleanup deletes audit logs older than specified days
func (r *AuditRepository) Cleanup(ctx context.Context, olderThanDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)

	result, err := r.pool.Exec(ctx, `
		DELETE FROM audit_logs WHERE created_at < $1
	`, cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
