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

// SaleRepository handles sale database operations
type SaleRepository struct {
	pool          *pgxpool.Pool
	inventoryRepo *InventoryRepository
}

// NewSaleRepository creates a new sale repository
func NewSaleRepository(pool *pgxpool.Pool, inventoryRepo *InventoryRepository) *SaleRepository {
	return &SaleRepository{
		pool:          pool,
		inventoryRepo: inventoryRepo,
	}
}

// CreateSaleRequest contains all data needed to create a sale
type CreateSaleRequest struct {
	Items          []CreateSaleItemRequest
	Payments       []CreatePaymentRequest
	EmployeeID     uuid.UUID
	ShiftID        *uuid.UUID
	CustomerName   *string
	CustomerPhone  *string
	DiscountType   *string
	DiscountValue  *decimal.Decimal
	DiscountReason *string
	Notes          *string
}

// CreateSaleItemRequest contains data for a sale item
type CreateSaleItemRequest struct {
	ProductID      uuid.UUID
	Quantity       decimal.Decimal
	UnitPrice      decimal.Decimal
	DiscountType   *string
	DiscountValue  decimal.Decimal
	DiscountAmount decimal.Decimal
}

// CreatePaymentRequest contains data for a payment
type CreatePaymentRequest struct {
	PaymentMethod  models.PaymentMethod
	Amount         decimal.Decimal
	AmountTendered *decimal.Decimal
	ChangeGiven    *decimal.Decimal
	CardType       *string
	CardLastFour   *string
	ReferenceNo    *string
	BankName       *string
	AccountNo      *string
	VoucherCode    *string
	Notes          *string
}

// Create creates a new sale with items and payments in a single transaction
func (r *SaleRepository) Create(ctx context.Context, req *CreateSaleRequest) (*models.Sale, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	// Generate invoice number (format: INV-YYYYMMDD-XXXXX)
	invoiceNo, err := r.generateInvoiceNumber(ctx, tx, now)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Calculate totals first - validate all items before inserting anything
	var subtotal, taxAmount, itemDiscountAmount, totalAmount decimal.Decimal
	itemCount := len(req.Items)

	// Prepare sale items - validate all products first
	type preparedItem struct {
		SaleItem   models.SaleItem
		ProductQty decimal.Decimal
		NewQty     decimal.Decimal
	}
	var preparedItems []preparedItem

	for i := range req.Items {
		item := &req.Items[i]

		// Get product details and lock for update
		var product struct {
			ID       uuid.UUID
			Name     string
			SKU      *string
			Barcode  *string
			Price    decimal.Decimal
			Cost     decimal.Decimal
			TaxRate  decimal.Decimal
			Quantity decimal.Decimal
		}

		err := tx.QueryRow(ctx, `
			SELECT p.id, p.name, p.sku, p.barcode, p.price, p.cost, p.tax_rate, COALESCE(i.quantity, 0)
			FROM products p
			LEFT JOIN inventory_items i ON p.id = i.product_id
			WHERE p.id = $1 AND p.is_active = true
			FOR UPDATE OF p
		`, item.ProductID).Scan(
			&product.ID, &product.Name, &product.SKU, &product.Barcode,
			&product.Price, &product.Cost, &product.TaxRate, &product.Quantity,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, fmt.Errorf("product not found: %s", item.ProductID)
			}
			return nil, fmt.Errorf("failed to get product: %w", err)
		}

		// Check stock availability
		if product.Quantity.LessThan(item.Quantity) {
			return nil, fmt.Errorf("insufficient stock for %s: available %s, requested %s",
				product.Name, product.Quantity.String(), item.Quantity.String())
		}

		// Calculate item totals
		itemSubtotal := item.UnitPrice.Mul(item.Quantity)
		itemTax := itemSubtotal.Mul(product.TaxRate).Div(decimal.NewFromInt(100))
		itemTotal := itemSubtotal.Add(itemTax).Sub(item.DiscountAmount)

		preparedItems = append(preparedItems, preparedItem{
			SaleItem: models.SaleItem{
				ID:             uuid.New(),
				ProductID:      product.ID,
				ProductName:    product.Name,
				ProductSKU:     product.SKU,
				ProductBarcode: product.Barcode,
				Quantity:       item.Quantity,
				UnitPrice:      item.UnitPrice,
				CostPrice:      product.Cost,
				DiscountType:   item.DiscountType,
				DiscountValue:  item.DiscountValue,
				DiscountAmount: item.DiscountAmount,
				TaxRate:        product.TaxRate,
				TaxAmount:      itemTax,
				Subtotal:       itemSubtotal,
				Total:          itemTotal,
				CreatedAt:      now,
			},
			ProductQty: product.Quantity,
			NewQty:     product.Quantity.Sub(item.Quantity),
		})

		subtotal = subtotal.Add(itemSubtotal)
		taxAmount = taxAmount.Add(itemTax)
		itemDiscountAmount = itemDiscountAmount.Add(item.DiscountAmount)
	}

	// Apply sale-level discount if any
	discountAmount := itemDiscountAmount
	if req.DiscountValue != nil && req.DiscountType != nil {
		if *req.DiscountType == "percentage" {
			saleDiscount := subtotal.Mul(*req.DiscountValue).Div(decimal.NewFromInt(100))
			discountAmount = discountAmount.Add(saleDiscount)
		} else {
			discountAmount = discountAmount.Add(*req.DiscountValue)
		}
	}

	totalAmount = subtotal.Add(taxAmount).Sub(discountAmount)

	// Calculate payment totals
	var amountPaid decimal.Decimal
	for _, paymentReq := range req.Payments {
		amountPaid = amountPaid.Add(paymentReq.Amount)
	}

	// Determine payment status
	paymentStatus := models.PaymentStatusPending
	changeAmount := decimal.Zero
	if amountPaid.GreaterThanOrEqual(totalAmount) {
		paymentStatus = models.PaymentStatusPaid
		changeAmount = amountPaid.Sub(totalAmount)
	} else if amountPaid.GreaterThan(decimal.Zero) {
		paymentStatus = models.PaymentStatusPartial
	}

	// Create sale record
	sale := &models.Sale{
		ID:             uuid.New(),
		InvoiceNo:      invoiceNo,
		Subtotal:       subtotal,
		TaxAmount:      taxAmount,
		DiscountAmount: discountAmount,
		TotalAmount:    totalAmount,
		ItemCount:      itemCount,
		PaymentStatus:  paymentStatus,
		AmountPaid:     amountPaid,
		ChangeAmount:   changeAmount,
		EmployeeID:     req.EmployeeID,
		ShiftID:        req.ShiftID,
		CustomerName:   req.CustomerName,
		CustomerPhone:  req.CustomerPhone,
		DiscountType:   req.DiscountType,
		DiscountValue:  req.DiscountValue,
		DiscountReason: req.DiscountReason,
		Notes:          req.Notes,
		Status:         models.SaleStatusCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Insert sale FIRST
	_, err = tx.Exec(ctx, `
		INSERT INTO sales (
			id, invoice_no, subtotal, tax_amount, discount_amount, total_amount, item_count,
			payment_status, amount_paid, change_amount, discount_type, discount_value, discount_reason,
			employee_id, shift_id, customer_name, customer_phone, status, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`,
		sale.ID, sale.InvoiceNo, sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount,
		sale.ItemCount, sale.PaymentStatus, sale.AmountPaid, sale.ChangeAmount, sale.DiscountType,
		sale.DiscountValue, sale.DiscountReason, sale.EmployeeID, sale.ShiftID, sale.CustomerName,
		sale.CustomerPhone, sale.Status, sale.Notes, sale.CreatedAt, sale.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert sale: %w", err)
	}

	// Now insert sale items and update inventory
	for _, prepared := range preparedItems {
		saleItem := prepared.SaleItem
		saleItem.SaleID = sale.ID

		// Insert sale item
		_, err = tx.Exec(ctx, `
			INSERT INTO sale_items (
				id, sale_id, product_id, product_name, product_sku, product_barcode,
				quantity, unit_price, cost_price, discount_type, discount_value, discount_amount,
				tax_rate, tax_amount, subtotal, total, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		`,
			saleItem.ID, saleItem.SaleID, saleItem.ProductID, saleItem.ProductName,
			saleItem.ProductSKU, saleItem.ProductBarcode, saleItem.Quantity, saleItem.UnitPrice,
			saleItem.CostPrice, saleItem.DiscountType, saleItem.DiscountValue, saleItem.DiscountAmount,
			saleItem.TaxRate, saleItem.TaxAmount, saleItem.Subtotal, saleItem.Total, saleItem.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert sale item: %w", err)
		}

		// Deduct inventory
		_, err = tx.Exec(ctx, `
			UPDATE inventory_items SET quantity = $1, updated_at = $2 WHERE product_id = $3
		`, prepared.NewQty, now, saleItem.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to update inventory: %w", err)
		}

		// Record stock adjustment
		_, err = tx.Exec(ctx, `
			INSERT INTO stock_adjustments (
				id, product_id, adjustment_type, quantity_before, quantity_change, quantity_after,
				reason, reference_type, reference_id, adjusted_by, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			uuid.New(), saleItem.ProductID, "sale", prepared.ProductQty, saleItem.Quantity.Neg(), prepared.NewQty,
			fmt.Sprintf("Sale: %s", invoiceNo), "sale", sale.ID, req.EmployeeID, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to record stock adjustment: %w", err)
		}

		sale.Items = append(sale.Items, saleItem)
	}

	// Insert payments
	for _, paymentReq := range req.Payments {
		payment := models.Payment{
			ID:             uuid.New(),
			SaleID:         sale.ID,
			PaymentMethod:  paymentReq.PaymentMethod,
			Amount:         paymentReq.Amount,
			AmountTendered: paymentReq.AmountTendered,
			ChangeGiven:    paymentReq.ChangeGiven,
			CardType:       paymentReq.CardType,
			CardLastFour:   paymentReq.CardLastFour,
			ReferenceNo:    paymentReq.ReferenceNo,
			BankName:       paymentReq.BankName,
			AccountNo:      paymentReq.AccountNo,
			VoucherCode:    paymentReq.VoucherCode,
			Status:         models.PaymentRecordCompleted,
			Notes:          paymentReq.Notes,
			ProcessedBy:    &req.EmployeeID,
			CreatedAt:      now,
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO payments (
				id, sale_id, payment_method, amount, amount_tendered, change_given,
				card_type, card_last_four, reference_no, bank_name, account_no,
				voucher_code, status, notes, processed_by, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		`,
			payment.ID, payment.SaleID, payment.PaymentMethod, payment.Amount,
			payment.AmountTendered, payment.ChangeGiven, payment.CardType, payment.CardLastFour,
			payment.ReferenceNo, payment.BankName, payment.AccountNo, payment.VoucherCode,
			payment.Status, payment.Notes, payment.ProcessedBy, payment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert payment: %w", err)
		}

		sale.Payments = append(sale.Payments, payment)
	}

	// Update shift totals if shift is provided
	if req.ShiftID != nil {
		_, err = tx.Exec(ctx, `
			UPDATE shifts SET 
				total_sales = total_sales + $1,
				transaction_count = transaction_count + 1,
				updated_at = $2
			WHERE id = $3
		`, totalAmount, now, *req.ShiftID)
		if err != nil {
			return nil, fmt.Errorf("failed to update shift totals: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return sale, nil
}

// generateInvoiceNumber generates a unique invoice number
func (r *SaleRepository) generateInvoiceNumber(ctx context.Context, tx pgx.Tx, t time.Time) (string, error) {
	dateStr := t.Format("20060102")
	prefix := fmt.Sprintf("INV-%s-", dateStr)

	var count int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM sales WHERE invoice_no LIKE $1
	`, prefix+"%").Scan(&count)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%05d", prefix, count+1), nil
}

// GetByID retrieves a sale by ID with items and payments
func (r *SaleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Sale, error) {
	sale := &models.Sale{}

	err := r.pool.QueryRow(ctx, `
		SELECT 
			s.id, s.invoice_no, s.subtotal, s.tax_amount, s.discount_amount, s.total_amount,
			s.item_count, s.payment_status, s.amount_paid, s.change_amount, s.discount_type,
			s.discount_value, s.discount_reason, s.employee_id, s.shift_id, s.customer_name,
			s.customer_phone, s.status, s.voided_at, s.voided_by, s.void_reason, s.notes,
			s.created_at, s.updated_at, u.name as employee_name
		FROM sales s
		LEFT JOIN users u ON s.employee_id = u.id
		WHERE s.id = $1
	`, id).Scan(
		&sale.ID, &sale.InvoiceNo, &sale.Subtotal, &sale.TaxAmount, &sale.DiscountAmount,
		&sale.TotalAmount, &sale.ItemCount, &sale.PaymentStatus, &sale.AmountPaid, &sale.ChangeAmount,
		&sale.DiscountType, &sale.DiscountValue, &sale.DiscountReason, &sale.EmployeeID, &sale.ShiftID,
		&sale.CustomerName, &sale.CustomerPhone, &sale.Status, &sale.VoidedAt, &sale.VoidedBy,
		&sale.VoidReason, &sale.Notes, &sale.CreatedAt, &sale.UpdatedAt, &sale.EmployeeName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Get items
	itemRows, err := r.pool.Query(ctx, `
		SELECT 
			id, sale_id, product_id, product_name, product_sku, product_barcode,
			quantity, unit_price, cost_price, discount_type, discount_value, discount_amount,
			tax_rate, tax_amount, subtotal, total, is_refunded, refunded_quantity, created_at
		FROM sale_items
		WHERE sale_id = $1
		ORDER BY created_at
	`, id)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item models.SaleItem
		err := itemRows.Scan(
			&item.ID, &item.SaleID, &item.ProductID, &item.ProductName, &item.ProductSKU,
			&item.ProductBarcode, &item.Quantity, &item.UnitPrice, &item.CostPrice,
			&item.DiscountType, &item.DiscountValue, &item.DiscountAmount, &item.TaxRate,
			&item.TaxAmount, &item.Subtotal, &item.Total, &item.IsRefunded, &item.RefundedQuantity,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sale.Items = append(sale.Items, item)
	}

	// Get payments
	paymentRows, err := r.pool.Query(ctx, `
		SELECT 
			id, sale_id, payment_method, amount, amount_tendered, change_given,
			card_type, card_last_four, reference_no, bank_name, account_no,
			voucher_code, status, notes, processed_by, created_at
		FROM payments
		WHERE sale_id = $1
		ORDER BY created_at
	`, id)
	if err != nil {
		return nil, err
	}
	defer paymentRows.Close()

	for paymentRows.Next() {
		var payment models.Payment
		err := paymentRows.Scan(
			&payment.ID, &payment.SaleID, &payment.PaymentMethod, &payment.Amount,
			&payment.AmountTendered, &payment.ChangeGiven, &payment.CardType, &payment.CardLastFour,
			&payment.ReferenceNo, &payment.BankName, &payment.AccountNo, &payment.VoucherCode,
			&payment.Status, &payment.Notes, &payment.ProcessedBy, &payment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sale.Payments = append(sale.Payments, payment)
	}

	return sale, nil
}

// GetByInvoiceNo retrieves a sale by invoice number
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

// List retrieves sales with pagination and filters
func (r *SaleRepository) List(ctx context.Context, employeeID, shiftID *uuid.UUID, status *string, startDate, endDate *time.Time, limit, offset int) ([]models.Sale, int, error) {
	args := []interface{}{}
	argIndex := 1

	whereClause := "WHERE 1=1"

	if employeeID != nil {
		whereClause += fmt.Sprintf(" AND s.employee_id = $%d", argIndex)
		args = append(args, *employeeID)
		argIndex++
	}

	if shiftID != nil {
		whereClause += fmt.Sprintf(" AND s.shift_id = $%d", argIndex)
		args = append(args, *shiftID)
		argIndex++
	}

	if status != nil {
		whereClause += fmt.Sprintf(" AND s.status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	if startDate != nil {
		whereClause += fmt.Sprintf(" AND s.created_at >= $%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		whereClause += fmt.Sprintf(" AND s.created_at <= $%d", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sales s %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get sales
	query := fmt.Sprintf(`
		SELECT 
			s.id, s.invoice_no, s.subtotal, s.tax_amount, s.discount_amount, s.total_amount,
			s.item_count, s.payment_status, s.amount_paid, s.change_amount,
			s.employee_id, s.shift_id, s.customer_name, s.status, s.created_at, s.updated_at,
			u.name as employee_name,
			(SELECT payment_method FROM payments WHERE sale_id = s.id LIMIT 1) as payment_method
		FROM sales s
		LEFT JOIN users u ON s.employee_id = u.id
		%s
		ORDER BY s.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sales []models.Sale
	for rows.Next() {
		var s models.Sale
		var paymentMethod *string
		err := rows.Scan(
			&s.ID, &s.InvoiceNo, &s.Subtotal, &s.TaxAmount, &s.DiscountAmount, &s.TotalAmount,
			&s.ItemCount, &s.PaymentStatus, &s.AmountPaid, &s.ChangeAmount,
			&s.EmployeeID, &s.ShiftID, &s.CustomerName, &s.Status, &s.CreatedAt, &s.UpdatedAt,
			&s.EmployeeName,
			&paymentMethod,
		)
		if err != nil {
			return nil, 0, err
		}

		if paymentMethod != nil {
			s.Payments = []models.Payment{
				{PaymentMethod: models.PaymentMethod(*paymentMethod)},
			}
		}

		sales = append(sales, s)
	}

	return sales, total, nil
}

// VoidSale voids a sale and restores inventory
func (r *SaleRepository) VoidSale(ctx context.Context, saleID, voidedBy uuid.UUID, reason string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	// Get sale and verify it can be voided
	var status string
	var shiftID *uuid.UUID
	var totalAmount decimal.Decimal
	err = tx.QueryRow(ctx, `
		SELECT status, shift_id, total_amount FROM sales WHERE id = $1 FOR UPDATE
	`, saleID).Scan(&status, &shiftID, &totalAmount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("sale not found")
		}
		return err
	}

	if status != string(models.SaleStatusCompleted) {
		return fmt.Errorf("sale cannot be voided: current status is %s", status)
	}

	// Get sale items to restore inventory
	rows, err := tx.Query(ctx, `
		SELECT product_id, quantity FROM sale_items WHERE sale_id = $1
	`, saleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type itemRestore struct {
		ProductID uuid.UUID
		Quantity  decimal.Decimal
	}
	var itemsToRestore []itemRestore

	for rows.Next() {
		var ir itemRestore
		if err := rows.Scan(&ir.ProductID, &ir.Quantity); err != nil {
			return err
		}
		itemsToRestore = append(itemsToRestore, ir)
	}

	// Restore inventory for each item
	for _, item := range itemsToRestore {
		var currentQty decimal.Decimal
		err = tx.QueryRow(ctx, `
			SELECT quantity FROM inventory_items WHERE product_id = $1 FOR UPDATE
		`, item.ProductID).Scan(&currentQty)
		if err != nil {
			return fmt.Errorf("failed to get inventory: %w", err)
		}

		newQty := currentQty.Add(item.Quantity)
		_, err = tx.Exec(ctx, `
			UPDATE inventory_items SET quantity = $1, updated_at = $2 WHERE product_id = $3
		`, newQty, now, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to restore inventory: %w", err)
		}

		// Record adjustment
		_, err = tx.Exec(ctx, `
			INSERT INTO stock_adjustments (
				id, product_id, adjustment_type, quantity_before, quantity_change, quantity_after,
				reason, reference_type, reference_id, adjusted_by, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			uuid.New(), item.ProductID, "return", currentQty, item.Quantity, newQty,
			fmt.Sprintf("Void sale: %s", reason), "sale_void", saleID, voidedBy, now,
		)
		if err != nil {
			return fmt.Errorf("failed to record adjustment: %w", err)
		}
	}

	// Update sale status
	_, err = tx.Exec(ctx, `
		UPDATE sales SET 
			status = $1, payment_status = $2, voided_at = $3, voided_by = $4, void_reason = $5, updated_at = $6
		WHERE id = $7
	`, models.SaleStatusVoided, models.PaymentStatusVoided, now, voidedBy, reason, now, saleID)
	if err != nil {
		return fmt.Errorf("failed to update sale: %w", err)
	}

	// Update payments status
	_, err = tx.Exec(ctx, `
		UPDATE payments SET status = $1 WHERE sale_id = $2
	`, models.PaymentRecordRefunded, saleID)
	if err != nil {
		return fmt.Errorf("failed to update payments: %w", err)
	}

	// Update shift totals if applicable
	if shiftID != nil {
		_, err = tx.Exec(ctx, `
			UPDATE shifts SET 
				total_refunds = total_refunds + $1,
				refund_count = refund_count + 1,
				updated_at = $2
			WHERE id = $3
		`, totalAmount, now, *shiftID)
		if err != nil {
			return fmt.Errorf("failed to update shift: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetDailySummary gets sales summary for a date
func (r *SaleRepository) GetDailySummary(ctx context.Context, date time.Time) (map[string]interface{}, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var totalSales, totalTax, totalDiscount, totalAmount decimal.Decimal
	var transactionCount, itemCount int

	err := r.pool.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(subtotal), 0),
			COALESCE(SUM(tax_amount), 0),
			COALESCE(SUM(discount_amount), 0),
			COALESCE(SUM(total_amount), 0),
			COUNT(*),
			COALESCE(SUM(item_count), 0)
		FROM sales 
		WHERE created_at >= $1 AND created_at < $2 AND status = 'completed'
	`, startOfDay, endOfDay).Scan(&totalSales, &totalTax, &totalDiscount, &totalAmount, &transactionCount, &itemCount)
	if err != nil {
		return nil, err
	}

	// Get payment method breakdown
	rows, err := r.pool.Query(ctx, `
		SELECT p.payment_method, COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN sales s ON p.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at < $2 AND s.status = 'completed'
		GROUP BY p.payment_method
	`, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	paymentBreakdown := make(map[string]string)
	for rows.Next() {
		var method string
		var amount decimal.Decimal
		if err := rows.Scan(&method, &amount); err != nil {
			return nil, err
		}
		paymentBreakdown[method] = amount.String()
	}

	return map[string]interface{}{
		"date":              date.Format("2006-01-02"),
		"total_sales":       totalSales.String(),
		"total_tax":         totalTax.String(),
		"total_discount":    totalDiscount.String(),
		"total_amount":      totalAmount.String(),
		"transaction_count": transactionCount,
		"item_count":        itemCount,
		"payment_breakdown": paymentBreakdown,
	}, nil
}
