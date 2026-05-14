package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

// GetByID retrieves a permission by ID
func (r *PermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	query := `
		SELECT id, key, name, description, category, created_at
		FROM permissions
		WHERE id = $1
	`

	permission := &models.Permission{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&permission.ID,
		&permission.Key,
		&permission.Name,
		&permission.Description,
		&permission.Category,
		&permission.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by ID: %w", err)
	}

	return permission, nil
}

// GetByKey retrieves a permission by key
func (r *PermissionRepository) GetByKey(ctx context.Context, key string) (*models.Permission, error) {
	query := `
		SELECT id, key, name, description, category, created_at
		FROM permissions
		WHERE key = $1
	`

	permission := &models.Permission{}
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&permission.ID,
		&permission.Key,
		&permission.Name,
		&permission.Description,
		&permission.Category,
		&permission.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by key: %w", err)
	}

	return permission, nil
}

// List retrieves all permissions
func (r *PermissionRepository) List(ctx context.Context) ([]*models.Permission, error) {
	query := `
		SELECT id, key, name, description, category, created_at
		FROM permissions
		ORDER BY category, key
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		if err := rows.Scan(
			&permission.ID,
			&permission.Key,
			&permission.Name,
			&permission.Description,
			&permission.Category,
			&permission.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// ListByCategory retrieves permissions grouped by category
func (r *PermissionRepository) ListByCategory(ctx context.Context) (map[string][]*models.Permission, error) {
	permissions, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	return groupPermissionsByCategory(permissions), nil
}
