package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

type categoryAPIError struct {
	Status  int
	Code    string
	Message string
}

func invalidCategoryIDError() *categoryAPIError {
	return &categoryAPIError{
		Status:  fiber.StatusBadRequest,
		Code:    "INVALID_ID",
		Message: "Invalid category ID format",
	}
}

func respondCategoryError(c *fiber.Ctx, err *categoryAPIError) error {
	return c.Status(err.Status).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}

func respondCategoryInternalError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": message,
		},
	})
}

func parseCategoryIDParam(c *fiber.Ctx) (uuid.UUID, *categoryAPIError) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, invalidCategoryIDError()
	}
	return id, nil
}

func categoryResponse(category *models.Category, productCount *int) CategoryResponse {
	response := CategoryResponse{
		ID:           category.ID.String(),
		Name:         category.Name,
		Description:  category.Description,
		ProductCount: productCount,
		IsActive:     category.IsActive,
		CreatedAt:    category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	return response
}

func isCategoryNotFound(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not found")
}

func categoryNotFoundResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Category not found",
		},
	})
}

func validateCategoryName(name string) *categoryAPIError {
	if strings.TrimSpace(name) == "" {
		return &categoryAPIError{
			Status:  fiber.StatusBadRequest,
			Code:    "VALIDATION_ERROR",
			Message: "Category name is required",
		}
	}
	return nil
}

func collectCategoryUUIDs(categories []*models.Category) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(categories))
	for _, category := range categories {
		if category == nil {
			continue
		}
		ids = append(ids, category.ID)
	}
	return ids
}

func productCountPointer(counts map[uuid.UUID]int, id uuid.UUID) *int {
	count := counts[id]
	return &count
}

func getCategoryProductCounts(store categoryHandlerStore, ctx *fiber.Ctx, ids []uuid.UUID) (map[uuid.UUID]int, error) {
	counts, err := store.GetProductCounts(ctx.UserContext(), ids)
	if err == nil {
		return counts, nil
	}

	// Compatibility fallback for stores that only expose the single-count path.
	if len(ids) == 0 {
		return map[uuid.UUID]int{}, nil
	}

	fallback := make(map[uuid.UUID]int, len(ids))
	for _, id := range ids {
		count, countErr := store.GetProductCount(ctx.UserContext(), id)
		if countErr != nil {
			return nil, errors.New("failed to load category product counts")
		}
		fallback[id] = count
	}
	return fallback, nil
}
