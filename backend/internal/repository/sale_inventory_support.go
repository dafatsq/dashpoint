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

type stockAdjustmentRecord struct {
	ProductID      uuid.UUID
	AdjustmentType string
	QuantityBefore decimal.Decimal
	QuantityChange decimal.Decimal
	QuantityAfter  decimal.Decimal
	Reason         string
	ReferenceType  string
	ReferenceID    uuid.UUID
	AdjustedBy     uuid.UUID
	CreatedAt      time.Time
}

func saleInventoryReason(invoiceNo string) string {
	return fmt.Sprintf("Sale invoice %s", invoiceNo)
}

func voidedSaleInventoryReason(invoiceNo string) string {
	return fmt.Sprintf("Voided sale invoice %s", invoiceNo)
}

func generateInvoiceNumber(ctx context.Context, tx pgx.Tx, t time.Time) (string, error) {
	dateStr := t.Format("20060102")
	prefix := "INV-" + dateStr + "-"

	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM sales WHERE invoice_no LIKE $1`, prefix+"%").Scan(&count); err != nil {
		return "", err
	}
	return prefix + fmt.Sprintf("%05d", count+1), nil
}

func (r *SaleRepository) generateInvoiceNumber(ctx context.Context, tx pgx.Tx, t time.Time) (string, error) {
	return generateInvoiceNumber(ctx, tx, t)
}

func insertSaleItemTx(ctx context.Context, tx pgx.Tx, saleItem models.SaleItem) error {
	_, err := tx.Exec(ctx, `
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
	return err
}

func insertPaymentTx(ctx context.Context, tx pgx.Tx, payment models.Payment) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO payments (
			id, sale_id, payment_method, amount, amount_tendered, change_given,
			reference_no, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		payment.ID, payment.SaleID, payment.PaymentMethod, payment.Amount,
		payment.AmountTendered, payment.ChangeGiven, payment.ReferenceNo,
		payment.Status, payment.CreatedAt,
	)
	return err
}

func getInventoryQuantityForUpdateTx(ctx context.Context, tx pgx.Tx, productID uuid.UUID) (decimal.Decimal, error) {
	var quantity decimal.Decimal
	err := tx.QueryRow(ctx, `SELECT quantity FROM inventory_items WHERE product_id = $1 FOR UPDATE`, productID).Scan(&quantity)
	return quantity, err
}

func setInventoryQuantityTx(ctx context.Context, tx pgx.Tx, productID uuid.UUID, quantity decimal.Decimal, updatedAt time.Time) error {
	_, err := tx.Exec(ctx, `UPDATE inventory_items SET quantity = $1, updated_at = $2 WHERE product_id = $3`, quantity, updatedAt, productID)
	return err
}

func insertStockAdjustmentTx(ctx context.Context, tx pgx.Tx, record stockAdjustmentRecord) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO stock_adjustments (
			id, product_id, adjustment_type, quantity_before, quantity_change, quantity_after,
			reason, reference_type, reference_id, adjusted_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		uuid.New(), record.ProductID, record.AdjustmentType, record.QuantityBefore, record.QuantityChange,
		record.QuantityAfter, record.Reason, record.ReferenceType, record.ReferenceID, record.AdjustedBy, record.CreatedAt,
	)
	return err
}
