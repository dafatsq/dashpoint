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

type salePreparedItem struct {
	SaleItem   models.SaleItem
	ProductQty decimal.Decimal
	NewQty     decimal.Decimal
}

// Create creates a new sale with items and payments in a single transaction.
func (r *SaleRepository) Create(ctx context.Context, req *CreateSaleRequest) (*models.Sale, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	invoiceNo, err := r.generateInvoiceNumber(ctx, tx, now)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	preparedItems, subtotal, taxAmount, itemDiscountAmount, err := prepareSaleItems(ctx, tx, req.Items, now)
	if err != nil {
		return nil, err
	}

	discountAmount := calculateSaleDiscount(subtotal, itemDiscountAmount, req.DiscountType, req.DiscountValue)
	totalAmount := subtotal.Add(taxAmount).Sub(discountAmount)
	if err := validateSaleFinancialIntegrity(totalAmount, req.Payments); err != nil {
		return nil, err
	}
	amountPaid, paymentStatus, changeAmount := calculateSalePaymentStatus(totalAmount, req.Payments)

	sale := &models.Sale{
		ID:             uuid.New(),
		InvoiceNo:      invoiceNo,
		Subtotal:       subtotal,
		TaxAmount:      taxAmount,
		DiscountAmount: discountAmount,
		TotalAmount:    totalAmount,
		ItemCount:      len(req.Items),
		PaymentStatus:  paymentStatus,
		AmountPaid:     amountPaid,
		ChangeAmount:   changeAmount,
		EmployeeID:     req.EmployeeID,
		ShiftID:        req.ShiftID,
		DiscountType:   req.DiscountType,
		DiscountValue:  req.DiscountValue,
		DiscountReason: req.DiscountReason,
		Notes:          req.Notes,
		Status:         models.SaleStatusCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := insertSaleTx(ctx, tx, sale); err != nil {
		return nil, fmt.Errorf("failed to insert sale: %w", err)
	}
	if err := insertSaleItemsAndAdjustInventoryTx(ctx, tx, sale, preparedItems, req.EmployeeID, invoiceNo, now); err != nil {
		return nil, err
	}
	if err := insertSalePaymentsTx(ctx, tx, sale, req.Payments, req.EmployeeID, now); err != nil {
		return nil, err
	}
	if req.ShiftID != nil {
		if err := updateShiftSalesTotalsTx(ctx, tx, *req.ShiftID, totalAmount, now); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return sale, nil
}

// ValidateCart verifies live product state for a cart without creating a sale.
func (r *SaleRepository) ValidateCart(ctx context.Context, req *ValidateSaleCartRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, _, _, _, err := prepareSaleItems(ctx, tx, req.Items, time.Now()); err != nil {
		return err
	}

	return nil
}

func prepareSaleItems(ctx context.Context, tx pgx.Tx, items []CreateSaleItemRequest, now time.Time) ([]salePreparedItem, decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	var subtotal, taxAmount, itemDiscountAmount decimal.Decimal
	preparedItems := make([]salePreparedItem, 0, len(items))

	for i := range items {
		item := &items[i]
		product, err := loadSaleProductForUpdate(ctx, tx, item.ProductID)
		if err != nil {
			return nil, decimal.Zero, decimal.Zero, decimal.Zero, err
		}
		if err := validateSaleItemUnitPrice(product, item); err != nil {
			return nil, decimal.Zero, decimal.Zero, decimal.Zero, err
		}

		productQty := product.Inventory.Quantity
		if productQty.LessThan(item.Quantity) {
			return nil, decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("insufficient stock for %s: available %s, requested %s", product.Name, productQty.String(), item.Quantity.String())
		}

		itemSubtotal := product.Price.Mul(item.Quantity)
		itemTax := itemSubtotal.Mul(product.TaxRate).Div(decimal.NewFromInt(100))
		if err := validateSaleItemDiscount(product, item, itemSubtotal, itemTax); err != nil {
			return nil, decimal.Zero, decimal.Zero, decimal.Zero, err
		}

		prepared := buildPreparedSaleItem(product, item, productQty, now)
		preparedItems = append(preparedItems, prepared)

		subtotal = subtotal.Add(prepared.SaleItem.Subtotal)
		taxAmount = taxAmount.Add(prepared.SaleItem.TaxAmount)
		itemDiscountAmount = itemDiscountAmount.Add(item.DiscountAmount)
	}

	return preparedItems, subtotal, taxAmount, itemDiscountAmount, nil
}

func validateSaleItemUnitPrice(product *models.Product, item *CreateSaleItemRequest) error {
	if item.UnitPrice.Equal(product.Price) {
		return nil
	}
	return fmt.Errorf("product price changed for %s: current price is %s, submitted price is %s", product.Name, product.Price.String(), item.UnitPrice.String())
}

func validateSaleItemDiscount(product *models.Product, item *CreateSaleItemRequest, itemSubtotal decimal.Decimal, itemTax decimal.Decimal) error {
	if item.DiscountValue.LessThan(decimal.Zero) || item.DiscountAmount.LessThan(decimal.Zero) {
		return fmt.Errorf("discount cannot be negative for %s", product.Name)
	}
	if item.DiscountAmount.GreaterThan(itemSubtotal.Add(itemTax)) {
		return fmt.Errorf("discount exceeds item total for %s", product.Name)
	}
	return nil
}

func buildPreparedSaleItem(product *models.Product, item *CreateSaleItemRequest, productQty decimal.Decimal, now time.Time) salePreparedItem {
	unitPrice := product.Price
	itemSubtotal := unitPrice.Mul(item.Quantity)
	itemTax := itemSubtotal.Mul(product.TaxRate).Div(decimal.NewFromInt(100))
	itemTotal := itemSubtotal.Add(itemTax).Sub(item.DiscountAmount)

	return salePreparedItem{
		SaleItem: models.SaleItem{
			ID:             uuid.New(),
			ProductID:      product.ID,
			ProductName:    product.Name,
			ProductSKU:     product.SKU,
			ProductBarcode: product.Barcode,
			Quantity:       item.Quantity,
			UnitPrice:      unitPrice,
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
		ProductQty: productQty,
		NewQty:     productQty.Sub(item.Quantity),
	}
}

func loadSaleProductForUpdate(ctx context.Context, tx pgx.Tx, productID uuid.UUID) (*models.Product, error) {
	var product models.Product
	var qty decimal.Decimal
	err := tx.QueryRow(ctx, `
		SELECT p.id, p.name, p.sku, p.barcode, p.price, p.cost, p.tax_rate, COALESCE(i.quantity, 0)
		FROM products p
		LEFT JOIN inventory_items i ON p.id = i.product_id
		WHERE p.id = $1 AND p.is_active = true
		FOR UPDATE OF p
	`, productID).Scan(
		&product.ID, &product.Name, &product.SKU, &product.Barcode,
		&product.Price, &product.Cost, &product.TaxRate, &qty,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found: %s", productID)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	product.Inventory = &models.InventoryItem{Quantity: qty}
	return &product, nil
}

func calculateSaleDiscount(subtotal, itemDiscountAmount decimal.Decimal, discountType *string, discountValue *decimal.Decimal) decimal.Decimal {
	discountAmount := itemDiscountAmount
	if discountType == nil || discountValue == nil {
		return discountAmount
	}
	if *discountType == "percentage" {
		return discountAmount.Add(subtotal.Mul(*discountValue).Div(decimal.NewFromInt(100)))
	}
	return discountAmount.Add(*discountValue)
}

func calculateSalePaymentStatus(totalAmount decimal.Decimal, payments []CreatePaymentRequest) (decimal.Decimal, models.PaymentStatus, decimal.Decimal) {
	amountPaid := decimal.Zero
	for _, payment := range payments {
		amountPaid = amountPaid.Add(payment.Amount)
	}

	paymentStatus := models.PaymentStatusPending
	changeAmount := decimal.Zero
	if amountPaid.GreaterThanOrEqual(totalAmount) {
		paymentStatus = models.PaymentStatusPaid
		changeAmount = amountPaid.Sub(totalAmount)
	} else if amountPaid.GreaterThan(decimal.Zero) {
		paymentStatus = models.PaymentStatusPartial
	}

	return amountPaid, paymentStatus, changeAmount
}

func validateSaleFinancialIntegrity(totalAmount decimal.Decimal, payments []CreatePaymentRequest) error {
	if totalAmount.LessThan(decimal.Zero) {
		return fmt.Errorf("sale total cannot be negative")
	}

	amountPaid := decimal.Zero
	for _, payment := range payments {
		amountPaid = amountPaid.Add(payment.Amount)
		if err := validateSalePaymentTender(payment); err != nil {
			return err
		}
	}
	if !amountPaid.Equal(totalAmount) {
		return fmt.Errorf("payment amount does not match sale total")
	}
	return nil
}

func validateSalePaymentTender(payment CreatePaymentRequest) error {
	if payment.PaymentMethod != models.PaymentMethodCash {
		if payment.AmountTendered != nil || payment.ChangeGiven != nil {
			return fmt.Errorf("amount tendered and change are only valid for cash payments")
		}
		return nil
	}

	if payment.AmountTendered == nil {
		if payment.ChangeGiven != nil && !payment.ChangeGiven.Equal(decimal.Zero) {
			return fmt.Errorf("cash change requires amount tendered")
		}
		return nil
	}

	if payment.AmountTendered.LessThan(payment.Amount) {
		return fmt.Errorf("cash amount tendered is less than payment amount")
	}
	expectedChange := payment.AmountTendered.Sub(payment.Amount)
	actualChange := decimal.Zero
	if payment.ChangeGiven != nil {
		actualChange = *payment.ChangeGiven
	}
	if !actualChange.Equal(expectedChange) {
		return fmt.Errorf("cash change does not match amount tendered")
	}
	return nil
}

func insertSaleTx(ctx context.Context, tx pgx.Tx, sale *models.Sale) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO sales (
			id, invoice_no, subtotal, tax_amount, discount_amount, total_amount, item_count,
			payment_status, amount_paid, change_amount, discount_type, discount_value, discount_reason,
			employee_id, shift_id, status, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`,
		sale.ID, sale.InvoiceNo, sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount,
		sale.ItemCount, sale.PaymentStatus, sale.AmountPaid, sale.ChangeAmount, sale.DiscountType,
		sale.DiscountValue, sale.DiscountReason, sale.EmployeeID, sale.ShiftID, sale.Status,
		sale.Notes, sale.CreatedAt, sale.UpdatedAt,
	)
	return err
}

func insertSaleItemsAndAdjustInventoryTx(ctx context.Context, tx pgx.Tx, sale *models.Sale, preparedItems []salePreparedItem, employeeID uuid.UUID, invoiceNo string, now time.Time) error {
	for _, prepared := range preparedItems {
		saleItem := prepared.SaleItem
		saleItem.SaleID = sale.ID

		if err := insertSaleItemTx(ctx, tx, saleItem); err != nil {
			return fmt.Errorf("failed to insert sale item: %w", err)
		}
		if err := setInventoryQuantityTx(ctx, tx, saleItem.ProductID, prepared.NewQty, now); err != nil {
			return fmt.Errorf("failed to update inventory: %w", err)
		}
		if err := insertStockAdjustmentTx(ctx, tx, stockAdjustmentRecord{
			ProductID:      saleItem.ProductID,
			AdjustmentType: "sale",
			QuantityBefore: prepared.ProductQty,
			QuantityChange: saleItem.Quantity.Neg(),
			QuantityAfter:  prepared.NewQty,
			Reason:         saleInventoryReason(invoiceNo),
			ReferenceType:  "sale",
			ReferenceID:    sale.ID,
			AdjustedBy:     employeeID,
			CreatedAt:      now,
		}); err != nil {
			return fmt.Errorf("failed to record stock adjustment: %w", err)
		}

		sale.Items = append(sale.Items, saleItem)
	}
	return nil
}

func insertSalePaymentsTx(ctx context.Context, tx pgx.Tx, sale *models.Sale, paymentReqs []CreatePaymentRequest, employeeID uuid.UUID, now time.Time) error {
	for _, paymentReq := range paymentReqs {
		payment := models.Payment{
			ID:             uuid.New(),
			SaleID:         sale.ID,
			PaymentMethod:  paymentReq.PaymentMethod,
			Amount:         paymentReq.Amount,
			AmountTendered: paymentReq.AmountTendered,
			ChangeGiven:    paymentReq.ChangeGiven,
			ReferenceNo:    paymentReq.ReferenceNo,
			Status:         models.PaymentRecordCompleted,
			CreatedAt:      now,
		}
		if err := insertPaymentTx(ctx, tx, payment); err != nil {
			return fmt.Errorf("failed to insert payment: %w", err)
		}
		sale.Payments = append(sale.Payments, payment)
	}
	return nil
}

func updateShiftSalesTotalsTx(ctx context.Context, tx pgx.Tx, shiftID uuid.UUID, totalAmount decimal.Decimal, now time.Time) error {
	_, err := tx.Exec(ctx, `
		UPDATE shifts SET
			total_sales = total_sales + $1,
			transaction_count = transaction_count + 1,
			updated_at = $2
		WHERE id = $3
	`, totalAmount, now, shiftID)
	if err != nil {
		return fmt.Errorf("failed to update shift totals: %w", err)
	}
	return nil
}
