package repository

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *RoleRepository) UpdatePermissions(ctx context.Context, roleID uuid.UUID, permissionKeys []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start role permissions transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var roleExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`, roleID).Scan(&roleExists); err != nil {
		return fmt.Errorf("failed to check role: %w", err)
	}
	if !roleExists {
		return pgx.ErrNoRows
	}

	keys := uniquePermissionKeys(permissionKeys)
	if len(keys) > 0 {
		rows, err := tx.Query(ctx, `SELECT key FROM permissions WHERE key = ANY($1)`, keys)
		if err != nil {
			return fmt.Errorf("failed to validate permissions: %w", err)
		}
		defer rows.Close()

		found := make(map[string]bool, len(keys))
		for rows.Next() {
			var key string
			if err := rows.Scan(&key); err != nil {
				return fmt.Errorf("failed to scan permission key: %w", err)
			}
			found[key] = true
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("failed to read permission keys: %w", err)
		}
		if len(found) != len(keys) {
			return fmt.Errorf("invalid permission key")
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
		return fmt.Errorf("failed to clear role permissions: %w", err)
	}

	if len(keys) > 0 {
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT $1, id
			FROM permissions
			WHERE key = ANY($2)
		`, roleID, keys); err != nil {
			return fmt.Errorf("failed to insert role permissions: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit role permissions: %w", err)
	}

	return nil
}

func uniquePermissionKeys(keys []string) []string {
	seen := make(map[string]bool, len(keys))
	unique := make([]string, 0, len(keys))
	for _, key := range keys {
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, key)
	}
	sort.Strings(unique)
	return unique
}
