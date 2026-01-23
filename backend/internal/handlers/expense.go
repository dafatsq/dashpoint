package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

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
	categories, err := h.repo.ListCategories(c.Context(), true)
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

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": category,
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
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

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
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

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

	// Determine final state for inventory logic
	finalProductID := existing.ProductID
	if req.ProductID != nil {
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
	if req.Quantity != nil {
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
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
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
