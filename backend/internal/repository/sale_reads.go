package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

const saleSummarySelectColumns = `
		SELECT
			s.id, s.invoice_no, s.subtotal, s.tax_amount, s.discount_amount, s.total_amount,
			s.item_count, s.payment_status, s.amount_paid, s.change_amount, s.discount_type,
			s.discount_value, s.discount_reason, s.employee_id, s.shift_id,
			s.status, s.voided_at, s.voided_by, s.void_reason, s.notes,
			s.created_at, s.updated_at, u.name as employee_name
`

const saleListSelectColumns = `
		SELECT
			s.id, s.invoice_no, s.subtotal, s.tax_amount, s.discount_amount, s.total_amount,
			s.item_count, s.payment_status, s.amount_paid, s.change_amount,
			s.employee_id, s.shift_id, s.status, s.created_at, s.updated_at,
			u.name as employee_name,
			(SELECT payment_method FROM payments WHERE sale_id = s.id LIMIT 1) as payment_method
`

type saleScanner interface {
	Scan(dest ...any) error
}

type saleRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanSaleSummary(scanner saleScanner) (*models.Sale, error) {
	sale := &models.Sale{}
	if err := scanner.Scan(
		&sale.ID, &sale.InvoiceNo, &sale.Subtotal, &sale.TaxAmount, &sale.DiscountAmount,
		&sale.TotalAmount, &sale.ItemCount, &sale.PaymentStatus, &sale.AmountPaid, &sale.ChangeAmount,
		&sale.DiscountType, &sale.DiscountValue, &sale.DiscountReason, &sale.EmployeeID, &sale.ShiftID,
		&sale.Status, &sale.VoidedAt, &sale.VoidedBy,
		&sale.VoidReason, &sale.Notes, &sale.CreatedAt, &sale.UpdatedAt, &sale.EmployeeName,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return sale, nil
}

func scanSaleListRow(scanner saleScanner) (*models.Sale, error) {
	var (
		sale          models.Sale
		paymentMethod *string
	)
	if err := scanner.Scan(
		&sale.ID, &sale.InvoiceNo, &sale.Subtotal, &sale.TaxAmount, &sale.DiscountAmount, &sale.TotalAmount,
		&sale.ItemCount, &sale.PaymentStatus, &sale.AmountPaid, &sale.ChangeAmount,
		&sale.EmployeeID, &sale.ShiftID, &sale.Status, &sale.CreatedAt, &sale.UpdatedAt,
		&sale.EmployeeName, &paymentMethod,
	); err != nil {
		return nil, err
	}
	if paymentMethod != nil {
		sale.Payments = []models.Payment{{PaymentMethod: models.PaymentMethod(*paymentMethod)}}
	}
	return &sale, nil
}

func collectSales(rows saleRows) ([]models.Sale, error) {
	sales := make([]models.Sale, 0)
	for rows.Next() {
		sale, err := scanSaleListRow(rows)
		if err != nil {
			return nil, err
		}
		sales = append(sales, *sale)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sales, nil
}

// GetByID retrieves a sale by ID with items and payments.
func (r *SaleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Sale, error) {
	sale, err := scanSaleSummary(r.pool.QueryRow(ctx, `
`+saleSummarySelectColumns+`
		FROM sales s
		LEFT JOIN users u ON s.employee_id = u.id
		WHERE s.id = $1
	`, id))
	if err != nil {
		return nil, err
	}
	if sale == nil {
		return nil, nil
	}

	items, err := loadSaleItems(ctx, r.pool, id)
	if err != nil {
		return nil, err
	}
	payments, err := loadSalePayments(ctx, r.pool, id)
	if err != nil {
		return nil, err
	}
	sale.Items = items
	sale.Payments = payments

	return sale, nil
}

// GetByInvoiceNo retrieves a sale by invoice number.
func (r *SaleRepository) GetByInvoiceNo(ctx context.Context, invoiceNo string) (*models.Sale, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM sales WHERE invoice_no = $1`, invoiceNo).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return r.GetByID(ctx, id)
}

// List retrieves sales with pagination and filters.
func (r *SaleRepository) List(ctx context.Context, filter *SaleFilter) ([]models.Sale, int, error) {
	args := []interface{}{}
	argIndex := 1
	whereClause := "WHERE 1=1"

	if filter.EmployeeID != nil {
		whereClause += fmt.Sprintf(" AND s.employee_id = $%d", argIndex)
		args = append(args, *filter.EmployeeID)
		argIndex++
	}
	if filter.ShiftID != nil {
		whereClause += fmt.Sprintf(" AND s.shift_id = $%d", argIndex)
		args = append(args, *filter.ShiftID)
		argIndex++
	}
	if filter.Status != nil {
		whereClause += fmt.Sprintf(" AND s.status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}
	if filter.StartDate != nil {
		whereClause += fmt.Sprintf(" AND s.created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}
	if filter.EndDate != nil {
		whereClause += fmt.Sprintf(" AND s.created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}
	if filter.InvoiceSearch != nil {
		whereClause += fmt.Sprintf(" AND s.invoice_no ILIKE $%d", argIndex)
		args = append(args, "%"+*filter.InvoiceSearch+"%")
		argIndex++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sales s %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
%s
		FROM sales s
		LEFT JOIN users u ON s.employee_id = u.id
		%s
		ORDER BY s.created_at DESC
		LIMIT $%d OFFSET $%d
	`, saleListSelectColumns, whereClause, argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	sales, err := collectSales(rows)
	if err != nil {
		return nil, 0, err
	}

	return sales, total, nil
}

func loadSaleItems(ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, saleID uuid.UUID) ([]models.SaleItem, error) {
	rows, err := db.Query(ctx, `
		SELECT
			id, sale_id, product_id, product_name, product_sku, product_barcode,
			quantity, unit_price, cost_price, discount_type, discount_value, discount_amount,
			tax_rate, tax_amount, subtotal, total, created_at
		FROM sale_items
		WHERE sale_id = $1
		ORDER BY created_at
	`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.SaleItem
	for rows.Next() {
		var item models.SaleItem
		if err := rows.Scan(
			&item.ID, &item.SaleID, &item.ProductID, &item.ProductName, &item.ProductSKU,
			&item.ProductBarcode, &item.Quantity, &item.UnitPrice, &item.CostPrice,
			&item.DiscountType, &item.DiscountValue, &item.DiscountAmount, &item.TaxRate,
			&item.TaxAmount, &item.Subtotal, &item.Total, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func loadSalePayments(ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, saleID uuid.UUID) ([]models.Payment, error) {
	rows, err := db.Query(ctx, `
		SELECT
			id, sale_id, payment_method, amount, amount_tendered, change_given,
			reference_no, status, created_at
		FROM payments
		WHERE sale_id = $1
		ORDER BY created_at
	`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next() {
		var payment models.Payment
		if err := rows.Scan(
			&payment.ID, &payment.SaleID, &payment.PaymentMethod, &payment.Amount,
			&payment.AmountTendered, &payment.ChangeGiven, &payment.ReferenceNo,
			&payment.Status, &payment.CreatedAt,
		); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	return payments, nil
}
