package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

// List handles GET /api/v1/expenses
func (h *ExpenseHandler) List(c *fiber.Ctx) error {
	limit, offset, err := parseExpensePagination(c)
	if err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}

	categoryID, err := parseOptionalExpenseUUIDField(stringPointerFromQuery(c, "category_id"), "Invalid category ID")
	if err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}

	var startDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		parsed, parseErr := parseExpenseDateField(startStr, "Invalid start date format (use YYYY-MM-DD)")
		if parseErr != nil {
			return expenseMessage(c, fiber.StatusBadRequest, parseErr.Error())
		}
		startDate = &parsed
	}

	var endDate *time.Time
	if endStr := c.Query("end_date"); endStr != "" {
		parsed, parseErr := parseExpenseDateField(endStr, "Invalid end date format (use YYYY-MM-DD)")
		if parseErr != nil {
			return expenseMessage(c, fiber.StatusBadRequest, parseErr.Error())
		}
		endDate = &parsed
	}

	expenses, total, repoErr := h.repo.List(c.Context(), categoryID, startDate, endDate, limit, offset)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to list expenses")
	}

	response := make([]ExpenseResponse, 0, len(expenses))
	for i := range expenses {
		response = append(response, expenseToResponse(&expenses[i]))
	}

	return c.JSON(fiber.Map{
		"data":   response,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Create handles POST /api/v1/expenses
func (h *ExpenseHandler) Create(c *fiber.Ctx) error {
	var req CreateExpenseRequest
	if err := parseExpenseBody(c, &req); err != nil {
		return err
	}
	if err := validateCreateExpenseRequest(&req); err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}

	userID := middleware.GetUserID(c)
	expense, err := h.createExpenseModel(c.Context(), req, userID)
	if err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}
	if expense.ProductID != nil {
		if ok, staleErr := h.requireExpenseProductCurrent(c, *expense.ProductID, req.ExpectedProductUpdatedAt); !ok {
			return staleErr
		}
	}

	created, err := h.createExpense(c.Context(), expense, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if len(err.Error()) >= len("Failed to adjust inventory") && err.Error()[:26] == "Failed to adjust inventory" {
			status = fiber.StatusBadRequest
		}
		return expenseMessage(c, status, err.Error())
	}

	newValues := expenseAuditValues(created)
	if created.CategoryID != nil {
		categories, repoErr := h.repo.ListCategories(c.Context(), "all")
		if repoErr == nil {
			if categoryName := categoryNameByID(categories, *created.CategoryID); categoryName != nil {
				newValues["category"] = *categoryName
			}
		}
	}
	audit.LogWithValues(c, models.AuditActionExpenseCreate, models.AuditEntityExpense, created.ID.String(), "Created expense: "+created.Description, nil, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": expenseToResponse(created)})
}

// Get handles GET /api/v1/expenses/:id
func (h *ExpenseHandler) Get(c *fiber.Ctx) error {
	id, err := parseExpenseParamID(c, "id", "Invalid expense ID")
	if err != nil {
		return err
	}

	expense, repoErr := h.repo.GetByID(c.Context(), id)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to get expense")
	}
	if expense == nil {
		return expenseMessage(c, fiber.StatusNotFound, "Expense not found")
	}

	return c.JSON(fiber.Map{"data": expenseToResponse(expense)})
}

// Update handles PATCH /api/v1/expenses/:id
func (h *ExpenseHandler) Update(c *fiber.Ctx) error {
	id, err := parseExpenseParamID(c, "id", "Invalid expense ID")
	if err != nil {
		return err
	}

	userID := middleware.GetUserID(c)

	var req UpdateExpenseRequest
	if err := parseExpenseBody(c, &req); err != nil {
		return err
	}
	if err := validateUpdateExpenseRequest(&req); err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}

	if req.CategoryID != nil && *req.CategoryID != "" {
		if _, parseErr := parseOptionalExpenseUUIDField(req.CategoryID, "Invalid category ID"); parseErr != nil {
			return expenseMessage(c, fiber.StatusBadRequest, parseErr.Error())
		}
	}
	if req.ProductID != nil && *req.ProductID != "" {
		if _, parseErr := parseOptionalExpenseUUIDField(req.ProductID, "Invalid product ID"); parseErr != nil {
			return expenseMessage(c, fiber.StatusBadRequest, parseErr.Error())
		}
	}
	if req.Quantity != nil && *req.Quantity != "" {
		if _, parseErr := parseOptionalExpensePositiveDecimalField(req.Quantity, "Invalid quantity"); parseErr != nil {
			return expenseMessage(c, fiber.StatusBadRequest, parseErr.Error())
		}
	}

	tx, repoErr := h.repo.BeginTx(c.Context())
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Internal server error")
	}
	defer tx.Rollback(c.Context())

	existing, repoErr := h.repo.GetByIDWithTx(c.Context(), tx, id)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to get expense")
	}
	if existing == nil {
		return expenseMessage(c, fiber.StatusNotFound, "Expense not found")
	}
	if ok, staleErr := requireExpenseExpectedUpdatedAt(c, req.ExpectedUpdatedAt, existing.UpdatedAt); !ok {
		return staleErr
	}

	oldValues := expenseAuditValues(existing)

	finalProductID := existing.ProductID
	if req.ProductID != nil {
		if *req.ProductID == "" {
			finalProductID = nil
		} else {
			finalProductID, _ = parseOptionalExpenseUUIDField(req.ProductID, "Invalid product ID")
		}
	}

	finalQuantity := existing.Quantity
	if req.Quantity != nil {
		if *req.Quantity == "" {
			finalQuantity = nil
		} else {
			finalQuantity, _ = parseOptionalExpensePositiveDecimalField(req.Quantity, "Invalid quantity")
		}
	}

	finalCategoryID := existing.CategoryID
	if req.CategoryID != nil {
		if *req.CategoryID == "" {
			finalCategoryID = nil
		} else {
			finalCategoryID, _ = parseOptionalExpenseUUIDField(req.CategoryID, "Invalid category ID")
		}
	}

	categoryUnchanged := sameUUIDPointer(existing.CategoryID, finalCategoryID)
	finalIsInventoryPurchase, err := h.isInventoryPurchaseCategoryWithArchived(c.Context(), finalCategoryID, categoryUnchanged)
	if err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}
	if !finalIsInventoryPurchase {
		finalProductID = nil
		finalQuantity = nil
	}
	if finalIsInventoryPurchase && finalProductID == nil {
		return expenseMessage(c, fiber.StatusBadRequest, "Product is required for Inventory Purchase")
	}
	if finalIsInventoryPurchase && finalQuantity == nil {
		return expenseMessage(c, fiber.StatusBadRequest, "Quantity is required for Inventory Purchase")
	}

	finalAppliesInventory := existing.AppliesInventory
	if req.AppliesInventory != nil {
		finalAppliesInventory = *req.AppliesInventory
	}
	if !finalIsInventoryPurchase {
		finalAppliesInventory = false
	}
	if finalIsInventoryPurchase && finalProductID != nil {
		if ok, staleErr := h.requireExpenseProductCurrent(c, *finalProductID, req.ExpectedProductUpdatedAt); !ok {
			return staleErr
		}
	}

	if err := h.syncExpenseInventory(c.Context(), tx, id, userID, existing, finalAppliesInventory, finalProductID, finalQuantity); err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}

	existing.CategoryID = finalCategoryID
	existing.ProductID = finalProductID
	existing.Quantity = finalQuantity
	existing.AppliesInventory = finalAppliesInventory

	if req.Amount != nil {
		amount, parseErr := parseRequiredAmount(*req.Amount)
		if parseErr != nil {
			return expenseMessage(c, fiber.StatusBadRequest, parseErr.Error())
		}
		existing.Amount = amount
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.ExpenseDate != nil {
		date, parseErr := parseExpenseDateField(*req.ExpenseDate, "Invalid expense date format")
		if parseErr != nil {
			return expenseMessage(c, fiber.StatusBadRequest, parseErr.Error())
		}
		existing.ExpenseDate = date
	}
	if req.Vendor != nil {
		existing.Vendor = req.Vendor
	}
	if req.ReferenceNumber != nil {
		existing.ReferenceNumber = req.ReferenceNumber
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	updated, repoErr := h.repo.UpdateWithTx(c.Context(), tx, existing)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to update expense")
	}
	if err := tx.Commit(c.Context()); err != nil {
		return expenseInternalError(c, err, "Failed to update expense")
	}

	newValues := expenseAuditValues(updated)
	audit.LogWithValues(c, models.AuditActionExpenseUpdate, models.AuditEntityExpense, id.String(), "Updated expense: "+updated.Description, oldValues, newValues)

	return c.JSON(fiber.Map{"data": expenseToResponse(updated)})
}

// Delete handles DELETE /api/v1/expenses/:id
func (h *ExpenseHandler) Delete(c *fiber.Ctx) error {
	id, err := parseExpenseParamID(c, "id", "Invalid expense ID")
	if err != nil {
		return err
	}

	userID := middleware.GetUserID(c)

	tx, repoErr := h.repo.BeginTx(c.Context())
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Internal server error")
	}
	defer tx.Rollback(c.Context())

	expense, repoErr := h.repo.GetByIDWithTx(c.Context(), tx, id)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to retrieve expense")
	}
	if expense == nil {
		return expenseMessage(c, fiber.StatusNotFound, "Expense not found")
	}
	if ok, staleErr := requireExpenseExpectedUpdatedAt(c, expectedUpdatedAtFromQuery(c), expense.UpdatedAt); !ok {
		return staleErr
	}

	if expense.AppliesInventory && expense.ProductID != nil && expense.Quantity != nil {
		if err := h.ensureExpenseInventoryProductActive(c.Context(), expense.ProductID); err != nil {
			return expenseMessage(c, fiber.StatusConflict, err.Error())
		}
		if _, err := h.inventoryRepo.AdjustStockWithTx(
			c.Context(),
			tx,
			*expense.ProductID,
			models.AdjustmentPurchase,
			expense.Quantity.Neg(),
			expenseInventoryReason("delete revert", expense.ID),
			stringPtr("expense"),
			&expense.ID,
			userID,
		); err != nil {
			return expenseMessage(c, fiber.StatusBadRequest, "Failed to revert inventory: "+err.Error())
		}
	}

	if repoErr := h.repo.DeleteWithTx(c.Context(), tx, id); repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to delete expense")
	}
	if err := tx.Commit(c.Context()); err != nil {
		return expenseInternalError(c, err, "Failed to delete expense")
	}

	audit.LogFromFiber(c, models.AuditActionExpenseDelete, models.AuditEntityExpense, id.String(), "Deleted expense")
	return c.JSON(fiber.Map{"message": "Expense deleted successfully"})
}

// GetSummary handles GET /api/v1/expenses/summary
func (h *ExpenseHandler) GetSummary(c *fiber.Ctx) error {
	now := time.Now().In(reportBusinessLocation)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, reportBusinessLocation)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	if startStr := c.Query("start_date"); startStr != "" {
		parsed, err := parseExpenseDateField(startStr, "Invalid start date format (use YYYY-MM-DD)")
		if err != nil {
			return expenseMessage(c, fiber.StatusBadRequest, err.Error())
		}
		startOfMonth = parsed
	}
	if endStr := c.Query("end_date"); endStr != "" {
		parsed, err := parseExpenseDateField(endStr, "Invalid end date format (use YYYY-MM-DD)")
		if err != nil {
			return expenseMessage(c, fiber.StatusBadRequest, err.Error())
		}
		endOfMonth = parsed
	}

	summary, repoErr := h.repo.GetSummary(c.Context(), startOfMonth, endOfMonth)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to get expense summary")
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"total_amount":  summary.TotalAmount.String(),
			"expense_count": summary.ExpenseCount,
			"by_category":   summary.ByCategory,
			"start_date":    startOfMonth.Format(reportDateLayout),
			"end_date":      endOfMonth.Format(reportDateLayout),
		},
	})
}

// GetMonthlyTotals handles GET /api/v1/expenses/monthly
func (h *ExpenseHandler) GetMonthlyTotals(c *fiber.Ctx) error {
	months := c.QueryInt("months", 6)
	if months < 1 || months > 24 {
		months = 6
	}

	totals, err := h.repo.GetMonthlyTotals(c.Context(), months)
	if err != nil {
		return expenseInternalError(c, err, "Failed to get monthly expense totals")
	}

	return c.JSON(fiber.Map{"data": totals})
}

func stringPointerFromQuery(c *fiber.Ctx, key string) *string {
	value := c.Query(key)
	if value == "" {
		return nil
	}
	return &value
}
