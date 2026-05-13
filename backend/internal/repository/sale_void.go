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

// VoidSale voids a sale and restores inventory.
func (r *SaleRepository) VoidSale(ctx context.Context, saleID, voidedBy uuid.UUID, reason string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	status, shiftID, totalAmount, err := loadVoidableSaleTx(ctx, tx, saleID)
	if err != nil {
		return err
	}
	if status != string(models.SaleStatusCompleted) {
		return fmt.Errorf("sale cannot be voided: current status is %s", status)
	}

	itemsToRestore, err := loadSaleRestoreItemsTx(ctx, tx, saleID)
	if err != nil {
		return err
	}
	if err := restoreSaleInventoryTx(ctx, tx, itemsToRestore, saleID, voidedBy, reason, now); err != nil {
		return err
	}
	if err := updateVoidedSaleTx(ctx, tx, saleID, voidedBy, reason, now); err != nil {
		return err
	}
	if err := refundPaymentsTx(ctx, tx, saleID); err != nil {
		return err
	}
	if shiftID != nil {
		if err := updateShiftRefundTotalsTx(ctx, tx, *shiftID, totalAmount, now); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type saleRestoreItem struct {
	ProductID uuid.UUID
	Quantity  decimal.Decimal
}

func loadVoidableSaleTx(ctx context.Context, tx pgx.Tx, saleID uuid.UUID) (string, *uuid.UUID, decimal.Decimal, error) {
	var status string
	var shiftID *uuid.UUID
	var totalAmount decimal.Decimal
	err := tx.QueryRow(ctx, `
		SELECT status, shift_id, total_amount FROM sales WHERE id = $1 FOR UPDATE
	`, saleID).Scan(&status, &shiftID, &totalAmount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil, decimal.Zero, fmt.Errorf("sale not found")
		}
		return "", nil, decimal.Zero, err
	}
	return status, shiftID, totalAmount, nil
}

func loadSaleRestoreItemsTx(ctx context.Context, tx pgx.Tx, saleID uuid.UUID) ([]saleRestoreItem, error) {
	rows, err := tx.Query(ctx, `SELECT product_id, quantity FROM sale_items WHERE sale_id = $1`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []saleRestoreItem
	for rows.Next() {
		var item saleRestoreItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func restoreSaleInventoryTx(ctx context.Context, tx pgx.Tx, items []saleRestoreItem, saleID, voidedBy uuid.UUID, reason string, now time.Time) error {
	for _, item := range items {
		currentQty, err := getInventoryQuantityForUpdateTx(ctx, tx, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get inventory: %w", err)
		}
		newQty := currentQty.Add(item.Quantity)
		if err := setInventoryQuantityTx(ctx, tx, item.ProductID, newQty, now); err != nil {
			return fmt.Errorf("failed to restore inventory: %w", err)
		}
		if err := insertStockAdjustmentTx(ctx, tx, stockAdjustmentRecord{
			ProductID:      item.ProductID,
			AdjustmentType: "return",
			QuantityBefore: currentQty,
			QuantityChange: item.Quantity,
			QuantityAfter:  newQty,
			Reason:         fmt.Sprintf("Void sale: %s", reason),
			ReferenceType:  "sale_void",
			ReferenceID:    saleID,
			AdjustedBy:     voidedBy,
			CreatedAt:      now,
		}); err != nil {
			return fmt.Errorf("failed to record adjustment: %w", err)
		}
	}
	return nil
}

func updateVoidedSaleTx(ctx context.Context, tx pgx.Tx, saleID, voidedBy uuid.UUID, reason string, now time.Time) error {
	_, err := tx.Exec(ctx, `
		UPDATE sales SET
			status = $1, payment_status = $2, voided_at = $3, voided_by = $4, void_reason = $5, updated_at = $6
		WHERE id = $7
	`, models.SaleStatusVoided, models.PaymentStatusVoided, now, voidedBy, reason, now, saleID)
	if err != nil {
		return fmt.Errorf("failed to update sale: %w", err)
	}
	return nil
}

func refundPaymentsTx(ctx context.Context, tx pgx.Tx, saleID uuid.UUID) error {
	_, err := tx.Exec(ctx, `UPDATE payments SET status = $1 WHERE sale_id = $2`, models.PaymentRecordRefunded, saleID)
	if err != nil {
		return fmt.Errorf("failed to update payments: %w", err)
	}
	return nil
}

func updateShiftRefundTotalsTx(ctx context.Context, tx pgx.Tx, shiftID uuid.UUID, totalAmount decimal.Decimal, now time.Time) error {
	_, err := tx.Exec(ctx, `
		UPDATE shifts SET
			total_refunds = total_refunds + $1,
			refund_count = refund_count + 1,
			updated_at = $2
		WHERE id = $3
	`, totalAmount, now, shiftID)
	if err != nil {
		return fmt.Errorf("failed to update shift: %w", err)
	}
	return nil
}
