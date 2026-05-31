package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

const shiftSelectColumns = `
	SELECT
		s.id, s.opened_by, s.started_at, s.ended_at, s.opening_cash, s.closing_cash,
		s.expected_cash, s.cash_difference, s.total_sales, s.total_voided,
		(
			SELECT COALESCE(SUM(p.amount), 0)
			FROM payments p
			JOIN sales s2 ON p.sale_id = s2.id
			WHERE s2.shift_id = s.id AND p.payment_method = 'cash' AND p.status = 'completed'
		) as total_cash_sales,
		s.transaction_count, s.void_count, s.status, s.notes, s.created_at, s.updated_at,
		u.name as opened_by_name,
		s.closed_by, cu.name as closed_by_name,
		(
			SELECT COALESCE(
				json_agg(
					json_build_object(
						'id', c.id,
						'shift_id', c.shift_id,
						'type', c.type,
						'amount', c.amount,
						'reason', c.reason,
						'performed_by', c.performed_by,
						'created_at', c.created_at,
						'performed_by_name', pu.name
					) ORDER BY c.created_at ASC
				),
				'[]'::json
			)
			FROM cash_drawer_operations c
			LEFT JOIN users pu ON c.performed_by = pu.id
			WHERE c.shift_id = s.id
		) as operations
	FROM shifts s
	LEFT JOIN users u ON s.opened_by = u.id
	LEFT JOIN users cu ON s.closed_by = cu.id
`

// GetByID retrieves a shift by ID.
func (r *ShiftRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Shift, error) {
	query := shiftSelectColumns + ` WHERE s.id = $1`
	return querySingleShift(ctx, r.pool, query, id)
}

// GetCurrentOpenShift gets the current shared open shift.
func (r *ShiftRepository) GetCurrentOpenShift(ctx context.Context) (*models.Shift, error) {
	query := shiftSelectColumns + `
		WHERE s.status = 'open'
		ORDER BY s.started_at DESC
		LIMIT 1
	`
	return querySingleShift(ctx, r.pool, query)
}

// List retrieves shifts with pagination.
func (r *ShiftRepository) List(ctx context.Context, filter *ShiftFilter) ([]models.Shift, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM shifts
		WHERE ($1::uuid IS NULL OR opened_by = $1)
		AND ($2::timestamp IS NULL OR started_at >= $2)
		AND ($3::timestamp IS NULL OR started_at < $3)
	`

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, filter.OpenedByID, filter.StartDate, filter.EndDate).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := shiftSelectColumns + `
		WHERE ($1::uuid IS NULL OR s.opened_by = $1)
		AND ($2::timestamp IS NULL OR s.started_at >= $2)
		AND ($3::timestamp IS NULL OR s.started_at < $3)
		ORDER BY s.started_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.pool.Query(ctx, query, filter.OpenedByID, filter.StartDate, filter.EndDate, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var shifts []models.Shift
	for rows.Next() {
		shift, err := scanShift(rows)
		if err != nil {
			return nil, 0, err
		}
		shifts = append(shifts, *shift)
	}
	return shifts, total, nil
}

type shiftRowScanner interface {
	Scan(...any) error
}

func querySingleShift(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, query string, args ...any) (*models.Shift, error) {
	shift, err := scanShift(db.QueryRow(ctx, query, args...))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return shift, nil
}

func scanShift(scanner shiftRowScanner) (*models.Shift, error) {
	shift := &models.Shift{}
	err := scanner.Scan(
		&shift.ID,
		&shift.OpenedBy,
		&shift.StartedAt,
		&shift.EndedAt,
		&shift.OpeningCash,
		&shift.ClosingCash,
		&shift.ExpectedCash,
		&shift.CashDifference,
		&shift.TotalSales,
		&shift.TotalVoided,
		&shift.TotalCashSales,
		&shift.TransactionCount,
		&shift.VoidCount,
		&shift.Status,
		&shift.Notes,
		&shift.CreatedAt,
		&shift.UpdatedAt,
		&shift.OpenedByName,
		&shift.ClosedBy,
		&shift.ClosedByName,
		&shift.Operations,
	)
	if err != nil {
		return nil, err
	}
	return shift, nil
}
