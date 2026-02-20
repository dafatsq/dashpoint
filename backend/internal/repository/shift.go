package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// ShiftRepository handles shift database operations
type ShiftRepository struct {
	pool *pgxpool.Pool
}

// NewShiftRepository creates a new shift repository
func NewShiftRepository(pool *pgxpool.Pool) *ShiftRepository {
	return &ShiftRepository{pool: pool}
}

// Create starts a new shift
func (r *ShiftRepository) Create(ctx context.Context, shift *models.Shift) error {
	now := time.Now()
	shift.ID = uuid.New()
	shift.StartedAt = now
	shift.Status = models.ShiftStatusOpen
	shift.CreatedAt = now
	shift.UpdatedAt = now

	query := `
		INSERT INTO shifts (
			id, employee_id, started_at, opening_cash, status, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		shift.ID,
		shift.EmployeeID,
		shift.StartedAt,
		shift.OpeningCash,
		shift.Status,
		shift.Notes,
		shift.CreatedAt,
		shift.UpdatedAt,
	)

	return err
}

// GetByID retrieves a shift by ID
func (r *ShiftRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Shift, error) {
	query := `
		SELECT 
			s.id, s.employee_id, s.started_at, s.ended_at, s.opening_cash, s.closing_cash,
			s.expected_cash, s.cash_difference, s.total_sales, s.total_refunds,
			(
				SELECT COALESCE(SUM(p.amount), 0)
				FROM payments p
				JOIN sales s2 ON p.sale_id = s2.id
				WHERE s2.shift_id = s.id AND p.payment_method = 'cash' AND p.status = 'completed'
			) as total_cash_sales,
			s.transaction_count, s.refund_count, s.status, s.notes, s.created_at, s.updated_at,
			u.name as employee_name,
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
		LEFT JOIN users u ON s.employee_id = u.id
		LEFT JOIN users cu ON s.closed_by = cu.id
		WHERE s.id = $1
	`

	shift := &models.Shift{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&shift.ID,
		&shift.EmployeeID,
		&shift.StartedAt,
		&shift.EndedAt,
		&shift.OpeningCash,
		&shift.ClosingCash,
		&shift.ExpectedCash,
		&shift.CashDifference,
		&shift.TotalSales,
		&shift.TotalRefunds,
		&shift.TotalCashSales,
		&shift.TransactionCount,
		&shift.RefundCount,
		&shift.Status,
		&shift.Notes,
		&shift.CreatedAt,
		&shift.UpdatedAt,
		&shift.EmployeeName,
		&shift.ClosedBy,
		&shift.ClosedByName,
		&shift.Operations,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return shift, nil
}

// GetOpenShiftByEmployee gets the current open shift for an employee
// GetOpenShiftByEmployee gets the current open shift for an employee, or falls back to any open shift
func (r *ShiftRepository) GetOpenShiftByEmployee(ctx context.Context, employeeID uuid.UUID) (*models.Shift, error) {
	// First, try to find a shift specifically for this employee
	query := `
		SELECT 
			s.id, s.employee_id, s.started_at, s.ended_at, s.opening_cash, s.closing_cash,
			s.expected_cash, s.cash_difference, s.total_sales, s.total_refunds,
			(
				SELECT COALESCE(SUM(p.amount), 0)
				FROM payments p
				JOIN sales s2 ON p.sale_id = s2.id
				WHERE s2.shift_id = s.id AND p.payment_method = 'cash' AND p.status = 'completed'
			) as total_cash_sales,
			s.transaction_count, s.refund_count, s.status, s.notes, s.created_at, s.updated_at,
			u.name as employee_name,
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
		LEFT JOIN users u ON s.employee_id = u.id
		LEFT JOIN users cu ON s.closed_by = cu.id
		WHERE s.employee_id = $1 AND s.status = 'open'
		ORDER BY s.started_at DESC
		LIMIT 1
	`

	shift := &models.Shift{}
	err := r.pool.QueryRow(ctx, query, employeeID).Scan(
		&shift.ID,
		&shift.EmployeeID,
		&shift.StartedAt,
		&shift.EndedAt,
		&shift.OpeningCash,
		&shift.ClosingCash,
		&shift.ExpectedCash,
		&shift.CashDifference,
		&shift.TotalSales,
		&shift.TotalRefunds,
		&shift.TotalCashSales,
		&shift.TransactionCount,
		&shift.RefundCount,
		&shift.Status,
		&shift.Notes,
		&shift.CreatedAt,
		&shift.UpdatedAt,
		&shift.EmployeeName,
		&shift.ClosedBy,
		&shift.ClosedByName,
		&shift.Operations,
	)

	if err == nil {
		return shift, nil
	}

	if err != pgx.ErrNoRows {
		return nil, err
	}

	// Falls back to ANY open shift (shared drawer mode)
	fallbackQuery := `
		SELECT 
			s.id, s.employee_id, s.started_at, s.ended_at, s.opening_cash, s.closing_cash,
			s.expected_cash, s.cash_difference, s.total_sales, s.total_refunds,
			(
				SELECT COALESCE(SUM(p.amount), 0)
				FROM payments p
				JOIN sales s2 ON p.sale_id = s2.id
				WHERE s2.shift_id = s.id AND p.payment_method = 'cash' AND p.status = 'completed'
			) as total_cash_sales,
			s.transaction_count, s.refund_count, s.status, s.notes, s.created_at, s.updated_at,
			u.name as employee_name,
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
		LEFT JOIN users u ON s.employee_id = u.id
		LEFT JOIN users cu ON s.closed_by = cu.id
		WHERE s.status = 'open'
		ORDER BY s.started_at DESC
		LIMIT 1
	`

	err = r.pool.QueryRow(ctx, fallbackQuery).Scan(
		&shift.ID,
		&shift.EmployeeID,
		&shift.StartedAt,
		&shift.EndedAt,
		&shift.OpeningCash,
		&shift.ClosingCash,
		&shift.ExpectedCash,
		&shift.CashDifference,
		&shift.TotalSales,
		&shift.TotalRefunds,
		&shift.TotalCashSales,
		&shift.TransactionCount,
		&shift.RefundCount,
		&shift.Status,
		&shift.Notes,
		&shift.CreatedAt,
		&shift.UpdatedAt,
		&shift.EmployeeName,
		&shift.ClosedBy,
		&shift.ClosedByName,
		&shift.Operations,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return shift, nil
}

// CloseShift closes a shift
func (r *ShiftRepository) CloseShift(ctx context.Context, shiftID uuid.UUID, closingCash decimal.Decimal, notes *string, closedBy uuid.UUID) error {
	now := time.Now()

	// expected_cash = opening + sales - refunds + pay_ins - pay_outs
	query := `
		UPDATE shifts 
		SET 
			ended_at = $1,
			closing_cash = $2,
			expected_cash = opening_cash 
				+ COALESCE((
					SELECT SUM(p.amount) 
					FROM payments p 
					JOIN sales s ON p.sale_id = s.id 
					WHERE s.shift_id = $5 
					AND p.payment_method = 'cash' 
					AND p.status = 'completed'
				), 0)
				- COALESCE((
					SELECT SUM(p.amount) 
					FROM payments p 
					JOIN sales s ON p.sale_id = s.id 
					WHERE s.shift_id = $5 
					AND p.payment_method = 'cash' 
					AND p.status = 'refunded'
				), 0)
				+ COALESCE((SELECT SUM(amount) FROM cash_drawer_operations WHERE shift_id = $5 AND type = 'pay_in'), 0)
				- COALESCE((SELECT SUM(amount) FROM cash_drawer_operations WHERE shift_id = $5 AND type = 'pay_out'), 0),
			cash_difference = $2 - (opening_cash 
				+ COALESCE((
					SELECT SUM(p.amount) 
					FROM payments p 
					JOIN sales s ON p.sale_id = s.id 
					WHERE s.shift_id = $5 
					AND p.payment_method = 'cash' 
					AND p.status = 'completed'
				), 0)
				- COALESCE((
					SELECT SUM(p.amount) 
					FROM payments p 
					JOIN sales s ON p.sale_id = s.id 
					WHERE s.shift_id = $5 
					AND p.payment_method = 'cash' 
					AND p.status = 'refunded'
				), 0)
				+ COALESCE((SELECT SUM(amount) FROM cash_drawer_operations WHERE shift_id = $5 AND type = 'pay_in'), 0)
				- COALESCE((SELECT SUM(amount) FROM cash_drawer_operations WHERE shift_id = $5 AND type = 'pay_out'), 0)),
			status = 'closed',
			notes = COALESCE($3, notes),
			updated_at = $4,
			closed_by = $6
		WHERE id = $5 AND status = 'open'
	`

	result, err := r.pool.Exec(ctx, query, now, closingCash, notes, now, shiftID, closedBy)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("shift not found or already closed")
	}

	return nil
}

// UpdateSalesTotals updates the shift sales totals
func (r *ShiftRepository) UpdateSalesTotals(ctx context.Context, shiftID uuid.UUID, saleAmount decimal.Decimal, isRefund bool) error {
	now := time.Now()

	var query string
	if isRefund {
		query = `
			UPDATE shifts 
			SET 
				total_refunds = total_refunds + $1,
				refund_count = refund_count + 1,
				updated_at = $2
			WHERE id = $3
		`
	} else {
		query = `
			UPDATE shifts 
			SET 
				total_sales = total_sales + $1,
				transaction_count = transaction_count + 1,
				updated_at = $2
			WHERE id = $3
		`
	}

	_, err := r.pool.Exec(ctx, query, saleAmount, now, shiftID)
	return err
}

// List retrieves shifts with pagination
func (r *ShiftRepository) List(ctx context.Context, employeeID *uuid.UUID, limit, offset int) ([]models.Shift, int, error) {
	countQuery := `SELECT COUNT(*) FROM shifts WHERE ($1::uuid IS NULL OR employee_id = $1)`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, employeeID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			s.id, s.employee_id, s.started_at, s.ended_at, s.opening_cash, s.closing_cash,
			s.expected_cash, s.cash_difference, s.total_sales, s.total_refunds,
			s.transaction_count, s.refund_count, s.status, s.notes, s.created_at, s.updated_at,
			u.name as employee_name,
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
		LEFT JOIN users u ON s.employee_id = u.id
		LEFT JOIN users cu ON s.closed_by = cu.id
		WHERE ($1::uuid IS NULL OR s.employee_id = $1)
		ORDER BY s.started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, employeeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var shifts []models.Shift
	for rows.Next() {
		var s models.Shift
		err := rows.Scan(
			&s.ID,
			&s.EmployeeID,
			&s.StartedAt,
			&s.EndedAt,
			&s.OpeningCash,
			&s.ClosingCash,
			&s.ExpectedCash,
			&s.CashDifference,
			&s.TotalSales,
			&s.TotalRefunds,
			&s.TransactionCount,
			&s.RefundCount,
			&s.Status,
			&s.Notes,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.EmployeeName,
			&s.ClosedBy,
			&s.ClosedByName,
			&s.Operations,
		)
		if err != nil {
			return nil, 0, err
		}
		shifts = append(shifts, s)
	}

	return shifts, total, nil
}
