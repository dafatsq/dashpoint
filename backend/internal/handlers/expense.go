package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// ExpenseHandler handles expense-related HTTP requests
type ExpenseHandler struct {
	repo          *repository.ExpenseRepository
	inventoryRepo *repository.InventoryRepository
	productRepo   *repository.ProductRepository
}

// NewExpenseHandler creates a new expense handler
func NewExpenseHandler(repo *repository.ExpenseRepository, inventoryRepo *repository.InventoryRepository, productRepo *repository.ProductRepository) *ExpenseHandler {
	return &ExpenseHandler{
		repo:          repo,
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
	}
}

// --- Request/Response types ---

type CreateExpenseRequest struct {
	CategoryID      *string `json:"category_id"`
	ProductID       *string `json:"product_id"`
	Quantity        *string `json:"quantity"`
	Amount          string  `json:"amount"`
	Description     string  `json:"description"`
	ExpenseDate     string  `json:"expense_date"`
	Vendor          *string `json:"vendor"`
	ReferenceNumber *string `json:"reference_number"`
	Notes           *string `json:"notes"`
}

type UpdateExpenseRequest struct {
	CategoryID      *string `json:"category_id"`
	ProductID       *string `json:"product_id"`
	Quantity        *string `json:"quantity"`
	Amount          *string `json:"amount"`
	Description     *string `json:"description"`
	ExpenseDate     *string `json:"expense_date"`
	Vendor          *string `json:"vendor"`
	ReferenceNumber *string `json:"reference_number"`
	Notes           *string `json:"notes"`
}

type ExpenseResponse struct {
	ID              string  `json:"id"`
	CategoryID      *string `json:"category_id,omitempty"`
	CategoryName    *string `json:"category_name,omitempty"`
	ProductID       *string `json:"product_id,omitempty"`
	ProductName     *string `json:"product_name,omitempty"`
	Quantity        *string `json:"quantity,omitempty"`
	Amount          string  `json:"amount"`
	Description     string  `json:"description"`
	ExpenseDate     string  `json:"expense_date"`
	Vendor          *string `json:"vendor,omitempty"`
	ReferenceNumber *string `json:"reference_number,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	CreatedBy       string  `json:"created_by"`
	CreatedByName   *string `json:"created_by_name,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func expenseToResponse(e *models.Expense) ExpenseResponse {
	resp := ExpenseResponse{
		ID:              e.ID.String(),
		Amount:          e.Amount.String(),
		Description:     e.Description,
		ExpenseDate:     e.ExpenseDate.Format("2006-01-02"),
		Vendor:          e.Vendor,
		ReferenceNumber: e.ReferenceNumber,
		Notes:           e.Notes,
		CreatedBy:       e.CreatedBy.String(),
		CreatedByName:   e.CreatedByName,
		CreatedAt:       e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       e.UpdatedAt.Format(time.RFC3339),
	}
	if e.CategoryID != nil {
		catID := e.CategoryID.String()
		resp.CategoryID = &catID
	}
	resp.CategoryName = e.CategoryName
	if e.ProductID != nil {
		prodID := e.ProductID.String()
		resp.ProductID = &prodID
	}
	resp.ProductName = e.ProductName
	if e.Quantity != nil {
		qty := e.Quantity.String()
		resp.Quantity = &qty
	}
	return resp
}

// --- Category Handlers ---

// ListCategories handles GET /api/v1/expenses/categories
func (h *ExpenseHandler) ListCategories(c *fiber.Ctx) error {
	status := c.Query("status", "")
	if status == "" {
		if c.Query("active_only", "true") == "true" {
			status = "active"
		} else {
			status = "all"
		}
	}

	categories, err := h.repo.ListCategories(c.Context(), status)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list expense categories")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to list expense categories",
		})
	}

	return c.JSON(fiber.Map{
		"data": categories,
	})
}

// CreateCategory handles POST /api/v1/expenses/categories
func (h *ExpenseHandler) CreateCategory(c *fiber.Ctx) error {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Category name is required",
		})
	}

	category, err := h.repo.CreateCategory(c.Context(), req.Name, req.Description)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create expense category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create expense category",
		})
	}

	// Audit log
	newValues := map[string]interface{}{
		"affected_category": category.Name,
		"name":              category.Name,
	}
	if category.Description != nil {
		newValues["description"] = *category.Description
	}
	audit.LogWithValues(c, models.AuditActionCategoryCreate, models.AuditEntityCategory, category.ID.String(), "Created expense category: "+category.Name, nil, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": category,
	})
}

// GetCategory handles GET /api/v1/expenses/categories/:id
func (h *ExpenseHandler) GetCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid category ID",
		})
	}

	category, err := h.repo.GetCategoryByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get expense category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get expense category",
		})
	}

	if category == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Expense category not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": category,
	})
}

// UpdateCategory handles PATCH /api/v1/expenses/categories/:id
func (h *ExpenseHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid category ID",
		})
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		IsActive    *bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	category, err := h.repo.GetCategoryByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get expense category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get expense category",
		})
	}

	if category == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Expense category not found",
		})
	}

	// Capture old values for audit
	oldValues := map[string]interface{}{
		"affected_category": category.Name,
		"name":              category.Name,
		"is_active":         category.IsActive,
	}
	if category.Description != nil {
		oldValues["description"] = *category.Description
	}

	// Update fields
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = req.Description
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	updated, err := h.repo.UpdateCategory(c.Context(), category)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update expense category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update expense category",
		})
	}

	// Audit log
	newValues := map[string]interface{}{
		"affected_category": updated.Name,
		"name":              updated.Name,
		"is_active":         updated.IsActive,
	}
	if updated.Description != nil {
		newValues["description"] = *updated.Description
	}
	audit.LogWithValues(c, models.AuditActionCategoryUpdate, models.AuditEntityCategory, updated.ID.String(), "Updated expense category: "+updated.Name, oldValues, newValues)

	return c.JSON(fiber.Map{
		"data": updated,
	})
}

// DeleteCategory handles DELETE /api/v1/expenses/categories/:id
func (h *ExpenseHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid category ID",
		})
	}

	// Fetch category for audit
	category, _ := h.repo.GetCategoryByID(c.Context(), id)
	categoryName := "Unknown"
	if category != nil {
		categoryName = category.Name
	}

	err = h.repo.DeleteCategory(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete expense category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete expense category",
		})
	}

	// Audit log
	audit.LogWithValues(c, models.AuditActionCategoryArchive, models.AuditEntityCategory, id.String(), "Archived expense category: "+categoryName, map[string]interface{}{
		"affected_category": categoryName,
	}, nil)

	return c.JSON(fiber.Map{
		"message": "Expense category archived successfully",
	})
}

// PermanentDeleteCategory handles DELETE /api/v1/expenses/categories/:id/permanent
func (h *ExpenseHandler) PermanentDeleteCategory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid category ID",
		})
	}

	// Fetch category for audit
	category, _ := h.repo.GetCategoryByID(c.Context(), id)
	categoryName := "Unknown"
	if category != nil {
		categoryName = category.Name
	}

	err = h.repo.PermanentDeleteCategory(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to permanently delete expense category")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Audit log
	audit.LogWithValues(c, models.AuditActionCategoryDelete, models.AuditEntityCategory, id.String(), "Permanently deleted expense category: "+categoryName, map[string]interface{}{
		"affected_category": categoryName,
	}, nil)

	return c.JSON(fiber.Map{
		"message": "Expense category permanently deleted successfully",
	})
}

// --- Expense Handlers ---

// List handles GET /api/v1/expenses
func (h *ExpenseHandler) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	var categoryID *uuid.UUID
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		id, err := uuid.Parse(catIDStr)
		if err == nil {
			categoryID = &id
		}
	}

	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		t, err := time.Parse("2006-01-02", startStr)
		if err == nil {
			startDate = &t
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err == nil {
			endDate = &t
		}
	}

	expenses, total, err := h.repo.List(c.Context(), categoryID, startDate, endDate, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list expenses")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to list expenses",
		})
	}

	var resp []ExpenseResponse
	for _, e := range expenses {
		resp = append(resp, expenseToResponse(&e))
	}

	return c.JSON(fiber.Map{
		"data":   resp,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Create handles POST /api/v1/expenses
func (h *ExpenseHandler) Create(c *fiber.Ctx) error {
	var req CreateExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Amount == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Amount is required",
		})
	}
	if req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Description is required",
		})
	}
	if req.ExpenseDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Expense date is required",
		})
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid amount",
		})
	}

	// Parse date
	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid expense date format (use YYYY-MM-DD)",
		})
	}

	// Get user ID from context
	userID := middleware.GetUserID(c)

	expense := &models.Expense{
		Amount:          amount,
		Description:     req.Description,
		ExpenseDate:     expenseDate,
		Vendor:          req.Vendor,
		ReferenceNumber: req.ReferenceNumber,
		Notes:           req.Notes,
		CreatedBy:       userID,
	}

	// Parse category ID if provided
	if req.CategoryID != nil && *req.CategoryID != "" {
		catID, err := uuid.Parse(*req.CategoryID)
		if err == nil {
			expense.CategoryID = &catID
		}
	}

	// Parse product ID and quantity if provided (for inventory purchases)
	if req.ProductID != nil && *req.ProductID != "" {
		prodID, err := uuid.Parse(*req.ProductID)
		if err == nil {
			expense.ProductID = &prodID
		}
	}

	if req.Quantity != nil && *req.Quantity != "" {
		qty, err := decimal.NewFromString(*req.Quantity)
		if err == nil && qty.GreaterThan(decimal.Zero) {
			expense.Quantity = &qty
		}
	}

	// If this is an inventory purchase with product and quantity, use transaction
	var created *models.Expense
	if expense.ProductID != nil && expense.Quantity != nil {
		// Start transaction
		tx, err := h.repo.BeginTx(c.Context())
		if err != nil {
			log.Error().Err(err).Msg("Failed to begin transaction")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create expense",
			})
		}
		defer tx.Rollback(c.Context())

		// Create expense
		created, err = h.repo.CreateWithTx(c.Context(), tx, expense)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create expense")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create expense",
				"error":   err.Error(),
			})
		}

		// Adjust inventory
		log.Info().Msgf("Adjusting inventory for product %s, quantity %s", created.ProductID.String(), created.Quantity.String())
		_, err = h.inventoryRepo.AdjustStockWithTx(
			c.Context(),
			tx,
			*created.ProductID,
			models.AdjustmentPurchase,
			*created.Quantity,
			stringPtr("Inventory purchase - Expense ID: "+created.ID.String()),
			stringPtr("expense"),
			&created.ID,
			userID,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to adjust inventory for purchase")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to adjust inventory: " + err.Error(),
			})
		}

		// Commit transaction
		if err := tx.Commit(c.Context()); err != nil {
			log.Error().Err(err).Msg("Failed to commit transaction")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create expense",
			})
		}
	} else {
		// No inventory adjustment needed, create expense directly
		created, err = h.repo.Create(c.Context(), expense)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create expense")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create expense",
				"error":   err.Error(),
			})
		}
	}

	log.Info().Msgf("Created expense: %+v", created)

	log.Info().Msg("Converting expense to response")
	response := expenseToResponse(created)
	log.Info().Msgf("Response: %+v", response)

	// Audit log with new values
	newValues := map[string]interface{}{
		"amount":       created.Amount.String(),
		"description":  created.Description,
		"expense_date": created.ExpenseDate.Format("2006-01-02"),
	}
	if created.CategoryID != nil {
		newValues["category_id"] = created.CategoryID.String()
		// Also find and log category name
		categories, err := h.repo.ListCategories(c.Context(), "all")
		if err == nil {
			for _, cat := range categories {
				if cat.ID == *created.CategoryID {
					newValues["category"] = cat.Name
					break
				}
			}
		}
	}
	if created.ProductID != nil {
		newValues["product_id"] = created.ProductID.String()
	}
	if created.Vendor != nil {
		newValues["vendor"] = *created.Vendor
	}
	if created.Notes != nil {
		newValues["notes"] = *created.Notes
	}
	audit.LogWithValues(c, models.AuditActionExpenseCreate, models.AuditEntityExpense, created.ID.String(), "Created expense: "+created.Description, nil, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": response,
	})
}

func stringPtr(s string) *string {
	return &s
}

// Get handles GET /api/v1/expenses/:id
func (h *ExpenseHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid expense ID",
		})
	}

	expense, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get expense")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get expense",
		})
	}

	if expense == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Expense not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": expenseToResponse(expense),
	})
}

// Update handles PATCH /api/v1/expenses/:id
func (h *ExpenseHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid expense ID",
		})
	}

	// Get user ID from context
	userID := middleware.GetUserID(c)

	var req UpdateExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Start transaction
	tx, err := h.repo.BeginTx(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}
	defer tx.Rollback(c.Context())

	// Get existing expense
	existing, err := h.repo.GetByIDWithTx(c.Context(), tx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get expense")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get expense",
		})
	}
	if existing == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Expense not found",
		})
	}

	// Capture old values for audit BEFORE any modifications
	oldValues := map[string]interface{}{
		"amount":       existing.Amount.String(),
		"description":  existing.Description,
		"expense_date": existing.ExpenseDate.Format("2006-01-02"),
	}
	if existing.Vendor != nil {
		oldValues["vendor"] = *existing.Vendor
	}
	if existing.ReferenceNumber != nil {
		oldValues["reference_number"] = *existing.ReferenceNumber
	}
	if existing.Notes != nil {
		oldValues["notes"] = *existing.Notes
	}
	if existing.CategoryID != nil {
		if existing.CategoryName != nil {
			oldValues["category"] = *existing.CategoryName
		} else {
			oldValues["category"] = existing.CategoryID.String()
		}
	}
	if existing.ProductID != nil {
		if existing.ProductName != nil {
			oldValues["product"] = *existing.ProductName
		} else {
			oldValues["product"] = existing.ProductID.String()
		}
	}
	if existing.Quantity != nil {
		oldValues["quantity"] = existing.Quantity.String()
	}

	// Check if category is changing to a non-Inventory Purchase category
	// If so, we need to clear product and quantity
	isChangingToNonInventoryCategory := false
	if req.CategoryID != nil && *req.CategoryID != "" {
		newCatID, err := uuid.Parse(*req.CategoryID)
		if err == nil {
			// Get the new category name to check if it's Inventory Purchase
			categories, _ := h.repo.ListCategories(c.Context(), "all")
			for _, cat := range categories {
				if cat.ID == newCatID && cat.Name != "Inventory Purchase" {
					isChangingToNonInventoryCategory = true
					break
				}
			}
		}
	}

	// Determine final state for inventory logic
	finalProductID := existing.ProductID
	if isChangingToNonInventoryCategory {
		// Clear product when switching to non-inventory category
		finalProductID = nil
	} else if req.ProductID != nil {
		if *req.ProductID == "" {
			finalProductID = nil
		} else {
			pid, err := uuid.Parse(*req.ProductID)
			if err == nil {
				finalProductID = &pid
			}
		}
	}

	finalQuantity := existing.Quantity
	if isChangingToNonInventoryCategory {
		// Clear quantity when switching to non-inventory category
		finalQuantity = nil
	} else if req.Quantity != nil {
		if *req.Quantity == "" {
			finalQuantity = nil
		} else {
			qty, err := decimal.NewFromString(*req.Quantity)
			if err == nil && qty.GreaterThan(decimal.Zero) {
				finalQuantity = &qty
			}
		}
	}

	// Calculate inventory adjustments
	// We only act if something changed regarding product or quantity
	inventoryChanged := false
	if existing.ProductID != nil && finalProductID == nil {
		inventoryChanged = true
	} else if existing.ProductID == nil && finalProductID != nil {
		inventoryChanged = true
	} else if existing.ProductID != nil && finalProductID != nil && *existing.ProductID != *finalProductID {
		inventoryChanged = true
	} else if finalProductID != nil {
		// Same product, check quantity
		if existing.Quantity != nil && finalQuantity != nil && !existing.Quantity.Equal(*finalQuantity) {
			inventoryChanged = true
		} else if (existing.Quantity == nil) != (finalQuantity == nil) {
			inventoryChanged = true
		}
	}

	if inventoryChanged {
		log.Info().Msgf("Inventory impact detected for expense %s", id)

		// Strategy:
		// 1. If same product, calculate delta
		// 2. If different products, revert old and apply new

		isSameProduct := existing.ProductID != nil && finalProductID != nil && *existing.ProductID == *finalProductID

		if isSameProduct {
			// Calculate delta
			oldQ := decimal.Zero
			if existing.Quantity != nil {
				oldQ = *existing.Quantity
			}
			newQ := decimal.Zero
			if finalQuantity != nil {
				newQ = *finalQuantity
			}

			delta := newQ.Sub(oldQ)
			if !delta.IsZero() {
				log.Info().Msgf("Adjusting inventory (delta) for product %s: %s", finalProductID.String(), delta.String())
				_, err = h.inventoryRepo.AdjustStockWithTx(
					c.Context(),
					tx,
					*finalProductID,
					models.AdjustmentPurchase,
					delta,
					stringPtr("Update purchase qty - Expense: "+id.String()),
					stringPtr("expense"),
					&id,
					userID,
				)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"message": "Failed to update inventory: " + err.Error(),
					})
				}
			}
		} else {
			// Different products or one is missing

			// 1. Revert old if it existed
			if existing.ProductID != nil && existing.Quantity != nil {
				revertQty := existing.Quantity.Neg()
				log.Info().Msgf("Reverting old inventory for product %s: %s", existing.ProductID.String(), revertQty.String())
				_, err = h.inventoryRepo.AdjustStockWithTx(
					c.Context(),
					tx,
					*existing.ProductID,
					models.AdjustmentPurchase,
					revertQty,
					stringPtr("Revert purchase (edit) - Expense: "+id.String()),
					stringPtr("expense"),
					&id,
					userID,
				)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"message": "Failed to revert old inventory: " + err.Error(),
					})
				}
			}

			// 2. Apply new if it exists
			if finalProductID != nil && finalQuantity != nil {
				log.Info().Msgf("Applying new inventory for product %s: %s", finalProductID.String(), finalQuantity.String())
				_, err = h.inventoryRepo.AdjustStockWithTx(
					c.Context(),
					tx,
					*finalProductID,
					models.AdjustmentPurchase,
					*finalQuantity,
					stringPtr("Apply purchase (edit) - Expense: "+id.String()),
					stringPtr("expense"),
					&id,
					userID,
				)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"message": "Failed to apply new inventory: " + err.Error(),
					})
				}
			}
		}
	}

	// Update fields
	if req.CategoryID != nil {
		if *req.CategoryID == "" {
			existing.CategoryID = nil
		} else {
			catID, err := uuid.Parse(*req.CategoryID)
			if err == nil {
				existing.CategoryID = &catID
			}
		}
	}

	existing.ProductID = finalProductID
	existing.Quantity = finalQuantity

	if req.Amount != nil {
		amount, err := decimal.NewFromString(*req.Amount)
		if err != nil || amount.LessThanOrEqual(decimal.Zero) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid amount",
			})
		}
		existing.Amount = amount
	}

	if req.Description != nil {
		existing.Description = *req.Description
	}

	if req.ExpenseDate != nil {
		date, err := time.Parse("2006-01-02", *req.ExpenseDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid expense date format",
			})
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

	updated, err := h.repo.UpdateWithTx(c.Context(), tx, existing)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update expense")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update expense",
		})
	}

	if err := tx.Commit(c.Context()); err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update expense",
		})
	}

	// Audit log with old and new values
	newValues := map[string]interface{}{
		"amount":       updated.Amount.String(),
		"description":  updated.Description,
		"expense_date": updated.ExpenseDate.Format("2006-01-02"),
	}
	if updated.Vendor != nil {
		newValues["vendor"] = *updated.Vendor
	}
	if updated.ReferenceNumber != nil {
		newValues["reference_number"] = *updated.ReferenceNumber
	}
	if updated.Notes != nil {
		newValues["notes"] = *updated.Notes
	}
	if updated.CategoryID != nil {
		if updated.CategoryName != nil {
			newValues["category"] = *updated.CategoryName
		} else {
			newValues["category"] = updated.CategoryID.String()
		}
	}
	if updated.ProductID != nil {
		if updated.ProductName != nil {
			newValues["product"] = *updated.ProductName
		} else {
			newValues["product"] = updated.ProductID.String()
		}
	} else if oldValues["product"] != nil {
		// Product was removed
		newValues["product"] = "[removed]"
	}
	if updated.Quantity != nil {
		newValues["quantity"] = updated.Quantity.String()
	} else if oldValues["quantity"] != nil {
		// Quantity was removed
		newValues["quantity"] = "[removed]"
	}
	audit.LogWithValues(c, models.AuditActionExpenseUpdate, models.AuditEntityExpense, id.String(), "Updated expense: "+updated.Description, oldValues, newValues)

	return c.JSON(fiber.Map{
		"data": expenseToResponse(updated),
	})
}

// Delete handles DELETE /api/v1/expenses/:id
func (h *ExpenseHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid expense ID",
		})
	}

	// Get user ID from context for the adjustment record
	userID := middleware.GetUserID(c)

	// Start transaction
	tx, err := h.repo.BeginTx(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}
	defer tx.Rollback(c.Context())

	// Get the expense first to check if we need to revert inventory
	expense, err := h.repo.GetByIDWithTx(c.Context(), tx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get expense for deletion")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve expense",
		})
	}
	if expense == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Expense not found",
		})
	}

	// If this was an inventory purchase, revert the stock
	if expense.ProductID != nil && expense.Quantity != nil {
		// Reverting means removing the quantity we added
		// So we pass -Quantity
		revertQty := expense.Quantity.Neg()

		log.Info().Msgf("Reverting inventory for expense deletion: product %s, quantity %s", expense.ProductID.String(), revertQty.String())

		_, err = h.inventoryRepo.AdjustStockWithTx(
			c.Context(),
			tx,
			*expense.ProductID,
			models.AdjustmentPurchase, // Using purchase type to reflect it's related to purchase
			revertQty,
			stringPtr("Revert purchase - Expense Deleted: "+expense.ID.String()),
			stringPtr("expense"),
			&expense.ID,
			userID,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to revert inventory for expense deletion")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to revert inventory: " + err.Error(),
			})
		}
	}

	err = h.repo.DeleteWithTx(c.Context(), tx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete expense")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete expense",
		})
	}

	if err := tx.Commit(c.Context()); err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete expense",
		})
	}

	// Audit log
	audit.LogFromFiber(c, models.AuditActionExpenseDelete, models.AuditEntityExpense, id.String(), "Deleted expense")

	return c.JSON(fiber.Map{
		"message": "Expense deleted successfully",
	})
}

// GetSummary handles GET /api/v1/expenses/summary
func (h *ExpenseHandler) GetSummary(c *fiber.Ctx) error {
	// Default to current month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	if startStr := c.Query("start_date"); startStr != "" {
		t, err := time.Parse("2006-01-02", startStr)
		if err == nil {
			startOfMonth = t
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err == nil {
			endOfMonth = t
		}
	}

	summary, err := h.repo.GetSummary(c.Context(), startOfMonth, endOfMonth)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get expense summary")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get expense summary",
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"total_amount":  summary.TotalAmount.String(),
			"expense_count": summary.ExpenseCount,
			"by_category":   summary.ByCategory,
			"start_date":    startOfMonth.Format("2006-01-02"),
			"end_date":      endOfMonth.Format("2006-01-02"),
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
		log.Error().Err(err).Msg("Failed to get monthly expense totals")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get monthly expense totals",
		})
	}

	return c.JSON(fiber.Map{
		"data": totals,
	})
}
