package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type countHistoryQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type userCleanupTx interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Commit(context.Context) error
	Rollback(context.Context) error
}

type namedCleanupQuery struct {
	name  string
	query string
}

const (
	countSalesHistoryQuery   = `SELECT COUNT(*) FROM sales WHERE employee_id = $1`
	countExpenseHistoryQuery = `SELECT COUNT(*) FROM expenses WHERE created_by = $1`
)

// HasSalesHistory checks if a user has created any sales.
func (r *UserRepository) HasSalesHistory(ctx context.Context, userID uuid.UUID) (bool, error) {
	return hasRelatedHistory(ctx, r.pool, countSalesHistoryQuery, "sales", userID)
}

func hasSalesHistory(ctx context.Context, querier countHistoryQuerier, userID uuid.UUID) (bool, error) {
	return hasRelatedHistory(ctx, querier, countSalesHistoryQuery, "sales", userID)
}

// HasExpenseHistory checks if a user has created any expenses.
func (r *UserRepository) HasExpenseHistory(ctx context.Context, userID uuid.UUID) (bool, error) {
	return hasRelatedHistory(ctx, r.pool, countExpenseHistoryQuery, "expense", userID)
}

func hasRelatedHistory(ctx context.Context, querier countHistoryQuerier, query string, historyName string, userID uuid.UUID) (bool, error) {
	var count int
	err := querier.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to query %s history: %w", historyName, err)
	}
	return count > 0, nil
}

// PermanentDelete permanently deletes a user and related data.
func (r *UserRepository) PermanentDelete(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	return permanentDeleteTx(ctx, tx, userID)
}

func permanentDeleteTx(ctx context.Context, tx userCleanupTx, userID uuid.UUID) error {
	queries := []namedCleanupQuery{
		{name: "delete refresh tokens", query: `DELETE FROM refresh_tokens WHERE user_id = $1`},
		{name: "nullify audit logs", query: `UPDATE audit_logs SET user_id = NULL WHERE user_id = $1`},
		{name: "nullify stock adjustments", query: `UPDATE stock_adjustments SET adjusted_by = NULL WHERE adjusted_by = $1`},
		{name: "nullify sales voids", query: `UPDATE sales SET voided_by = NULL WHERE voided_by = $1`},
		{name: "delete shifts", query: `DELETE FROM shifts WHERE employee_id = $1`},
	}

	for _, query := range queries {
		if _, err := tx.Exec(ctx, query.query, userID); err != nil {
			return fmt.Errorf("failed to %s during permanent delete: %w", query.name, err)
		}
	}

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
