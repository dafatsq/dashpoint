package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

func expenseInventoryReason(action string, expenseID uuid.UUID) *string {
	return stringPtr(fmt.Sprintf("Expense inventory %s %s", action, expenseID.String()))
}

func (h *ExpenseHandler) isInventoryPurchaseCategory(ctx context.Context, categoryID *uuid.UUID) (bool, error) {
	return h.isInventoryPurchaseCategoryWithArchived(ctx, categoryID, false)
}

func (h *ExpenseHandler) isInventoryPurchaseCategoryWithArchived(ctx context.Context, categoryID *uuid.UUID, allowArchived bool) (bool, error) {
	if categoryID == nil {
		return false, nil
	}

	category, err := h.repo.GetCategoryByID(ctx, *categoryID)
	if err != nil {
		return false, fmt.Errorf("Failed to load expense category")
	}
	if category == nil {
		return false, fmt.Errorf("Invalid category ID")
	}
	if !category.IsActive && !allowArchived {
		return false, fmt.Errorf("Archived expense categories cannot be used")
	}

	return isInventoryPurchaseExpenseCategory(category), nil
}

func (h *ExpenseHandler) ensureExpenseInventoryProductActive(ctx context.Context, productID *uuid.UUID) error {
	if productID == nil {
		return nil
	}
	if h.productRepo == nil {
		return fmt.Errorf("Failed to load product")
	}

	product, err := h.productRepo.GetByID(ctx, *productID)
	if err != nil {
		return fmt.Errorf("Failed to load product")
	}
	if product == nil {
		return fmt.Errorf("Product not found")
	}
	if !product.IsActive {
		return fmt.Errorf("Archived products cannot be changed")
	}

	return nil
}

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

	isInventoryPurchase, err := h.isInventoryPurchaseCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if isInventoryPurchase {
		if productID == nil {
			return nil, fmt.Errorf("Product is required for Inventory Purchase")
		}
		if quantity == nil {
			return nil, fmt.Errorf("Quantity is required for Inventory Purchase")
		}
		if err := h.ensureExpenseInventoryProductActive(ctx, productID); err != nil {
			return nil, err
		}
	} else {
		productID = nil
		quantity = nil
	}

	return &models.Expense{
		CategoryID:       categoryID,
		ProductID:        productID,
		Quantity:         quantity,
		AppliesInventory: isInventoryPurchase && req.AppliesInventory,
		Amount:           amount,
		Description:      req.Description,
		ExpenseDate:      expenseDate,
		Vendor:           req.Vendor,
		ReferenceNumber:  req.ReferenceNumber,
		Notes:            req.Notes,
		CreatedBy:        userID,
	}, nil
}

func (h *ExpenseHandler) createExpense(ctx context.Context, expense *models.Expense, userID uuid.UUID) (*models.Expense, error) {
	if !expense.AppliesInventory || expense.ProductID == nil || expense.Quantity == nil {
		return h.repo.Create(ctx, expense)
	}
	if err := h.ensureExpenseInventoryProductActive(ctx, expense.ProductID); err != nil {
		return nil, err
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
		expenseInventoryReason("purchase", created.ID),
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

func (h *ExpenseHandler) syncExpenseInventory(ctx context.Context, tx pgx.Tx, expenseID uuid.UUID, userID uuid.UUID, existing *models.Expense, finalAppliesInventory bool, finalProductID *uuid.UUID, finalQuantity *decimal.Decimal) error {
	oldAppliesInventory := existing.AppliesInventory && existing.ProductID != nil && existing.Quantity != nil
	newAppliesInventory := finalAppliesInventory && finalProductID != nil && finalQuantity != nil

	inventoryChanged := false
	switch {
	case oldAppliesInventory && !newAppliesInventory:
		inventoryChanged = true
	case !oldAppliesInventory && newAppliesInventory:
		inventoryChanged = true
	case oldAppliesInventory && newAppliesInventory && *existing.ProductID != *finalProductID:
		inventoryChanged = true
	case oldAppliesInventory && newAppliesInventory:
		if existing.Quantity != nil && finalQuantity != nil && !existing.Quantity.Equal(*finalQuantity) {
			inventoryChanged = true
		}
	}

	if !inventoryChanged {
		return nil
	}

	isSameProduct := oldAppliesInventory && newAppliesInventory && *existing.ProductID == *finalProductID
	if isSameProduct {
		if err := h.ensureExpenseInventoryProductActive(ctx, finalProductID); err != nil {
			return err
		}
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
			expenseInventoryReason("quantity update", expenseID),
			stringPtr("expense"),
			&expenseID,
			userID,
		)
		if err != nil {
			return fmt.Errorf("Failed to update inventory: %s", err.Error())
		}
		return nil
	}

	if oldAppliesInventory {
		if err := h.ensureExpenseInventoryProductActive(ctx, existing.ProductID); err != nil {
			return err
		}
		_, err := h.inventoryRepo.AdjustStockWithTx(
			ctx,
			tx,
			*existing.ProductID,
			models.AdjustmentPurchase,
			existing.Quantity.Neg(),
			expenseInventoryReason("revert", expenseID),
			stringPtr("expense"),
			&expenseID,
			userID,
		)
		if err != nil {
			return fmt.Errorf("Failed to revert old inventory: %s", err.Error())
		}
	}

	if newAppliesInventory {
		if err := h.ensureExpenseInventoryProductActive(ctx, finalProductID); err != nil {
			return err
		}
		_, err := h.inventoryRepo.AdjustStockWithTx(
			ctx,
			tx,
			*finalProductID,
			models.AdjustmentPurchase,
			*finalQuantity,
			expenseInventoryReason("apply", expenseID),
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
