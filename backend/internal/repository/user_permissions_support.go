package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *UserRepository) GetUserPermissionsForRole(ctx context.Context, userID, roleID uuid.UUID) ([]string, error) {
	permissionSet, err := r.loadRolePermissionSet(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if err := r.applyUserPermissionOverrides(ctx, userID, permissionSet); err != nil {
		return nil, err
	}

	return permissionSetToSlice(permissionSet), nil
}

func (r *UserRepository) loadRolePermissionSet(ctx context.Context, roleID uuid.UUID) (map[string]bool, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.key
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
	`, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query role permissions: %w", err)
	}
	defer rows.Close()

	permissionSet := make(map[string]bool)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissionSet[key] = true
	}

	return permissionSet, nil
}

func (r *UserRepository) applyUserPermissionOverrides(ctx context.Context, userID uuid.UUID, permissionSet map[string]bool) error {
	rows, err := r.pool.Query(ctx, `
		SELECT p.key, up.allowed
		FROM user_permissions up
		JOIN permissions p ON up.permission_id = p.id
		WHERE up.user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var allowed bool
		if err := rows.Scan(&key, &allowed); err != nil {
			return fmt.Errorf("failed to scan user permission: %w", err)
		}
		if allowed {
			permissionSet[key] = true
			continue
		}
		delete(permissionSet, key)
	}

	return nil
}

func permissionSetToSlice(permissionSet map[string]bool) []string {
	permissions := make([]string, 0, len(permissionSet))
	for key := range permissionSet {
		permissions = append(permissions, key)
	}
	return permissions
}
