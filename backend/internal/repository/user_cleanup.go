package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type salesHistoryQuerier interface {
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

// HasSalesHistory checks if a user has created any sales.
func (r *UserRepository) HasSalesHistory(ctx context.Context, userID uuid.UUID) (bool, error) {
	return hasSalesHistory(ctx, r.pool, userID)
}

func hasSalesHistory(ctx context.Context, querier salesHistoryQuerier, userID uuid.UUID) (bool, error) {
	var tableExists bool
	err := querier.QueryRow(ctx, `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sales')`).Scan(&tableExists)
	if err != nil {
		return false, fmt.Errorf("failed to verify sales table existence: %w", err)
	}
	if !tableExists {
		return false, nil
	}

	var columnExists bool
	err = querier.QueryRow(ctx, `SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'sales' AND column_name = 'cashier_id')`).Scan(&columnExists)
	if err != nil {
		return false, fmt.Errorf("failed to verify sales.cashier_id existence: %w", err)
	}
	if !columnExists {
		return false, nil
	}

	query := `SELECT COUNT(*) FROM sales WHERE cashier_id = $1`
	var count int
	err = querier.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to query sales history: %w", err)
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
		{name: "delete user permissions", query: `DELETE FROM user_permissions WHERE user_id = $1`},
		{name: "nullify permission grants", query: `UPDATE user_permissions SET granted_by = NULL WHERE granted_by = $1`},
		{name: "delete refresh tokens", query: `DELETE FROM refresh_tokens WHERE user_id = $1`},
		{name: "nullify audit logs", query: `UPDATE audit_logs SET user_id = NULL WHERE user_id = $1`},
		{name: "nullify expenses", query: `UPDATE expenses SET created_by = NULL WHERE created_by = $1`},
		{name: "nullify stock adjustments", query: `UPDATE stock_adjustments SET adjusted_by = NULL WHERE adjusted_by = $1`},
		{name: "nullify payments", query: `UPDATE payments SET processed_by = NULL WHERE processed_by = $1`},
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
