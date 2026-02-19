package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// CashDrawerRepository handles cash drawer operation database operations
type CashDrawerRepository struct {
	pool *pgxpool.Pool
}

// NewCashDrawerRepository creates a new cash drawer repository
func NewCashDrawerRepository(pool *pgxpool.Pool) *CashDrawerRepository {
	return &CashDrawerRepository{pool: pool}
}

// Create adds a new cash drawer operation
func (r *CashDrawerRepository) Create(ctx context.Context, op *models.CashDrawerOperation) error {
	op.ID = uuid.New()
	op.CreatedAt = time.Now()

	query := `
		INSERT INTO cash_drawer_operations (id, shift_id, type, amount, reason, performed_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		op.ID,
		op.ShiftID,
		op.Type,
		op.Amount,
		op.Reason,
		op.PerformedBy,
		op.CreatedAt,
	)

	return err
}

// ListByShift gets all cash drawer operations for a shift
func (r *CashDrawerRepository) ListByShift(ctx context.Context, shiftID uuid.UUID) ([]models.CashDrawerOperation, error) {
	query := `
		SELECT 
			c.id, c.shift_id, c.type, c.amount, c.reason, c.performed_by, c.created_at,
			u.name as performed_by_name
		FROM cash_drawer_operations c
		LEFT JOIN users u ON c.performed_by = u.id
		WHERE c.shift_id = $1
		ORDER BY c.created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []models.CashDrawerOperation
	for rows.Next() {
		var op models.CashDrawerOperation
		err := rows.Scan(
			&op.ID,
			&op.ShiftID,
			&op.Type,
			&op.Amount,
			&op.Reason,
			&op.PerformedBy,
			&op.CreatedAt,
			&op.PerformedByName,
		)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}

	return ops, nil
}

// GetTotalsByShift gets the sum of pay-ins and pay-outs for a shift
func (r *CashDrawerRepository) GetTotalsByShift(ctx context.Context, shiftID uuid.UUID) (payInTotal, payOutTotal decimal.Decimal, err error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'pay_in' THEN amount ELSE 0 END), 0) as pay_in_total,
			COALESCE(SUM(CASE WHEN type = 'pay_out' THEN amount ELSE 0 END), 0) as pay_out_total
		FROM cash_drawer_operations
		WHERE shift_id = $1
	`

	err = r.pool.QueryRow(ctx, query, shiftID).Scan(&payInTotal, &payOutTotal)
	return
}
