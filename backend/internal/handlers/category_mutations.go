package handlers

import (
	"github.com/gofiber/fiber/v2"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/models"
)

func parseCategoryBody[T any](c *fiber.Ctx, req *T) error {
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}
	return nil
}

func categoryNameExistsResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    "NAME_EXISTS",
			"message": "A category with this name already exists in this location",
		},
	})
}

func categoryAuditValues(category *models.Category) map[string]interface{} {
	return map[string]interface{}{
		"name":        category.Name,
		"description": category.Description,
		"parent_id":   category.ParentID,
		"sort_order":  category.SortOrder,
		"is_active":   category.IsActive,
	}
}

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := parseCategoryBody(c, &req); err != nil {
		return err
	}

	if apiErr := validateCategoryName(req.Name); apiErr != nil {
		return respondCategoryError(c, apiErr)
	}

	parentID, apiErr := parseOptionalParentID(req.ParentID)
	if apiErr != nil {
		return respondCategoryError(c, apiErr)
	}

	exists, checkErr := h.categoryRepo.DuplicateSiblingExists(c.UserContext(), req.Name, parentID, nil)
	if checkErr != nil {
		return respondCategoryInternalError(c, "Failed to check for duplicate category")
	}
	if exists {
		return categoryNameExistsResponse(c)
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

	audit.LogWithValues(c, models.AuditActionCategoryCreate, models.AuditEntityCategory, category.ID.String(), "Created category", map[string]interface{}{}, categoryAuditValues(category))

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

	oldValues := categoryAuditValues(category)

	var req UpdateCategoryRequest
	if err := parseCategoryBody(c, &req); err != nil {
		return err
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

	if req.Name != nil || req.ParentID != nil {
		exists, checkErr := h.categoryRepo.DuplicateSiblingExists(c.UserContext(), category.Name, category.ParentID, &category.ID)
		if checkErr != nil {
			return respondCategoryInternalError(c, "Failed to check for duplicate category")
		}
		if exists {
			return categoryNameExistsResponse(c)
		}
	}

	if err := h.categoryRepo.Update(c.UserContext(), category); err != nil {
		if isCategoryNotFound(err) {
			return categoryNotFoundResponse(c)
		}
		return respondCategoryInternalError(c, "Failed to update category")
	}

	audit.LogWithValues(c, models.AuditActionCategoryUpdate, models.AuditEntityCategory, category.ID.String(), "Updated category", oldValues, categoryAuditValues(category))

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
		categoryAuditValues(category),
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
		categoryAuditValues(category),
		map[string]interface{}{"deleted_permanently": true},
	)

	return c.JSON(fiber.Map{"message": "Category permanently deleted successfully"})
}
