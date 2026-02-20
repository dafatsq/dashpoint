package handlers

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// ProductHandler handles product endpoints
type ProductHandler struct {
	productRepo   *repository.ProductRepository
	inventoryRepo *repository.InventoryRepository
	categoryRepo  *repository.CategoryRepository
	uploadDir     string
}

// NewProductHandler creates a new product handler
func NewProductHandler(
	productRepo *repository.ProductRepository,
	inventoryRepo *repository.InventoryRepository,
	categoryRepo *repository.CategoryRepository,
	uploadDir string,
) *ProductHandler {
	return &ProductHandler{
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		categoryRepo:  categoryRepo,
		uploadDir:     uploadDir,
	}
}

// deleteImageFile removes an uploaded image file from disk.
// imageURL is expected to be in the form "/uploads/<filename>".
func (h *ProductHandler) deleteImageFile(imageURL string) {
	if imageURL == "" || h.uploadDir == "" {
		return
	}
	// Strip leading "/uploads/" prefix to get just the filename
	filename := strings.TrimPrefix(imageURL, "/uploads/")
	// Sanitize to prevent directory traversal
	filename = filepath.Base(filename)
	if filename == "." || filename == "" {
		return
	}
	filePath := filepath.Join(h.uploadDir, filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		log.Warn().Err(err).Str("file", filePath).Msg("Failed to delete orphaned image file")
	}
}

// ProductResponse represents a product in API responses
type ProductResponse struct {
	ID                 string             `json:"id"`
	SKU                *string            `json:"sku,omitempty"`
	Barcode            *string            `json:"barcode,omitempty"`
	Name               string             `json:"name"`
	Description        *string            `json:"description,omitempty"`
	CategoryID         *string            `json:"category_id,omitempty"`
	CategoryName       *string            `json:"category_name,omitempty"`
	Price              string             `json:"price"`
	Cost               string             `json:"cost"`
	TaxRate            string             `json:"tax_rate"`
	Unit               string             `json:"unit"`
	IsActive           bool               `json:"is_active"`
	TrackInventory     bool               `json:"track_inventory"`
	AllowNegativeStock bool               `json:"allow_negative_stock"`
	ImageURL           *string            `json:"image_url,omitempty"`
	Inventory          *InventoryResponse `json:"inventory,omitempty"`
	CreatedAt          string             `json:"created_at"`
	UpdatedAt          string             `json:"updated_at"`
}

// InventoryResponse represents inventory in API responses
type InventoryResponse struct {
	Quantity          string `json:"quantity"`
	AvailableQuantity string `json:"available_quantity"`
	LowStockThreshold string `json:"low_stock_threshold"`
	IsLowStock        bool   `json:"is_low_stock"`
}

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	SKU                *string `json:"sku"`
	Barcode            *string `json:"barcode"`
	Name               string  `json:"name"`
	Description        *string `json:"description"`
	CategoryID         *string `json:"category_id"`
	Price              string  `json:"price"`
	Cost               *string `json:"cost"`
	TaxRate            *string `json:"tax_rate"`
	Unit               *string `json:"unit"`
	TrackInventory     *bool   `json:"track_inventory"`
	AllowNegativeStock *bool   `json:"allow_negative_stock"`
	InitialQuantity    *string `json:"initial_quantity"`
	LowStockThreshold  *string `json:"low_stock_threshold"`
	ImageURL           *string `json:"image_url"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	SKU                *string `json:"sku"`
	Barcode            *string `json:"barcode"`
	Name               *string `json:"name"`
	Description        *string `json:"description"`
	CategoryID         *string `json:"category_id"`
	Price              *string `json:"price"`
	Cost               *string `json:"cost"`
	TaxRate            *string `json:"tax_rate"`
	Unit               *string `json:"unit"`
	IsActive           *bool   `json:"is_active"`
	TrackInventory     *bool   `json:"track_inventory"`
	AllowNegativeStock *bool   `json:"allow_negative_stock"`
	ImageURL           *string `json:"image_url"`
}

// List handles GET /api/v1/products
func (h *ProductHandler) List(c *fiber.Ctx) error {
	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	search := c.Query("search", "")
	categoryIDStr := c.Query("category_id", "")
	activeOnlyStr := c.Query("active_only", "true")
	lowStock := c.Query("low_stock", "false") == "true"

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := repository.ProductFilter{
		Search:   search,
		LowStock: lowStock,
		Limit:    perPage,
		Offset:   (page - 1) * perPage,
	}

	if categoryIDStr != "" {
		catID, err := uuid.Parse(categoryIDStr)
		if err == nil {
			filter.CategoryID = &catID
		}
	}

	if activeOnlyStr == "true" {
		active := true
		filter.IsActive = &active
	} else if activeOnlyStr == "false" {
		active := false
		filter.IsActive = &active
	}

	products, total, err := h.productRepo.List(c.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list products")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve products",
		})
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
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid product ID format",
		})
	}

	product, err := h.productRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get product")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve product",
		})
	}

	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Product not found",
		})
	}

	return c.JSON(fiber.Map{
		"product": h.toProductResponse(product),
	})
}

// Lookup handles GET /api/v1/products/lookup
func (h *ProductHandler) Lookup(c *fiber.Ctx) error {
	code := c.Query("code", "")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Code (barcode or SKU) is required",
		})
	}

	product, err := h.productRepo.Lookup(c.Context(), code)
	if err != nil {
		log.Error().Err(err).Msg("Failed to lookup product")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to lookup product",
		})
	}

	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Product not found",
		})
	}

	return c.JSON(fiber.Map{
		"product": h.toProductResponse(product),
	})
}

// Create handles POST /api/v1/products
func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Name is required",
		})
	}

	// Parse price
	price, err := decimal.NewFromString(req.Price)
	if err != nil || price.LessThan(decimal.Zero) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Invalid price",
		})
	}

	// Check SKU uniqueness - if it belongs to an inactive product, offer to reactivate
	if req.SKU != nil && *req.SKU != "" {
		existingProduct, err := h.productRepo.GetBySKUIncludingInactive(c.Context(), *req.SKU)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check SKU")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to validate SKU",
			})
		}
		if existingProduct != nil {
			if !existingProduct.IsActive {
				// Reactivate the existing product instead of creating a new one
				existingProduct.IsActive = true
				existingProduct.Name = req.Name
				existingProduct.Description = req.Description
				existingProduct.Barcode = req.Barcode
				existingProduct.Price = price
				if req.Cost != nil {
					cost, _ := decimal.NewFromString(*req.Cost)
					existingProduct.Cost = cost
				}
				if req.TaxRate != nil {
					taxRate, _ := decimal.NewFromString(*req.TaxRate)
					existingProduct.TaxRate = taxRate
				}
				if req.Unit != nil {
					existingProduct.Unit = *req.Unit
				}
				if req.CategoryID != nil && *req.CategoryID != "" {
					catID, err := uuid.Parse(*req.CategoryID)
					if err == nil {
						existingProduct.CategoryID = &catID
					}
				}

				if err := h.productRepo.Update(c.Context(), existingProduct); err != nil {
					log.Error().Err(err).Msg("Failed to reactivate product")
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"code":    "INTERNAL_ERROR",
						"message": "Failed to reactivate product",
					})
				}

				// Fetch updated product with inventory
				reactivated, _ := h.productRepo.GetByID(c.Context(), existingProduct.ID)
				if reactivated != nil {
					existingProduct = reactivated
				}

				return c.Status(fiber.StatusCreated).JSON(fiber.Map{
					"message": "Product reactivated successfully",
					"product": h.toProductResponse(existingProduct),
				})
			}
			// SKU belongs to an active product
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "SKU_EXISTS",
				"message": "SKU is already in use by an active product",
			})
		}
	}

	// Check barcode uniqueness
	if req.Barcode != nil && *req.Barcode != "" {
		existingProduct, err := h.productRepo.GetByBarcodeIncludingInactive(c.Context(), *req.Barcode)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check barcode")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to validate barcode",
			})
		}
		if existingProduct != nil {
			if existingProduct.IsActive {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"code":    "BARCODE_EXISTS",
					"message": "Barcode is already in use by an active product",
				})
			}
			// Barcode belongs to an inactive product - check if SKU also matches for reactivation
			// If SKU doesn't match, we can't reuse the barcode
			if req.SKU == nil || *req.SKU == "" || existingProduct.SKU == nil || *existingProduct.SKU != *req.SKU {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"code":    "BARCODE_EXISTS",
					"message": "Barcode is used by an inactive product with a different SKU. Use the original SKU to reactivate it.",
				})
			}
			// If we get here, SKU check above will handle reactivation
		}
	}

	// Build product
	product := &models.Product{
		SKU:                req.SKU,
		Barcode:            req.Barcode,
		Name:               req.Name,
		Description:        req.Description,
		Price:              price,
		Cost:               decimal.Zero,
		TaxRate:            decimal.Zero,
		Unit:               "pcs",
		IsActive:           true,
		TrackInventory:     true,
		AllowNegativeStock: false,
	}

	// Parse optional fields
	if req.Cost != nil {
		cost, _ := decimal.NewFromString(*req.Cost)
		product.Cost = cost
	}

	if req.TaxRate != nil {
		taxRate, _ := decimal.NewFromString(*req.TaxRate)
		product.TaxRate = taxRate
	}

	if req.Unit != nil {
		product.Unit = *req.Unit
	}

	if req.TrackInventory != nil {
		product.TrackInventory = *req.TrackInventory
	}

	if req.AllowNegativeStock != nil {
		product.AllowNegativeStock = *req.AllowNegativeStock
	}

	if req.CategoryID != nil && *req.CategoryID != "" {
		catID, err := uuid.Parse(*req.CategoryID)
		if err == nil {
			product.CategoryID = &catID
		}
	}

	if req.ImageURL != nil {
		product.ImageURL = req.ImageURL
	}

	// Parse initial quantity
	var initialQty *decimal.Decimal
	if req.InitialQuantity != nil {
		qty, err := decimal.NewFromString(*req.InitialQuantity)
		if err == nil {
			initialQty = &qty
		}
	}

	// Create product
	if err := h.productRepo.Create(c.Context(), product, initialQty); err != nil {
		log.Error().Err(err).Msg("Failed to create product")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create product",
		})
	}

	// Set low stock threshold if provided
	if req.LowStockThreshold != nil && product.TrackInventory {
		threshold, err := decimal.NewFromString(*req.LowStockThreshold)
		if err == nil {
			h.inventoryRepo.UpdateThresholds(c.Context(), product.ID, threshold, decimal.Zero)
		}
	}

	// Fetch created product
	created, _ := h.productRepo.GetByID(c.Context(), product.ID)
	if created != nil {
		product = created
	}

	// Audit log with new values
	newValues := map[string]interface{}{
		"name":  product.Name,
		"price": product.Price.String(),
	}
	if product.SKU != nil {
		newValues["sku"] = *product.SKU
	}
	if product.ImageURL != nil {
		newValues["image_url"] = *product.ImageURL
	}
	audit.LogWithValues(c, models.AuditActionProductCreate, models.AuditEntityProduct, product.ID.String(), "Created product: "+product.Name, nil, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Product created successfully",
		"product": h.toProductResponse(product),
	})
}

// Update handles PATCH /api/v1/products/:id
func (h *ProductHandler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid product ID format",
		})
	}

	// Get existing product
	product, err := h.productRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get product")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve product",
		})
	}
	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Product not found",
		})
	}

	// Capture old values for audit
	oldValues := map[string]interface{}{
		"name":      product.Name,
		"price":     product.Price.String(),
		"is_active": product.IsActive,
	}
	if product.SKU != nil {
		oldValues["sku"] = *product.SKU
	}
	if product.ImageURL != nil {
		oldValues["image_url"] = *product.ImageURL
	}

	var req UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	// Update fields
	if req.SKU != nil {
		if *req.SKU != "" {
			exists, _ := h.productRepo.SKUExists(c.Context(), *req.SKU, &id)
			if exists {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"code":    "SKU_EXISTS",
					"message": "SKU is already in use",
				})
			}
		}
		product.SKU = req.SKU
	}

	if req.Barcode != nil {
		if *req.Barcode != "" {
			exists, _ := h.productRepo.BarcodeExists(c.Context(), *req.Barcode, &id)
			if exists {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"code":    "BARCODE_EXISTS",
					"message": "Barcode is already in use",
				})
			}
		}
		product.Barcode = req.Barcode
	}

	if req.Name != nil {
		product.Name = *req.Name
	}

	if req.Description != nil {
		product.Description = req.Description
	}

	if req.CategoryID != nil {
		if *req.CategoryID == "" {
			product.CategoryID = nil
		} else {
			catID, err := uuid.Parse(*req.CategoryID)
			if err == nil {
				product.CategoryID = &catID
			}
		}
	}

	if req.Price != nil {
		price, _ := decimal.NewFromString(*req.Price)
		product.Price = price
	}

	if req.Cost != nil {
		cost, _ := decimal.NewFromString(*req.Cost)
		product.Cost = cost
	}

	if req.TaxRate != nil {
		taxRate, _ := decimal.NewFromString(*req.TaxRate)
		product.TaxRate = taxRate
	}

	if req.Unit != nil {
		product.Unit = *req.Unit
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if req.TrackInventory != nil {
		product.TrackInventory = *req.TrackInventory
	}

	if req.AllowNegativeStock != nil {
		product.AllowNegativeStock = *req.AllowNegativeStock
	}

	if req.ImageURL != nil {
		// Empty string means clear the image
		oldImageURL := ""
		if product.ImageURL != nil {
			oldImageURL = *product.ImageURL
		}
		if *req.ImageURL == "" {
			product.ImageURL = nil
		} else {
			product.ImageURL = req.ImageURL
		}
		// Delete old file if image was replaced or cleared
		newImageURL := ""
		if product.ImageURL != nil {
			newImageURL = *product.ImageURL
		}
		if oldImageURL != "" && oldImageURL != newImageURL {
			h.deleteImageFile(oldImageURL)
		}
	}

	if err := h.productRepo.Update(c.Context(), product); err != nil {
		log.Error().Err(err).Msg("Failed to update product")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update product",
		})
	}

	// Audit log with old/new values
	newValues := map[string]interface{}{
		"name":      product.Name,
		"price":     product.Price.String(),
		"is_active": product.IsActive,
	}
	if product.SKU != nil {
		newValues["sku"] = *product.SKU
	}
	if product.ImageURL != nil {
		newValues["image_url"] = *product.ImageURL
	}
	audit.LogWithValues(c, models.AuditActionProductUpdate, models.AuditEntityProduct, id.String(), "Updated product: "+product.Name, oldValues, newValues)

	updated, _ := h.productRepo.GetByID(c.Context(), id)
	if updated != nil {
		product = updated
	}

	return c.JSON(fiber.Map{
		"message": "Product updated successfully",
		"product": h.toProductResponse(product),
	})
}

// Delete handles DELETE /api/v1/products/:id
func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid product ID format",
		})
	}

	if err := h.productRepo.Delete(c.Context(), id); err != nil {
		log.Error().Err(err).Msg("Failed to delete product")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to delete product",
		})
	}

	// Audit log
	audit.LogFromFiber(c, models.AuditActionProductDelete, models.AuditEntityProduct, id.String(), "Deleted product")

	return c.JSON(fiber.Map{
		"message": "Product deleted successfully",
	})
}

// PermanentDelete handles DELETE /api/v1/products/:id/permanent
func (h *ProductHandler) PermanentDelete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid product ID format",
		})
	}

	// Fetch the product first so we can clean up its image after deletion
	productToDelete, err := h.productRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get product")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve product",
		})
	}
	if productToDelete == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Product not found",
		})
	}

	// Check if product has sales history
	hasSales, err := h.productRepo.HasSalesHistory(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check sales history")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to check product sales history",
		})
	}

	if hasSales {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"code":    "HAS_SALES_HISTORY",
			"message": "Cannot permanently delete product with sales history. The product must remain archived.",
		})
	}

	if err := h.productRepo.PermanentDelete(c.Context(), id); err != nil {
		log.Error().Err(err).Msg("Failed to permanently delete product")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to permanently delete product",
		})
	}

	// Clean up image file after successful deletion
	if productToDelete.ImageURL != nil {
		h.deleteImageFile(*productToDelete.ImageURL)
	}

	return c.JSON(fiber.Map{
		"message": "Product permanently deleted",
	})
}

// GetInventory handles GET /api/v1/products/:id/inventory
func (h *ProductHandler) GetInventory(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid product ID format",
		})
	}

	inventory, err := h.inventoryRepo.GetByProductID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get inventory")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve inventory",
		})
	}

	if inventory == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Inventory not found",
		})
	}

	// Get adjustment history
	adjustments, total, _ := h.inventoryRepo.GetAdjustmentHistory(c.Context(), id, 10, 0)

	return c.JSON(fiber.Map{
		"inventory": fiber.Map{
			"product_id":          inventory.ProductID.String(),
			"quantity":            inventory.Quantity.String(),
			"available_quantity":  inventory.AvailableQuantity().String(),
			"reserved_quantity":   inventory.ReservedQuantity.String(),
			"low_stock_threshold": inventory.LowStockThreshold.String(),
			"reorder_quantity":    inventory.ReorderQuantity.String(),
			"is_low_stock":        inventory.IsLowStock(),
			"last_counted_at":     inventory.LastCountedAt,
			"last_restocked_at":   inventory.LastRestockedAt,
		},
		"recent_adjustments": adjustments,
		"total_adjustments":  total,
	})
}

// AdjustStock handles POST /api/v1/inventory/adjust
func (h *ProductHandler) AdjustStock(c *fiber.Ctx) error {
	var req struct {
		ProductID      string `json:"product_id"`
		AdjustmentType string `json:"adjustment_type"`
		Quantity       string `json:"quantity"`
		Reason         string `json:"reason"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	// Validate product ID
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_PRODUCT_ID",
			"message": "Invalid product ID format",
		})
	}

	// Validate adjustment type
	adjType := models.AdjustmentType(req.AdjustmentType)
	validTypes := map[models.AdjustmentType]bool{
		models.AdjustmentPurchase:   true,
		models.AdjustmentAdjustment: true,
		models.AdjustmentDamage:     true,
		models.AdjustmentLoss:       true,
		models.AdjustmentCount:      true,
	}
	if !validTypes[adjType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ADJUSTMENT_TYPE",
			"message": "Invalid adjustment type. Valid types: purchase, adjustment, damage, loss, count",
		})
	}

	// Parse quantity
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_QUANTITY",
			"message": "Invalid quantity",
		})
	}

	// Get current user
	userID := middleware.GetUserID(c)

	var reason *string
	if req.Reason != "" {
		reason = &req.Reason
	}

	var adjustment *models.StockAdjustment

	// Handle "count" adjustment type separately (sets absolute quantity)
	if adjType == models.AdjustmentCount {
		adjustment, err = h.inventoryRepo.SetQuantity(
			c.Context(),
			productID,
			quantity,
			reason,
			userID,
		)
	} else {
		// For damage and loss, negate the quantity (these remove stock)
		if adjType == models.AdjustmentDamage || adjType == models.AdjustmentLoss {
			quantity = quantity.Neg()
		}

		adjustment, err = h.inventoryRepo.AdjustStock(
			c.Context(),
			productID,
			adjType,
			quantity,
			reason,
			nil,
			nil,
			userID,
		)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to adjust stock")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "ADJUSTMENT_FAILED",
			"message": err.Error(),
		})
	}

	// Audit log
	product, _ := h.productRepo.GetByID(c.Context(), productID)
	productName := ""
	if product != nil {
		productName = product.Name
	}
	auditAction := models.AuditActionStockAdjust
	if adjType == models.AdjustmentCount {
		auditAction = models.AuditActionStockCount
	}
	auditValues := map[string]interface{}{
		"product_name":    productName,
		"adjustment_type": string(adjType),
		"quantity":        quantity.String(),
	}
	if req.Reason != "" {
		auditValues["reason"] = req.Reason
	}
	if adjustment != nil {
		auditValues["new_quantity"] = adjustment.QuantityAfter.String()
	}
	audit.LogWithValues(c, auditAction, models.AuditEntityInventory, productID.String(),
		"Stock adjusted: "+productName, nil, auditValues)

	return c.JSON(fiber.Map{
		"message":    "Stock adjusted successfully",
		"adjustment": adjustment,
	})
}

// GetLowStock handles GET /api/v1/inventory/low-stock
func (h *ProductHandler) GetLowStock(c *fiber.Ctx) error {
	products, err := h.inventoryRepo.GetLowStockProducts(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get low stock products")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve low stock products",
		})
	}

	return c.JSON(fiber.Map{
		"products": products,
		"count":    len(products),
	})
}

// Helper functions

func (h *ProductHandler) toProductResponse(p *models.Product) ProductResponse {
	response := ProductResponse{
		ID:                 p.ID.String(),
		SKU:                p.SKU,
		Barcode:            p.Barcode,
		Name:               p.Name,
		Description:        p.Description,
		Price:              p.Price.String(),
		Cost:               p.Cost.String(),
		TaxRate:            p.TaxRate.String(),
		Unit:               p.Unit,
		IsActive:           p.IsActive,
		TrackInventory:     p.TrackInventory,
		AllowNegativeStock: p.AllowNegativeStock,
		ImageURL:           p.ImageURL,
		CreatedAt:          p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if p.CategoryID != nil {
		catIDStr := p.CategoryID.String()
		response.CategoryID = &catIDStr
	}

	if p.Category != nil {
		response.CategoryName = &p.Category.Name
	}

	if p.Inventory != nil {
		response.Inventory = &InventoryResponse{
			Quantity:          p.Inventory.Quantity.String(),
			AvailableQuantity: p.Inventory.AvailableQuantity().String(),
			LowStockThreshold: p.Inventory.LowStockThreshold.String(),
			IsLowStock:        p.Inventory.IsLowStock(),
		}
	}

	return response
}
