package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"dashpoint/backend/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, name, password_hash, pin_hash, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.PasswordHash,
		user.PINHash,
		user.RoleID,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
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

// GetByEmail retrieves a user by email (case-insensitive)
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

// GetActiveUsers retrieves all active users with a specific PIN (for PIN login lookup)
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

// UpdateLastLogin updates the last login timestamp for a user
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// GetUserPermissions retrieves the effective permissions for a user
// This combines role permissions with user-specific overrides
func (r *UserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// First, get the user's role
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get role permissions
	rolePermsQuery := `
		SELECT p.key
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
	`
	rows, err := r.pool.Query(ctx, rolePermsQuery, user.RoleID)
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

	// Get user-specific overrides
	userPermsQuery := `
		SELECT p.key, up.allowed
		FROM user_permissions up
		JOIN permissions p ON up.permission_id = p.id
		WHERE up.user_id = $1
	`
	overrideRows, err := r.pool.Query(ctx, userPermsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer overrideRows.Close()

	for overrideRows.Next() {
		var key string
		var allowed bool
		if err := overrideRows.Scan(&key, &allowed); err != nil {
			return nil, fmt.Errorf("failed to scan user permission: %w", err)
		}
		if allowed {
			permissionSet[key] = true
		} else {
			delete(permissionSet, key)
		}
	}

	// Convert to slice
	permissions := make([]string, 0, len(permissionSet))
	for key := range permissionSet {
		permissions = append(permissions, key)
	}

	return permissions, nil
}

// List retrieves all users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int, activeOnly bool) ([]*models.User, int, error) {
	// Count query
	countQuery := `SELECT COUNT(*) FROM users`
	if activeOnly {
		countQuery += ` WHERE is_active = true`
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// List query
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
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// ListWithFilter retrieves users with optional active status filter
func (r *UserRepository) ListWithFilter(ctx context.Context, limit, offset int, isActive *bool) ([]*models.User, int, error) {
	// Count query
	countQuery := `SELECT COUNT(*) FROM users`
	var args []interface{}
	argNum := 1

	if isActive != nil {
		countQuery += fmt.Sprintf(` WHERE is_active = $%d`, argNum)
		args = append(args, *isActive)
		argNum++
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// List query
	query := `
		SELECT u.id, u.email, u.name, u.password_hash, u.pin_hash, u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
	`
	if isActive != nil {
		query += fmt.Sprintf(` WHERE u.is_active = $1`)
		query += fmt.Sprintf(` ORDER BY u.created_at DESC LIMIT $2 OFFSET $3`)
	} else {
		query += ` ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`
	}

	var rows pgx.Rows
	if isActive != nil {
		rows, err = r.pool.Query(ctx, query, *isActive, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx, query, limit, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// Update updates a user's basic information
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET name = $1, email = $2, role_id = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`

	user.UpdatedAt = time.Now()
	result, err := r.pool.Exec(ctx, query,
		user.Name,
		user.Email,
		user.RoleID,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	now := time.Now()
	result, err := r.pool.Exec(ctx, query, passwordHash, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// UpdatePIN updates a user's PIN
func (r *UserRepository) UpdatePIN(ctx context.Context, userID uuid.UUID, pinHash *string) error {
	query := `UPDATE users SET pin_hash = $1, updated_at = $2 WHERE id = $3`
	now := time.Now()
	result, err := r.pool.Exec(ctx, query, pinHash, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update PIN: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// Deactivate deactivates a user (soft delete)
func (r *UserRepository) Deactivate(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET is_active = false, updated_at = $1 WHERE id = $2`
	now := time.Now()
	result, err := r.pool.Exec(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// HasSalesHistory checks if a user has created any sales
func (r *UserRepository) HasSalesHistory(ctx context.Context, userID uuid.UUID) (bool, error) {
	// Check if sales table exists first
	var tableExists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sales')`).Scan(&tableExists)
	if err != nil {
		// If we can't check, assume no sales history (safe default)
		return false, nil
	}

	// If sales table doesn't exist, user has no sales history
	if !tableExists {
		return false, nil
	}

	// Check if cashier_id column exists
	var columnExists bool
	err = r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'sales' AND column_name = 'cashier_id')`).Scan(&columnExists)
	if err != nil || !columnExists {
		// Column doesn't exist or can't check - assume no sales history
		return false, nil
	}

	query := `SELECT COUNT(*) FROM sales WHERE cashier_id = $1`
	var count int
	err = r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		// Any error checking sales - assume no sales history (safe default)
		return false, nil
	}
	return count > 0, nil
}

// PermanentDelete permanently deletes a user and related data
func (r *UserRepository) PermanentDelete(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete user permissions (CASCADE handles this, but be explicit)
	_, err = tx.Exec(ctx, `DELETE FROM user_permissions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user permissions: %w", err)
	}

	// Nullify granted_by references in user_permissions
	_, err = tx.Exec(ctx, `UPDATE user_permissions SET granted_by = NULL WHERE granted_by = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to nullify granted_by references: %w", err)
	}

	// Delete refresh tokens (CASCADE handles this, but be explicit)
	_, err = tx.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}

	// Nullify audit log references
	_, err = tx.Exec(ctx, `UPDATE audit_logs SET user_id = NULL WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to nullify audit logs: %w", err)
	}

	// Nullify expense references
	_, err = tx.Exec(ctx, `UPDATE expenses SET created_by = NULL WHERE created_by = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to nullify expense references: %w", err)
	}

	// Nullify stock adjustment references
	_, err = tx.Exec(ctx, `UPDATE stock_adjustments SET adjusted_by = NULL WHERE adjusted_by = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to nullify stock adjustments: %w", err)
	}

	// Nullify payment processor references
	_, err = tx.Exec(ctx, `UPDATE payments SET processed_by = NULL WHERE processed_by = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to nullify payment references: %w", err)
	}

	// Nullify voided_by in sales
	_, err = tx.Exec(ctx, `UPDATE sales SET voided_by = NULL WHERE voided_by = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to nullify sales voided_by: %w", err)
	}

	// Delete shifts (only if no sales associated - already checked in handler)
	_, err = tx.Exec(ctx, `DELETE FROM shifts WHERE employee_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete shifts: %w", err)
	}

	// Delete the user
	result, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SetUserPermission sets a specific permission override for a user
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

// RemoveUserPermission removes a specific permission override for a user
func (r *UserRepository) RemoveUserPermission(ctx context.Context, userID, permissionID uuid.UUID) error {
	query := `DELETE FROM user_permissions WHERE user_id = $1 AND permission_id = $2`
	_, err := r.pool.Exec(ctx, query, userID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove user permission: %w", err)
	}
	return nil
}

// GetUserPermissionOverrides retrieves all permission overrides for a user
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

// EmailExists checks if an email is already in use by another user (case-insensitive)
func (r *UserRepository) EmailExists(ctx context.Context, email string, excludeUserID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE LOWER(email) = LOWER($1)`
	args := []interface{}{email}

	if excludeUserID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeUserID)
	}

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}
