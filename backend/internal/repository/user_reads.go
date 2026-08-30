package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

const userWithRoleSelect = `
	SELECT u.id, u.email, u.name, u.password_hash, u.pin_hash, u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at, u.token_version,
	       r.id, r.name, r.description, r.created_at, r.updated_at
	FROM users u
	JOIN roles r ON u.role_id = r.id
`

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return scanOptionalUser(r.pool.QueryRow(ctx, userWithRoleSelect+` WHERE u.id = $1`, id), "failed to get user by ID")
}

// GetByEmail retrieves a user by email (case-insensitive).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return scanOptionalUser(r.pool.QueryRow(ctx, userWithRoleSelect+` WHERE LOWER(u.email) = LOWER($1)`, email), "failed to get user by email")
}

// GetActiveUsersWithPIN retrieves all active users with a specific PIN (for PIN login lookup).
func (r *UserRepository) GetActiveUsersWithPIN(ctx context.Context) ([]*models.User, error) {
	rows, err := r.pool.Query(ctx, userWithRoleSelect+` WHERE u.is_active = true AND u.pin_hash IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("failed to query users with PIN: %w", err)
	}
	defer rows.Close()

	return collectUsers(rows)
}

// GetUserPermissions retrieves the effective permissions for a user.
func (r *UserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return r.GetUserPermissionsForRole(ctx, userID, user.RoleID)
}

// List retrieves all users with pagination.
func (r *UserRepository) List(ctx context.Context, limit, offset int, activeOnly bool) ([]*models.User, int, error) {
	countQuery := `SELECT COUNT(*) FROM users`
	if activeOnly {
		countQuery += ` WHERE is_active = true`
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	query := userWithRoleSelect
	if activeOnly {
		query += ` WHERE u.is_active = true`
	}
	query += ` ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users, err := collectUsers(rows)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// ListWithFilter retrieves users with optional active status, search, and role filters.
func (r *UserRepository) ListWithFilter(ctx context.Context, limit, offset int, isActive *bool, search, role, sortBy, sortDirection string) ([]*models.User, int, error) {
	var args []interface{}
	argNum := 1
	whereClauses := []string{}

	if isActive != nil {
		whereClauses = append(whereClauses, fmt.Sprintf(`u.is_active = $%d`, argNum))
		args = append(args, *isActive)
		argNum++
	}
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(`(u.name ILIKE $%d OR u.email ILIKE $%d)`, argNum, argNum))
		args = append(args, "%"+search+"%")
		argNum++
	}
	if role != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(`r.name ILIKE $%d`, argNum))
		args = append(args, role)
		argNum++
	}

	where := ""
	if len(whereClauses) > 0 {
		where = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM users u JOIN roles r ON u.role_id = r.id` + where
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	listArgs := append(args, limit, offset)
	query := userWithRoleSelect + where + fmt.Sprintf(` ORDER BY %s LIMIT $%d OFFSET $%d`, userOrderBy(sortBy, sortDirection), argNum, argNum+1)

	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users, err := collectUsers(rows)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func userOrderBy(sortBy, sortDirection string) string {
	columns := map[string]string{
		"name":       "LOWER(u.name)",
		"email":      "LOWER(COALESCE(u.email, ''))",
		"role":       "LOWER(r.name)",
		"created_at": "u.created_at",
	}
	column, ok := columns[sortBy]
	if !ok {
		column = columns["created_at"]
	}
	direction := "DESC"
	if sortDirection == "asc" {
		direction = "ASC"
	}
	return fmt.Sprintf("%s %s, u.id ASC", column, direction)
}

// NameExists checks if a user with the given name already exists (case-insensitive).
func (r *UserRepository) NameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE name ILIKE $1 AND is_active = true`
	args := []interface{}{name}

	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}

	var count int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check user name: %w", err)
	}
	return count > 0, nil
}

func scanOptionalUser(row pgx.Row, errPrefix string) (*models.User, error) {
	user := &models.User{Role: &models.Role{}}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.PINHash,
		&user.RoleID,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.TokenVersion,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Description,
		&user.Role.CreatedAt,
		&user.Role.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", errPrefix, err)
	}
	return user, nil
}

func collectUsers(rows pgx.Rows) ([]*models.User, error) {
	var users []*models.User
	for rows.Next() {
		user, err := scanOptionalUser(rows, "failed to scan user")
		if err != nil {
			return nil, err
		}
		if user != nil {
			users = append(users, user)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}
	return users, nil
}
