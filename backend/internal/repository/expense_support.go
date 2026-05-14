package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

const expenseSelectColumns = `
	SELECT e.id, e.category_id, ec.name as category_name, e.product_id, p.name as product_name,
	       e.quantity, e.amount, e.description,
	       e.expense_date, e.vendor, e.reference_number, e.notes,
	       e.created_by, u.name as created_by_name, e.created_at, e.updated_at
	FROM expenses e
	LEFT JOIN expense_categories ec ON e.category_id = ec.id
	LEFT JOIN products p ON e.product_id = p.id
	LEFT JOIN users u ON e.created_by = u.id
`

type expenseScanner interface {
	Scan(dest ...interface{}) error
}

func scanExpense(scanner expenseScanner) (*models.Expense, error) {
	var expense models.Expense
	if err := scanner.Scan(
		&expense.ID,
		&expense.CategoryID,
		&expense.CategoryName,
		&expense.ProductID,
		&expense.ProductName,
		&expense.Quantity,
		&expense.Amount,
		&expense.Description,
		&expense.ExpenseDate,
		&expense.Vendor,
		&expense.ReferenceNumber,
		&expense.Notes,
		&expense.CreatedBy,
		&expense.CreatedByName,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}
	return &expense, nil
}

func buildExpenseListBaseQuery(categoryID *uuid.UUID, startDate, endDate *time.Time) (string, []interface{}, int) {
	baseQuery := `
		FROM expenses e
		LEFT JOIN expense_categories ec ON e.category_id = ec.id
		LEFT JOIN products p ON e.product_id = p.id
		LEFT JOIN users u ON e.created_by = u.id
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if categoryID != nil {
		baseQuery += fmt.Sprintf(" AND e.category_id = $%d", argNum)
		args = append(args, *categoryID)
		argNum++
	}
	if startDate != nil {
		baseQuery += fmt.Sprintf(" AND e.expense_date >= $%d", argNum)
		args = append(args, *startDate)
		argNum++
	}
	if endDate != nil {
		baseQuery += fmt.Sprintf(" AND e.expense_date <= $%d", argNum)
		args = append(args, *endDate)
		argNum++
	}

	return baseQuery, args, argNum
}
