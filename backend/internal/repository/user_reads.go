package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT u.id, u.email, u.name, u.password_hash, u.pin_hash, u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`

	user := &models.User{Role: &models.Role{}}
	err := r.pool.QueryRow(ctx, query, id).Scan(
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
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email (case-insensitive).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT u.id, u.email, u.name, u.password_hash, u.pin_hash, u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE LOWER(u.email) = LOWER($1)
	`

	user := &models.User{Role: &models.Role{}}
	err := r.pool.QueryRow(ctx, query, email).Scan(
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
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetActiveUsersWithPIN retrieves all active users with a specific PIN (for PIN login lookup).
func (r *UserRepository) GetActiveUsersWithPIN(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT u.id, u.email, u.name, u.password_hash, u.pin_hash, u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.is_active = true AND u.pin_hash IS NOT NULL
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users with PIN: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		err := rows.Scan(
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
			&user.Role.ID,
			&user.Role.Name,
			&user.Role.Description,
			&user.Role.CreatedAt,
			&user.Role.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
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

	query := `
		SELECT u.id, u.email, u.name, u.password_hash, u.pin_hash, u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
	`
	if activeOnly {
		query += ` WHERE u.is_active = true`
	}
	query += ` ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.PINHash, &user.RoleID,
			&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
			&user.Role.ID, &user.Role.Name, &user.Role.Description, &user.Role.CreatedAt, &user.Role.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, total, nil
}

// ListWithFilter retrieves users with optional active status, search, and role filters.
func (r *UserRepository) ListWithFilter(ctx context.Context, limit, offset int, isActive *bool, search, role string) ([]*models.User, int, error) {
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
	query := `
		SELECT u.id, u.email, u.name, u.password_hash, u.pin_hash, u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
	` + where + fmt.Sprintf(` ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d`, argNum, argNum+1)

	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.PINHash, &user.RoleID,
			&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
			&user.Role.ID, &user.Role.Name, &user.Role.Description, &user.Role.CreatedAt, &user.Role.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, total, nil
}
