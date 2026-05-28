package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// Create starts a new shift.
func (r *ShiftRepository) Create(ctx context.Context, shift *models.Shift) error {
	now := time.Now()
	shift.ID = uuid.New()
	shift.StartedAt = now
	shift.Status = models.ShiftStatusOpen
	shift.CreatedAt = now
	shift.UpdatedAt = now

	_, err := r.pool.Exec(ctx, `
		INSERT INTO shifts (
			id, employee_id, started_at, opening_cash, status, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, shift.ID, shift.EmployeeID, shift.StartedAt, shift.OpeningCash, shift.Status, shift.Notes, shift.CreatedAt, shift.UpdatedAt)
	return err
}

// CloseShift closes a shift.
func (r *ShiftRepository) CloseShift(ctx context.Context, shiftID uuid.UUID, closingCash decimal.Decimal, notes *string, closedBy uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	openingCash, err := getOpenShiftOpeningCashTx(ctx, tx, shiftID)
	if err != nil {
		return err
	}
	cashSales, refundedCash, payIns, payOuts, err := loadShiftCashTotalsTx(ctx, tx, shiftID)
	if err != nil {
		return err
	}

	expectedCash := calculateExpectedCash(openingCash, cashSales, refundedCash, payIns, payOuts)
	cashDifference := closingCash.Sub(expectedCash)
	now := time.Now()

	tag, err := tx.Exec(ctx, `
		UPDATE shifts
		SET
			ended_at = $1,
			closing_cash = $2,
			expected_cash = $3,
			cash_difference = $4,
			status = 'closed',
			notes = COALESCE($5, notes),
			updated_at = $6,
			closed_by = $7
		WHERE id = $8 AND status = 'open'
	`, now, closingCash, expectedCash, cashDifference, notes, now, closedBy, shiftID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("shift not found or already closed")
	}

	return tx.Commit(ctx)
}

// UpdateSalesTotals updates the shift sales totals.
func (r *ShiftRepository) UpdateSalesTotals(ctx context.Context, shiftID uuid.UUID, saleAmount decimal.Decimal, isRefund bool) error {
	now := time.Now()
	query := `
		UPDATE shifts
		SET
			total_sales = total_sales + $1,
			transaction_count = transaction_count + 1,
			updated_at = $2
		WHERE id = $3
	`
	if isRefund {
		query = `
			UPDATE shifts
			SET
				total_refunds = total_refunds + $1,
				refund_count = refund_count + 1,
				updated_at = $2
			WHERE id = $3
		`
	}
	_, err := r.pool.Exec(ctx, query, saleAmount, now, shiftID)
	return err
}

func getOpenShiftOpeningCashTx(ctx context.Context, tx pgx.Tx, shiftID uuid.UUID) (decimal.Decimal, error) {
	var openingCash decimal.Decimal
	err := tx.QueryRow(ctx, `SELECT opening_cash FROM shifts WHERE id = $1 AND status = 'open' FOR UPDATE`, shiftID).Scan(&openingCash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return decimal.Zero, fmt.Errorf("shift not found or already closed")
		}
		return decimal.Zero, err
	}
	return openingCash, nil
}

func loadShiftCashTotalsTx(ctx context.Context, tx pgx.Tx, shiftID uuid.UUID) (decimal.Decimal, decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	var cashSales, refundedCash, payIns, payOuts decimal.Decimal
	err := tx.QueryRow(ctx, `
		SELECT
			COALESCE((
				SELECT SUM(p.amount)
				FROM payments p
				JOIN sales s ON p.sale_id = s.id
				WHERE s.shift_id = $1 AND p.payment_method = 'cash' AND p.status = 'completed'
			), 0),
			COALESCE((
				SELECT SUM(p.amount)
				FROM payments p
				JOIN sales s ON p.sale_id = s.id
				WHERE s.shift_id = $1 AND s.status = 'refunded' AND p.payment_method = 'cash' AND p.status = 'refunded'
			), 0),
			COALESCE((SELECT SUM(amount) FROM cash_drawer_operations WHERE shift_id = $1 AND type = 'pay_in'), 0),
			COALESCE((SELECT SUM(amount) FROM cash_drawer_operations WHERE shift_id = $1 AND type = 'pay_out'), 0)
	`, shiftID).Scan(&cashSales, &refundedCash, &payIns, &payOuts)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	return cashSales, refundedCash, payIns, payOuts, nil
}

func calculateExpectedCash(openingCash, cashSales, refundedCash, payIns, payOuts decimal.Decimal) decimal.Decimal {
	return openingCash.Add(cashSales).Sub(refundedCash).Add(payIns).Sub(payOuts)
}
