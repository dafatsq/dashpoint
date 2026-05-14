package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

func (h *ExpenseHandler) createExpenseModel(ctx context.Context, req CreateExpenseRequest, userID uuid.UUID) (*models.Expense, error) {
	amount, err := parseRequiredAmount(req.Amount)
	if err != nil {
		return nil, err
	}
	expenseDate, err := parseExpenseDateField(req.ExpenseDate, "Invalid expense date format (use YYYY-MM-DD)")
	if err != nil {
		return nil, err
	}

	categoryID, err := parseOptionalExpenseUUIDField(req.CategoryID, "Invalid category ID")
	if err != nil {
		return nil, err
	}
	productID, err := parseOptionalExpenseUUIDField(req.ProductID, "Invalid product ID")
	if err != nil {
		return nil, err
	}
	quantity, err := parseOptionalExpensePositiveDecimalField(req.Quantity, "Invalid quantity")
	if err != nil {
		return nil, err
	}

	return &models.Expense{
		CategoryID:      categoryID,
		ProductID:       productID,
		Quantity:        quantity,
		Amount:          amount,
		Description:     req.Description,
		ExpenseDate:     expenseDate,
		Vendor:          req.Vendor,
		ReferenceNumber: req.ReferenceNumber,
		Notes:           req.Notes,
		CreatedBy:       userID,
	}, nil
}

func (h *ExpenseHandler) createExpense(ctx context.Context, expense *models.Expense, userID uuid.UUID) (*models.Expense, error) {
	if expense.ProductID == nil || expense.Quantity == nil {
		return h.repo.Create(ctx, expense)
	}

	tx, err := h.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to create expense")
	}
	defer tx.Rollback(ctx)

	created, err := h.repo.CreateWithTx(ctx, tx, expense)
	if err != nil {
		return nil, fmt.Errorf("Failed to create expense")
	}

	if _, err := h.inventoryRepo.AdjustStockWithTx(
		ctx,
		tx,
		*created.ProductID,
		models.AdjustmentPurchase,
		*created.Quantity,
		stringPtr("Inventory purchase - Expense ID: "+created.ID.String()),
		stringPtr("expense"),
		&created.ID,
		userID,
	); err != nil {
		return nil, fmt.Errorf("Failed to adjust inventory: %s", err.Error())
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("Failed to create expense")
	}

	return created, nil
}

func (h *ExpenseHandler) syncExpenseInventory(ctx context.Context, tx pgx.Tx, expenseID uuid.UUID, userID uuid.UUID, existing *models.Expense, finalProductID *uuid.UUID, finalQuantity *decimal.Decimal) error {
	inventoryChanged := false
	switch {
	case existing.ProductID != nil && finalProductID == nil:
		inventoryChanged = true
	case existing.ProductID == nil && finalProductID != nil:
		inventoryChanged = true
	case existing.ProductID != nil && finalProductID != nil && *existing.ProductID != *finalProductID:
		inventoryChanged = true
	case finalProductID != nil:
		if existing.Quantity != nil && finalQuantity != nil && !existing.Quantity.Equal(*finalQuantity) {
			inventoryChanged = true
		} else if (existing.Quantity == nil) != (finalQuantity == nil) {
			inventoryChanged = true
		}
	}

	if !inventoryChanged {
		return nil
	}

	isSameProduct := existing.ProductID != nil && finalProductID != nil && *existing.ProductID == *finalProductID
	if isSameProduct {
		oldQuantity := decimal.Zero
		if existing.Quantity != nil {
			oldQuantity = *existing.Quantity
		}
		newQuantity := decimal.Zero
		if finalQuantity != nil {
			newQuantity = *finalQuantity
		}
		delta := newQuantity.Sub(oldQuantity)
		if delta.IsZero() {
			return nil
		}
		_, err := h.inventoryRepo.AdjustStockWithTx(
			ctx,
			tx,
			*finalProductID,
			models.AdjustmentPurchase,
			delta,
			stringPtr("Update purchase qty - Expense: "+expenseID.String()),
			stringPtr("expense"),
			&expenseID,
			userID,
		)
		if err != nil {
			return fmt.Errorf("Failed to update inventory: %s", err.Error())
		}
		return nil
	}

	if existing.ProductID != nil && existing.Quantity != nil {
		_, err := h.inventoryRepo.AdjustStockWithTx(
			ctx,
			tx,
			*existing.ProductID,
			models.AdjustmentPurchase,
			existing.Quantity.Neg(),
			stringPtr("Revert purchase (edit) - Expense: "+expenseID.String()),
			stringPtr("expense"),
			&expenseID,
			userID,
		)
		if err != nil {
			return fmt.Errorf("Failed to revert old inventory: %s", err.Error())
		}
	}

	if finalProductID != nil && finalQuantity != nil {
		_, err := h.inventoryRepo.AdjustStockWithTx(
			ctx,
			tx,
			*finalProductID,
			models.AdjustmentPurchase,
			*finalQuantity,
			stringPtr("Apply purchase (edit) - Expense: "+expenseID.String()),
			stringPtr("expense"),
			&expenseID,
			userID,
		)
		if err != nil {
			return fmt.Errorf("Failed to apply new inventory: %s", err.Error())
		}
	}

	return nil
}
