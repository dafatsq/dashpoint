package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"dashpoint/backend/internal/authz"
)

func (r *UserRepository) GetUserPermissionsForRole(ctx context.Context, userID, roleID uuid.UUID) ([]string, error) {
	_ = userID

	var roleName string
	err := r.pool.QueryRow(ctx, `SELECT name FROM roles WHERE id = $1`, roleID).Scan(&roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get role name: %w", err)
	}

	return authz.PermissionsForRole(roleName), nil
}
