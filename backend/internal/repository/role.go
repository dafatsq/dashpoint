package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"dashpoint/backend/internal/models"
)

// RoleRepository handles role database operations
type RoleRepository struct {
	pool *pgxpool.Pool
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE id = $1
	`

	role := &models.Role{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
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
		SELECT id, name, description, created_at, updated_at
		FROM roles
		ORDER BY 
			CASE name 
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
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *RoleRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	query := `
		SELECT p.id, p.key, p.name, p.description, p.category, p.created_at
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.key
	`

	rows, err := r.pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		p := &models.Permission{}
		err := rows.Scan(
			&p.ID,
			&p.Key,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, p)
	}

	return permissions, nil
}

// PermissionRepository handles permission database operations
type PermissionRepository struct {
	pool *pgxpool.Pool
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(pool *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{pool: pool}
}

// GetByID retrieves a permission by ID
func (r *PermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	query := `
		SELECT id, key, name, description, category, created_at
		FROM permissions
		WHERE id = $1
	`

	p := &models.Permission{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Key,
		&p.Name,
		&p.Description,
		&p.Category,
		&p.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by ID: %w", err)
	}

	return p, nil
}

// GetByKey retrieves a permission by key
func (r *PermissionRepository) GetByKey(ctx context.Context, key string) (*models.Permission, error) {
	query := `
		SELECT id, key, name, description, category, created_at
		FROM permissions
		WHERE key = $1
	`

	p := &models.Permission{}
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&p.ID,
		&p.Key,
		&p.Name,
		&p.Description,
		&p.Category,
		&p.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by key: %w", err)
	}

	return p, nil
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
		p := &models.Permission{}
		err := rows.Scan(
			&p.ID,
			&p.Key,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, p)
	}

	return permissions, nil
}

// ListByCategory retrieves permissions grouped by category
func (r *PermissionRepository) ListByCategory(ctx context.Context) (map[string][]*models.Permission, error) {
	permissions, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]*models.Permission)
	for _, p := range permissions {
		grouped[p.Category] = append(grouped[p.Category], p)
	}

	return grouped, nil
}
