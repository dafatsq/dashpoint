package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/repository"
)

func (h *ProductHandler) validateActiveProductCategory(c *fiber.Ctx, categoryID *uuid.UUID) bool {
	if categoryID == nil {
		return true
	}

	category, err := h.categoryRepo.GetByID(c.Context(), *categoryID)
	if err != nil {
		if isCategoryNotFound(err) {
			_ = productJSONError(c, fiber.StatusBadRequest, "INVALID_CATEGORY", "Category not found")
			return false
		}
		_ = productInternalError(c, err, "Failed to get category", "Failed to validate category")
		return false
	}
	if category == nil {
		_ = productJSONError(c, fiber.StatusBadRequest, "INVALID_CATEGORY", "Category not found")
		return false
	}
	if !category.IsActive {
		_ = productJSONError(c, fiber.StatusConflict, "CATEGORY_INACTIVE", "Archived categories cannot be used")
		return false
	}

	return true
}

func productCategoryWriteError(c *fiber.Ctx, err error) error {
	if errors.Is(err, repository.ErrProductCategoryNotFound) {
		return productJSONError(c, fiber.StatusBadRequest, "INVALID_CATEGORY", "Category not found")
	}
	if errors.Is(err, repository.ErrProductCategoryInactive) {
		return productJSONError(c, fiber.StatusConflict, "CATEGORY_INACTIVE", "Archived categories cannot be used")
	}
	return nil
}

func sameUUIDPointer(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
