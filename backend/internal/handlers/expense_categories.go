package handlers

import (
	"github.com/gofiber/fiber/v2"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

const inventoryPurchaseCategorySystemKey = "inventory_purchase"

func isInventoryPurchaseExpenseCategory(category *models.ExpenseCategory) bool {
	return category != nil && category.SystemKey != nil && *category.SystemKey == inventoryPurchaseCategorySystemKey
}

func expenseCategoryAuditValues(category *models.ExpenseCategory) map[string]interface{} {
	if category == nil {
		return nil
	}

	values := map[string]interface{}{
		"name":      category.Name,
		"is_active": category.IsActive,
	}
	if category.Description != nil {
		values["description"] = *category.Description
	}
	return values
}

func expenseCategoryName(category *models.ExpenseCategory) string {
	if category == nil {
		return "Unknown"
	}
	return category.Name
}

// ListCategories handles GET /api/v1/expenses/categories
func (h *ExpenseHandler) ListCategories(c *fiber.Ctx) error {
	status, parseErr := parseExpenseCategoryStatus(c)
	if parseErr != nil {
		return expenseMessage(c, fiber.StatusBadRequest, parseErr.Error())
	}

	categories, err := h.repo.ListCategories(c.Context(), status)
	if err != nil {
		return expenseInternalError(c, err, "Failed to list expense categories")
	}

	return c.JSON(fiber.Map{"data": categories})
}

// CreateCategory handles POST /api/v1/expenses/categories
func (h *ExpenseHandler) CreateCategory(c *fiber.Ctx) error {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := parseExpenseBody(c, &req); err != nil {
		return err
	}
	if err := validateExpenseCategoryName(&req.Name); err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}
	if err := validateExpenseCategoryDescription(&req.Description); err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}

	exists, repoErr := h.repo.CategoryNameExists(c.Context(), req.Name, nil)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to validate category name")
	}
	if exists {
		return middleware.JSONError(c, fiber.StatusConflict, "NAME_EXISTS", "Expense category with this name already exists")
	}

	category, err := h.repo.CreateCategory(c.Context(), req.Name, req.Description)
	if err != nil {
		return expenseInternalError(c, err, "Failed to create expense category")
	}

	newValues := map[string]interface{}{
		"affected_category": category.Name,
		"name":              category.Name,
	}
	if category.Description != nil {
		newValues["description"] = *category.Description
	}
	audit.LogWithValues(c, models.AuditActionCategoryCreate, models.AuditEntityCategory, category.ID.String(), "Created expense category: "+category.Name, nil, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": category})
}

// GetCategory handles GET /api/v1/expenses/categories/:id
func (h *ExpenseHandler) GetCategory(c *fiber.Ctx) error {
	id, err := parseExpenseParamID(c, "id", "Invalid category ID")
	if err != nil {
		return err
	}

	category, repoErr := h.repo.GetCategoryByID(c.Context(), id)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to get expense category")
	}
	if category == nil {
		return expenseMessage(c, fiber.StatusNotFound, "Expense category not found")
	}

	return c.JSON(fiber.Map{"data": category})
}

// UpdateCategory handles PATCH /api/v1/expenses/categories/:id
func (h *ExpenseHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := parseExpenseParamID(c, "id", "Invalid category ID")
	if err != nil {
		return err
	}

	var req struct {
		Name              *string `json:"name"`
		Description       *string `json:"description"`
		IsActive          *bool   `json:"is_active"`
		ExpectedUpdatedAt *string `json:"expected_updated_at"`
	}
	if err := parseExpenseBody(c, &req); err != nil {
		return err
	}
	if req.Name != nil {
		if err := validateExpenseCategoryName(req.Name); err != nil {
			return expenseMessage(c, fiber.StatusBadRequest, err.Error())
		}
	}
	if err := validateExpenseCategoryDescription(&req.Description); err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, err.Error())
	}

	category, repoErr := h.repo.GetCategoryByID(c.Context(), id)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to get expense category")
	}
	if category == nil {
		return expenseMessage(c, fiber.StatusNotFound, "Expense category not found")
	}
	if ok, staleErr := requireExpenseExpectedUpdatedAt(c, req.ExpectedUpdatedAt, category.UpdatedAt); !ok {
		return staleErr
	}
	if !category.IsActive && (req.IsActive == nil || !*req.IsActive) {
		return middleware.JSONError(c, fiber.StatusConflict, "CATEGORY_INACTIVE", "Archived expense categories cannot be changed")
	}

	oldValues := map[string]interface{}{
		"affected_category": category.Name,
		"name":              category.Name,
		"is_active":         category.IsActive,
	}
	if category.Description != nil {
		oldValues["description"] = *category.Description
	}

	if req.Name != nil {
		exists, err := h.repo.CategoryNameExists(c.Context(), *req.Name, &id)
		if err != nil {
			return expenseInternalError(c, err, "Failed to validate category name")
		}
		if exists {
			return middleware.JSONError(c, fiber.StatusConflict, "NAME_EXISTS", "Expense category with this name already exists")
		}
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = req.Description
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	updated, repoErr := h.repo.UpdateCategory(c.Context(), category)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to update expense category")
	}

	newValues := map[string]interface{}{
		"affected_category": updated.Name,
		"name":              updated.Name,
		"is_active":         updated.IsActive,
	}
	if updated.Description != nil {
		newValues["description"] = *updated.Description
	}
	audit.LogWithValues(c, models.AuditActionCategoryUpdate, models.AuditEntityCategory, updated.ID.String(), "Updated expense category: "+updated.Name, oldValues, newValues)

	return c.JSON(fiber.Map{"data": updated})
}

// DeleteCategory handles DELETE /api/v1/expenses/categories/:id
func (h *ExpenseHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := parseExpenseParamID(c, "id", "Invalid category ID")
	if err != nil {
		return err
	}

	category, repoErr := h.repo.GetCategoryByID(c.Context(), id)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to get expense category")
	}
	if category == nil {
		return expenseMessage(c, fiber.StatusNotFound, "Expense category not found")
	}
	categoryName := expenseCategoryName(category)
	if ok, staleErr := requireExpenseExpectedUpdatedAt(c, expectedUpdatedAtFromQuery(c), category.UpdatedAt); !ok {
		return staleErr
	}
	if !category.IsActive {
		return middleware.JSONError(c, fiber.StatusConflict, "CATEGORY_INACTIVE", "Expense category is already archived")
	}

	if repoErr := h.repo.DeleteCategory(c.Context(), id); repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to delete expense category")
	}

	oldValues := expenseCategoryAuditValues(category)
	if oldValues == nil {
		oldValues = map[string]interface{}{"name": categoryName}
	}
	newValues := map[string]interface{}{"deleted": true}

	audit.LogWithValues(c, models.AuditActionCategoryArchive, models.AuditEntityCategory, id.String(), "Archived expense category: "+categoryName, oldValues, newValues)

	return c.JSON(fiber.Map{"message": "Expense category archived successfully"})
}

// PermanentDeleteCategory handles DELETE /api/v1/expenses/categories/:id/permanent
func (h *ExpenseHandler) PermanentDeleteCategory(c *fiber.Ctx) error {
	id, err := parseExpenseParamID(c, "id", "Invalid category ID")
	if err != nil {
		return err
	}

	category, repoErr := h.repo.GetCategoryByID(c.Context(), id)
	if repoErr != nil {
		return expenseInternalError(c, repoErr, "Failed to get expense category")
	}
	if category == nil {
		return expenseMessage(c, fiber.StatusNotFound, "Expense category not found")
	}
	categoryName := expenseCategoryName(category)
	if ok, staleErr := requireExpenseExpectedUpdatedAt(c, expectedUpdatedAtFromQuery(c), category.UpdatedAt); !ok {
		return staleErr
	}

	// The inventory-purchase system category cannot be deleted.
	if isInventoryPurchaseExpenseCategory(category) {
		return expenseMessage(c, fiber.StatusForbidden, "The 'Inventory Purchase' category is a system category and cannot be deleted")
	}
	if category.IsActive {
		return middleware.JSONError(c, fiber.StatusConflict, "CATEGORY_ACTIVE", "Active expense categories cannot be permanently deleted")
	}

	if repoErr := h.repo.PermanentDeleteCategory(c.Context(), id); repoErr != nil {
		return expenseMessage(c, fiber.StatusBadRequest, repoErr.Error())
	}

	oldValues := expenseCategoryAuditValues(category)
	if oldValues == nil {
		oldValues = map[string]interface{}{"name": categoryName}
	}
	newValues := map[string]interface{}{"deleted_permanently": true}

	audit.LogWithValues(c, models.AuditActionCategoryDelete, models.AuditEntityCategory, id.String(), "Permanently deleted expense category: "+categoryName, oldValues, newValues)

	return c.JSON(fiber.Map{"message": "Expense category permanently deleted successfully"})
}
