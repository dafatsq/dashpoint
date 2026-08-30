package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"dashpoint/backend/internal/models"
)

var ErrRefreshTokenNotActive = errors.New("refresh token is not active")

// RefreshTokenRepository handles refresh token database operations
type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

type refreshTokenTx interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Commit(context.Context) error
	Rollback(context.Context) error
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

// Create creates a new refresh token record
func (r *RefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, family_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	token.ID = uuid.New()
	token.CreatedAt = time.Now()
	if token.FamilyID == uuid.Nil {
		token.FamilyID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
		token.FamilyID,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetByTokenHash retrieves a refresh token by its hash
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at, revoked_reason, family_id
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	token := &models.RefreshToken{}
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.RevokedAt,
		&token.RevokedReason,
		&token.FamilyID,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return token, nil
}

// Revoke revokes a refresh token
func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string, reason string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1, revoked_reason = $2
		WHERE token_hash = $3 AND revoked_at IS NULL
	`

	now := time.Now()
	_, err := r.pool.Exec(ctx, query, now, reason, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// Rotate revokes the current token and stores the replacement in one transaction.
func (r *RefreshTokenRepository) Rotate(ctx context.Context, currentTokenHash string, reason string, replacement *models.RefreshToken) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin refresh token rotation: %w", err)
	}
	defer tx.Rollback(ctx)

	return rotateRefreshTokenTx(ctx, tx, currentTokenHash, reason, replacement)
}

func rotateRefreshTokenTx(ctx context.Context, tx refreshTokenTx, currentTokenHash string, reason string, replacement *models.RefreshToken) error {
	now := time.Now()
	if replacement.ID == uuid.Nil {
		replacement.ID = uuid.New()
	}
	if replacement.FamilyID == uuid.Nil {
		replacement.FamilyID = uuid.New()
	}
	replacement.CreatedAt = now

	tag, err := tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $1, revoked_reason = $2
		WHERE token_hash = $3 AND revoked_at IS NULL
	`, now, reason, currentTokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token during rotation: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ErrRefreshTokenNotActive
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, family_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, replacement.ID, replacement.UserID, replacement.TokenHash, replacement.ExpiresAt, replacement.CreatedAt, replacement.FamilyID); err != nil {
		return fmt.Errorf("failed to store rotated refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit refresh token rotation: %w", err)
	}

	return nil
}

// RevokeAllForUser revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID, reason string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1, revoked_reason = $2
		WHERE user_id = $3 AND revoked_at IS NULL
	`

	now := time.Now()
	_, err := r.pool.Exec(ctx, query, now, reason, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all refresh tokens: %w", err)
	}

	return nil
}

// RevokeFamily revokes every live token minted from the same login (and its
// rotation descendants). Used to contain a compromised family when a revoked
// token is replayed.
func (r *RefreshTokenRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID, reason string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1, revoked_reason = $2
		WHERE family_id = $3 AND revoked_at IS NULL
	`

	now := time.Now()
	_, err := r.pool.Exec(ctx, query, now, reason, familyID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token family: %w", err)
	}

	return nil
}

// DeleteExpired deletes expired refresh tokens (for cleanup)
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < $1 OR (revoked_at IS NOT NULL AND revoked_at < $2)
	`

	now := time.Now()
	oneWeekAgo := now.Add(-7 * 24 * time.Hour)

	result, err := r.pool.Exec(ctx, query, now, oneWeekAgo)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return result.RowsAffected(), nil
}

// CountActiveForUser counts active refresh tokens for a user
func (r *RefreshTokenRepository) CountActiveForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM refresh_tokens
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > $2
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID, time.Now()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active tokens: %w", err)
	}

	return count, nil
}
