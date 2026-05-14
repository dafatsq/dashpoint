package handlers

import (
	"github.com/gofiber/fiber/v2"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/models"
)

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	if apiErr := validateCategoryName(req.Name); apiErr != nil {
		return respondCategoryError(c, apiErr)
	}

	parentID, apiErr := parseOptionalParentID(req.ParentID)
	if apiErr != nil {
		return respondCategoryError(c, apiErr)
	}

	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
		IsActive:    true,
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	if err := h.categoryRepo.Create(c.UserContext(), category); err != nil {
		return respondCategoryInternalError(c, "Failed to create category")
	}

	oldValues := map[string]interface{}{}
	newValues := map[string]interface{}{
		"name":        category.Name,
		"description": category.Description,
		"parent_id":   category.ParentID,
		"sort_order":  category.SortOrder,
		"is_active":   category.IsActive,
	}
	audit.LogWithValues(c, models.AuditActionCategoryCreate, models.AuditEntityCategory, category.ID.String(), "Created category", oldValues, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Category created successfully",
		"category": categoryResponse(category, nil),
	})
}

func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	id, apiErr := parseCategoryIDParam(c)
	if apiErr != nil {
		return respondCategoryError(c, apiErr)
	}

	category, err := h.categoryRepo.GetByID(c.UserContext(), id)
	if err != nil {
		if isCategoryNotFound(err) {
			return categoryNotFoundResponse(c)
		}
		return respondCategoryInternalError(c, "Failed to fetch category")
	}

	oldValues := map[string]interface{}{
		"name":        category.Name,
		"description": category.Description,
		"parent_id":   category.ParentID,
		"sort_order":  category.SortOrder,
		"is_active":   category.IsActive,
	}

	var req UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	if req.Name != nil {
		if apiErr := validateCategoryName(*req.Name); apiErr != nil {
			return respondCategoryError(c, apiErr)
		}
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = req.Description
	}
	if req.ParentID != nil {
		parentID, parentErr := parseOptionalParentID(req.ParentID)
		if parentErr != nil {
			return respondCategoryError(c, parentErr)
		}
		category.ParentID = parentID
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := h.categoryRepo.Update(c.UserContext(), category); err != nil {
		if isCategoryNotFound(err) {
			return categoryNotFoundResponse(c)
		}
		return respondCategoryInternalError(c, "Failed to update category")
	}

	newValues := map[string]interface{}{
		"name":        category.Name,
		"description": category.Description,
		"parent_id":   category.ParentID,
		"sort_order":  category.SortOrder,
		"is_active":   category.IsActive,
	}
	audit.LogWithValues(c, models.AuditActionCategoryUpdate, models.AuditEntityCategory, category.ID.String(), "Updated category", oldValues, newValues)

	count, err := h.categoryRepo.GetProductCount(c.UserContext(), id)
	if err != nil {
		return respondCategoryInternalError(c, "Failed to update category")
	}

	return c.JSON(fiber.Map{
		"message":  "Category updated successfully",
		"category": categoryResponse(category, &count),
	})
}

func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	id, apiErr := parseCategoryIDParam(c)
	if apiErr != nil {
		return respondCategoryError(c, apiErr)
	}

	category, err := h.categoryRepo.GetByID(c.UserContext(), id)
	if err != nil {
		if isCategoryNotFound(err) {
			return categoryNotFoundResponse(c)
		}
		return respondCategoryInternalError(c, "Failed to fetch category")
	}

	if err := h.categoryRepo.Delete(c.UserContext(), id); err != nil {
		if isCategoryNotFound(err) {
			return categoryNotFoundResponse(c)
		}
		return respondCategoryInternalError(c, "Failed to delete category")
	}

	audit.LogWithValues(
		c,
		models.AuditActionCategoryArchive,
		models.AuditEntityCategory,
		id.String(),
		"Archived category",
		map[string]interface{}{
			"name":        category.Name,
			"description": category.Description,
			"parent_id":   category.ParentID,
			"sort_order":  category.SortOrder,
			"is_active":   category.IsActive,
		},
		map[string]interface{}{"deleted": true},
	)

	return c.JSON(fiber.Map{"message": "Category deleted successfully"})
}

func (h *CategoryHandler) PermanentDelete(c *fiber.Ctx) error {
	id, apiErr := parseCategoryIDParam(c)
	if apiErr != nil {
		return respondCategoryError(c, apiErr)
	}

	productCount, err := h.categoryRepo.GetProductCount(c.UserContext(), id)
	if err != nil {
		if isCategoryNotFound(err) {
			return categoryNotFoundResponse(c)
		}
		return respondCategoryInternalError(c, "Failed to check category usage")
	}
	if productCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "HAS_PRODUCTS",
				"message": "Cannot delete category with associated products",
			},
		})
	}

	category, err := h.categoryRepo.GetByID(c.UserContext(), id)
	if err != nil {
		if isCategoryNotFound(err) {
			return categoryNotFoundResponse(c)
		}
		return respondCategoryInternalError(c, "Failed to fetch category")
	}

	if err := h.categoryRepo.PermanentDelete(c.UserContext(), id); err != nil {
		if isCategoryNotFound(err) {
			return categoryNotFoundResponse(c)
		}
		return respondCategoryInternalError(c, "Failed to permanently delete category")
	}

	audit.LogWithValues(
		c,
		models.AuditActionCategoryDelete,
		models.AuditEntityCategory,
		id.String(),
		"Permanently deleted category",
		map[string]interface{}{
			"name":        category.Name,
			"description": category.Description,
			"parent_id":   category.ParentID,
			"sort_order":  category.SortOrder,
			"is_active":   category.IsActive,
		},
		map[string]interface{}{"deleted_permanently": true},
	)

	return c.JSON(fiber.Map{"message": "Category permanently deleted successfully"})
}
