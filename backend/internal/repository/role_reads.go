package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at,
		       COALESCE(array_agg(p.key ORDER BY p.category, p.key) FILTER (WHERE p.key IS NOT NULL), '{}')
		FROM roles r
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		LEFT JOIN permissions p ON rp.permission_id = p.id
		WHERE r.id = $1
		GROUP BY r.id, r.name, r.description, r.created_at, r.updated_at
	`

	role := &models.Role{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
		&role.Permissions,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role by ID: %w", err)
	}

	return role, nil
}

// GetByName retrieves a role by name
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE name = $1
	`

	role := &models.Role{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	return role, nil
}

// List retrieves all roles
func (r *RoleRepository) List(ctx context.Context) ([]*models.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at,
		       COALESCE(array_agg(p.key ORDER BY p.category, p.key) FILTER (WHERE p.key IS NOT NULL), '{}')
		FROM roles r
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		LEFT JOIN permissions p ON rp.permission_id = p.id
		GROUP BY r.id, r.name, r.description, r.created_at, r.updated_at
		ORDER BY
			CASE r.name
				WHEN 'owner' THEN 1
				WHEN 'manager' THEN 2
				WHEN 'cashier' THEN 3
				ELSE 4
			END
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
			&role.UpdatedAt,
			&role.Permissions,
		); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}
