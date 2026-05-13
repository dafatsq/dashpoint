package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// List handles GET /api/v1/products
func (h *ProductHandler) List(c *fiber.Ctx) error {
	filter, page, perPage, err := parseListFilter(c)
	if err != nil {
		return productJSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	products, total, err := h.productRepo.List(c.Context(), filter)
	if err != nil {
		return productInternalError(c, err, "Failed to list products", "Failed to retrieve products")
	}

	responses := make([]ProductResponse, len(products))
	for i, p := range products {
		responses[i] = h.toProductResponse(p)
	}

	totalPages := (total + perPage - 1) / perPage
	return c.JSON(fiber.Map{
		"products":    responses,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
	})
}

// Get handles GET /api/v1/products/:id
func (h *ProductHandler) Get(c *fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id", "INVALID_ID", "Invalid product ID format")
	if err != nil {
		return err
	}

	product, repoErr := h.productRepo.GetByID(c.Context(), id)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get product", "Failed to retrieve product")
	}
	if product == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Product not found")
	}

	return c.JSON(fiber.Map{"product": h.toProductResponse(product)})
}

// Lookup handles GET /api/v1/products/lookup
func (h *ProductHandler) Lookup(c *fiber.Ctx) error {
	code := c.Query("code", "")
	if code == "" {
		return productJSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Code (barcode or SKU) is required")
	}

	product, err := h.productRepo.Lookup(c.Context(), code)
	if err != nil {
		return productInternalError(c, err, "Failed to lookup product", "Failed to lookup product")
	}
	if product == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Product not found")
	}

	return c.JSON(fiber.Map{"product": h.toProductResponse(product)})
}
