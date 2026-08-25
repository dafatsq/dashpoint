package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

// Create creates a new user.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return insertUser(ctx, r.pool, user)
}

// Update updates a user's basic information.
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET name = $1, email = $2, role_id = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`

	user.UpdatedAt = time.Now()
	result, err := r.pool.Exec(ctx, query, user.Name, user.Email, user.RoleID, user.IsActive, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// UpdateLastLogin updates the last login timestamp for a user.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// UpdatePassword updates a user's password.
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

// UpdatePIN updates a user's PIN.
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

// Deactivate deactivates a user (soft delete).
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

// EmailExists checks if an email is already in use by another user (case-insensitive).
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
