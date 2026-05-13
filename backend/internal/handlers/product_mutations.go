package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// Create handles POST /api/v1/products
func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return productJSONError(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	input, err := parseCreateProductInput(req)
	if err != nil {
		return productJSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if req.SKU != nil && *req.SKU != "" {
		existingProduct, repoErr := h.productRepo.GetBySKUIncludingInactive(c.Context(), *req.SKU)
		if repoErr != nil {
			return productInternalError(c, repoErr, "Failed to check SKU", "Failed to validate SKU")
		}
		if existingProduct != nil {
			if !existingProduct.IsActive {
				reactivated, reactivateErr := h.reactivateProduct(c, existingProduct, req, input.product.Price)
				if reactivateErr != nil {
					return reactivateErr
				}
				return c.Status(fiber.StatusCreated).JSON(fiber.Map{
					"message": "Product reactivated successfully",
					"product": h.toProductResponse(reactivated),
				})
			}
			return productJSONError(c, fiber.StatusConflict, "SKU_EXISTS", "SKU is already in use by an active product")
		}
	}

	if req.Barcode != nil && *req.Barcode != "" {
		existingProduct, repoErr := h.productRepo.GetByBarcodeIncludingInactive(c.Context(), *req.Barcode)
		if repoErr != nil {
			return productInternalError(c, repoErr, "Failed to check barcode", "Failed to validate barcode")
		}
		if existingProduct != nil {
			if existingProduct.IsActive {
				return productJSONError(c, fiber.StatusConflict, "BARCODE_EXISTS", "Barcode is already in use by an active product")
			}
			if req.SKU == nil || *req.SKU == "" || existingProduct.SKU == nil || *existingProduct.SKU != *req.SKU {
				return productJSONError(c, fiber.StatusConflict, "BARCODE_EXISTS", "Barcode is used by an inactive product with a different SKU. Use the original SKU to reactivate it.")
			}
		}
	}

	if err := h.productRepo.Create(c.Context(), input.product, input.initialQuantity); err != nil {
		return productInternalError(c, err, "Failed to create product", "Failed to create product")
	}

	if input.lowStockThreshold != nil && input.product.TrackInventory {
		if err := h.inventoryRepo.UpdateThresholds(c.Context(), input.product.ID, *input.lowStockThreshold, decimal.Zero); err != nil {
			log.Warn().Err(err).Str("product_id", input.product.ID.String()).Msg("Failed to update inventory thresholds after product creation")
		}
	}

	created, err := h.productRepo.GetByID(c.Context(), input.product.ID)
	if err != nil {
		return productInternalError(c, err, "Failed to fetch created product", "Failed to create product")
	}
	if created == nil {
		created = input.product
	}

	logProductAudit(c, models.AuditActionProductCreate, created.ID.String(), "Created product: "+created.Name, nil, buildProductAuditValues(created))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Product created successfully",
		"product": h.toProductResponse(created),
	})
}

// Update handles PATCH /api/v1/products/:id
func (h *ProductHandler) Update(c *fiber.Ctx) error {
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

	oldValues := buildProductAuditValues(product)

	var req UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return productJSONError(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	if req.SKU != nil {
		if *req.SKU != "" {
			exists, err := h.productRepo.SKUExists(c.Context(), *req.SKU, &id)
			if err != nil {
				return productInternalError(c, err, "Failed to validate SKU", "Failed to update product")
			}
			if exists {
				return productJSONError(c, fiber.StatusConflict, "SKU_EXISTS", "SKU is already in use")
			}
		}
		product.SKU = req.SKU
	}

	if req.Barcode != nil && *req.Barcode != "" {
		exists, err := h.productRepo.BarcodeExists(c.Context(), *req.Barcode, &id)
		if err != nil {
			return productInternalError(c, err, "Failed to validate barcode", "Failed to update product")
		}
		if exists {
			return productJSONError(c, fiber.StatusConflict, "BARCODE_EXISTS", "Barcode is already in use")
		}
		product.Barcode = req.Barcode
	} else if req.Barcode != nil {
		product.Barcode = req.Barcode
	}

	if err := applyUpdateProductRequest(product, req); err != nil {
		return productJSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if req.ImageURL != nil {
		oldImageURL := ""
		if product.ImageURL != nil {
			oldImageURL = *product.ImageURL
		}
		if *req.ImageURL == "" {
			product.ImageURL = nil
		} else {
			product.ImageURL = req.ImageURL
		}
		newImageURL := ""
		if product.ImageURL != nil {
			newImageURL = *product.ImageURL
		}
		if oldImageURL != "" && oldImageURL != newImageURL {
			h.deleteImageFile(oldImageURL)
		}
	}

	if err := h.productRepo.Update(c.Context(), product); err != nil {
		return productInternalError(c, err, "Failed to update product", "Failed to update product")
	}

	newValues := buildProductAuditValues(product)
	oldIsActive, ok := oldValues["is_active"].(bool)
	isRestore := req.IsActive != nil && *req.IsActive && ok && !oldIsActive
	action := models.AuditActionProductUpdate
	actionMsg := "Updated product: " + product.Name
	if isRestore {
		action = models.AuditActionProductRestore
		actionMsg = "Restored product: " + product.Name
	}

	logProductAudit(c, action, id.String(), actionMsg, oldValues, newValues)

	updated, err := h.productRepo.GetByID(c.Context(), id)
	if err != nil {
		return productInternalError(c, err, "Failed to fetch updated product", "Failed to update product")
	}
	if updated == nil {
		updated = product
	}

	return c.JSON(fiber.Map{
		"message": "Product updated successfully",
		"product": h.toProductResponse(updated),
	})
}
