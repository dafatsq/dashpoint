package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *UserRepository) GetUserPermissionsForRole(ctx context.Context, userID, roleID uuid.UUID) ([]string, error) {
	_ = userID

	var roleName string
	err := r.pool.QueryRow(ctx, `SELECT name FROM roles WHERE id = $1`, roleID).Scan(&roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get role name: %w", err)
	}

	query := `
		SELECT p.key
		FROM permissions p
	`
	args := []interface{}{}
	if roleName != "owner" {
		query += `
			JOIN role_permissions rp ON p.id = rp.permission_id
			WHERE rp.role_id = $1
		`
		args = append(args, roleID)
	}
	query += ` ORDER BY p.category, p.key`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	permissions := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read permissions: %w", err)
	}

	return permissions, nil
}
