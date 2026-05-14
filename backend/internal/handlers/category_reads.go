package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h *CategoryHandler) List(c *fiber.Ctx) error {
	status := c.Query("status")
	if status == "" {
		if c.Query("active_only") == "false" {
			status = "all"
		} else {
			status = "active"
		}
	}

	categories, err := h.categoryRepo.List(c.UserContext(), status)
	if err != nil {
		return respondCategoryInternalError(c, "Failed to fetch categories")
	}

	counts, err := getCategoryProductCounts(h.categoryRepo, c, collectCategoryUUIDs(categories))
	if err != nil {
		return respondCategoryInternalError(c, "Failed to fetch categories")
	}

	responses := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, categoryResponse(category, productCountPointer(counts, category.ID)))
	}

	return c.JSON(fiber.Map{"categories": responses})
}

func (h *CategoryHandler) Get(c *fiber.Ctx) error {
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

	count, err := h.categoryRepo.GetProductCount(c.UserContext(), id)
	if err != nil {
		return respondCategoryInternalError(c, "Failed to fetch category")
	}

	return c.JSON(fiber.Map{
		"category": categoryResponse(category, &count),
	})
}
