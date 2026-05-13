package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

// SetUserPermission sets a specific permission override for a user.
func (r *UserRepository) SetUserPermission(ctx context.Context, userID, permissionID uuid.UUID, allowed bool, grantedBy *uuid.UUID) error {
	query := `
		INSERT INTO user_permissions (user_id, permission_id, allowed, granted_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (user_id, permission_id) 
		DO UPDATE SET allowed = $3, granted_by = $4, updated_at = $5
	`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query, userID, permissionID, allowed, grantedBy, now)
	if err != nil {
		return fmt.Errorf("failed to set user permission: %w", err)
	}
	return nil
}

// RemoveUserPermission removes a specific permission override for a user.
func (r *UserRepository) RemoveUserPermission(ctx context.Context, userID, permissionID uuid.UUID) error {
	query := `DELETE FROM user_permissions WHERE user_id = $1 AND permission_id = $2`
	_, err := r.pool.Exec(ctx, query, userID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove user permission: %w", err)
	}
	return nil
}

// ClearUserPermissionOverrides removes all permission overrides for a user.
func (r *UserRepository) ClearUserPermissionOverrides(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_permissions WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear user permission overrides: %w", err)
	}
	return nil
}

// GetUserPermissionOverrides retrieves all permission overrides for a user.
func (r *UserRepository) GetUserPermissionOverrides(ctx context.Context, userID uuid.UUID) ([]*models.UserPermission, error) {
	query := `
		SELECT up.user_id, up.permission_id, up.allowed, up.granted_by, up.created_at, up.updated_at,
		       p.id, p.key, p.name, p.description, p.category, p.created_at
		FROM user_permissions up
		JOIN permissions p ON up.permission_id = p.id
		WHERE up.user_id = $1
		ORDER BY p.category, p.key
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permission overrides: %w", err)
	}
	defer rows.Close()

	var overrides []*models.UserPermission
	for rows.Next() {
		up := &models.UserPermission{Permission: &models.Permission{}}
		err := rows.Scan(
			&up.UserID,
			&up.PermissionID,
			&up.Allowed,
			&up.GrantedBy,
			&up.CreatedAt,
			&up.UpdatedAt,
			&up.Permission.ID,
			&up.Permission.Key,
			&up.Permission.Name,
			&up.Permission.Description,
			&up.Permission.Category,
			&up.Permission.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user permission: %w", err)
		}
		overrides = append(overrides, up)
	}

	return overrides, nil
}
