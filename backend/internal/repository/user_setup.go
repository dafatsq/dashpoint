package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"dashpoint/backend/internal/models"
)

// setupAdvisoryLockKey serializes initial-owner creation across concurrent
// requests and server replicas while the zero-active-users invariant is
// re-checked inside the setup transaction.
const setupAdvisoryLockKey int64 = 0x64736870 // "dshp"

// execQuerier abstracts the Exec surface shared by the connection pool and
// transactions so user inserts can run in either context.
type execQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

const insertUserQuery = `
	INSERT INTO users (id, email, name, password_hash, pin_hash, role_id, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

func insertUser(ctx context.Context, q execQuerier, user *models.User) error {
	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := q.Exec(ctx, insertUserQuery, user.ID, user.Email, user.Name, user.PasswordHash, user.PINHash, user.RoleID, user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// HasActiveUser reports whether at least one active user exists. The initial
// setup flow is only offered while this is false.
func (r *UserRepository) HasActiveUser(ctx context.Context) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE is_active = true)`

	var exists bool
	if err := r.pool.QueryRow(ctx, query).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check for active users: %w", err)
	}
	return exists, nil
}

// CreateInitialOwner creates the first owner account, but only while the
// database has zero active users. A transaction-scoped advisory lock plus an
// in-transaction re-check guarantees that exactly one concurrent request can
// ever win; all losers observe created == false.
func (r *UserRepository) CreateInitialOwner(ctx context.Context, user *models.User) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to begin initial setup: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, setupAdvisoryLockKey); err != nil {
		return false, fmt.Errorf("failed to lock initial setup: %w", err)
	}

	var activeUsers int
	query := `SELECT COUNT(*) FROM users WHERE is_active = true`
	if err := tx.QueryRow(ctx, query).Scan(&activeUsers); err != nil {
		return false, fmt.Errorf("failed to check for active users: %w", err)
	}
	if activeUsers > 0 {
		return false, nil
	}

	if err := insertUser(ctx, tx, user); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit initial setup: %w", err)
	}

	return true, nil
}
