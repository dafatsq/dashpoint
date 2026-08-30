package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

func parseListFilter(c *fiber.Ctx) (repository.ProductFilter, int, int, error) {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	search := c.Query("search", "")
	categoryIDStr := c.Query("category_id", "")
	activeOnlyStr := c.Query("active_only", "true")
	lowStock := c.Query("low_stock", "false") == "true"
	sortBy := c.Query("sort_by", "name")
	sortDirection := c.Query("sort_direction", "asc")
	validSortFields := map[string]bool{
		"name": true, "sku": true, "category": true, "price": true, "tax_rate": true, "stock": true,
	}
	if !validSortFields[sortBy] || (sortDirection != "asc" && sortDirection != "desc") {
		return repository.ProductFilter{}, 0, 0, fmt.Errorf("invalid sort parameters")
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := repository.ProductFilter{
		Search:        search,
		LowStock:      lowStock,
		Limit:         perPage,
		Offset:        (page - 1) * perPage,
		SortBy:        sortBy,
		SortDirection: sortDirection,
	}

	if categoryIDStr != "" {
		catID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			return repository.ProductFilter{}, 0, 0, err
		}
		filter.CategoryID = &catID
	}

	if activeOnlyStr == "true" {
		active := true
		filter.IsActive = &active
	} else if activeOnlyStr == "false" {
		active := false
		filter.IsActive = &active
	}

	return filter, page, perPage, nil
}

// Product text field limits match the database columns (products.name
// varchar(255), sku/barcode varchar(100)) so oversized input gets a clear
// validation error instead of an opaque DB length 500.
const (
	maxProductNameLen    = 255
	maxProductSKULen     = 100
	maxProductBarcodeLen = 100
)

func validateProductTextLength(field, value string, maxLen int) error {
	if len(value) > maxLen {
		return fmt.Errorf("%s must be at most %d characters", field, maxLen)
	}
	return nil
}

func parseCreateProductInput(req CreateProductRequest) (*productCreateInput, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if err := validateProductTextLength("name", req.Name, maxProductNameLen); err != nil {
		return nil, err
	}
	if req.SKU != nil {
		if err := validateProductTextLength("sku", *req.SKU, maxProductSKULen); err != nil {
			return nil, err
		}
	}
	if req.Barcode != nil {
		if err := validateProductTextLength("barcode", *req.Barcode, maxProductBarcodeLen); err != nil {
			return nil, err
		}
	}

	price, err := parseDecimalField(req.Price, "price", false)
	if err != nil {
		return nil, err
	}
	cost, err := parseOptionalDecimalField(req.Cost, "cost", false)
	if err != nil {
		return nil, err
	}
	taxRate, err := parseOptionalDecimalField(req.TaxRate, "tax_rate", false)
	if err != nil {
		return nil, err
	}
	categoryID, err := parseOptionalUUIDField(req.CategoryID, "category_id")
	if err != nil {
		return nil, err
	}
	initialQuantity, err := parseOptionalDecimalField(req.InitialQuantity, "initial_quantity", false)
	if err != nil {
		return nil, err
	}
	lowStockThreshold, err := parseOptionalDecimalField(req.LowStockThreshold, "low_stock_threshold", false)
	if err != nil {
		return nil, err
	}

	product := &models.Product{
		SKU:         req.SKU,
		Barcode:     req.Barcode,
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  categoryID,
		Price:       price,
		Cost:        decimal.Zero,
		TaxRate:     decimal.Zero,
		IsActive:    true,
		ImageURL:    req.ImageURL,
	}
	if cost != nil {
		product.Cost = *cost
	}
	if taxRate != nil {
		product.TaxRate = *taxRate
	}

	return &productCreateInput{
		product:           product,
		initialQuantity:   initialQuantity,
		lowStockThreshold: lowStockThreshold,
	}, nil
}

func applyUpdateProductRequest(product *models.Product, req UpdateProductRequest) error {
	if req.Name != nil {
		if err := validateProductTextLength("name", *req.Name, maxProductNameLen); err != nil {
			return err
		}
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.CategoryID != nil {
		if *req.CategoryID == "" {
			product.CategoryID = nil
		} else {
			categoryID, err := parseOptionalUUIDField(req.CategoryID, "category_id")
			if err != nil {
				return err
			}
			product.CategoryID = categoryID
		}
	}
	if req.Price != nil {
		price, err := parseDecimalField(*req.Price, "price", false)
		if err != nil {
			return err
		}
		product.Price = price
	}
	if req.Cost != nil {
		cost, err := parseDecimalField(*req.Cost, "cost", false)
		if err != nil {
			return err
		}
		product.Cost = cost
	}
	if req.TaxRate != nil {
		taxRate, err := parseDecimalField(*req.TaxRate, "tax_rate", false)
		if err != nil {
			return err
		}
		product.TaxRate = taxRate
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}
	return nil
}

func parseStockAdjustmentRequest(c *fiber.Ctx) (*stockAdjustmentRequest, error) {
	var req struct {
		ProductID         string  `json:"product_id"`
		AdjustmentType    string  `json:"adjustment_type"`
		Quantity          string  `json:"quantity"`
		Reason            string  `json:"reason"`
		ExpectedUpdatedAt *string `json:"expected_updated_at"`
	}

	if err := parseStrictJSONBody(c, &req, productMaxJSONBodyBytes); err != nil {
		return nil, newProductRequestError(fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, newProductRequestError(fiber.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID format")
	}

	adjType := models.AdjustmentType(req.AdjustmentType)
	validTypes := map[models.AdjustmentType]bool{
		models.AdjustmentPurchase:   true,
		models.AdjustmentAdjustment: true,
		models.AdjustmentDamage:     true,
		models.AdjustmentLoss:       true,
		models.AdjustmentCount:      true,
	}
	if !validTypes[adjType] {
		return nil, newProductRequestError(fiber.StatusBadRequest, "INVALID_ADJUSTMENT_TYPE", "Invalid adjustment type. Valid types: purchase, adjustment, damage, loss, count")
	}

	quantity, err := parseDecimalField(req.Quantity, "quantity", true)
	if err != nil {
		return nil, newProductRequestError(fiber.StatusBadRequest, "INVALID_QUANTITY", "Invalid quantity")
	}
	if err := validateStockAdjustmentQuantity(adjType, quantity); err != nil {
		return nil, newProductRequestError(fiber.StatusBadRequest, "INVALID_QUANTITY", err.Error())
	}

	var reason *string
	if strings.TrimSpace(req.Reason) != "" {
		trimmedReason := strings.TrimSpace(req.Reason)
		reason = &trimmedReason
	}

	return &stockAdjustmentRequest{
		ProductID:         productID,
		AdjustmentType:    adjType,
		Quantity:          quantity,
		Reason:            reason,
		ExpectedUpdatedAt: req.ExpectedUpdatedAt,
	}, nil
}

func validateStockAdjustmentQuantity(adjType models.AdjustmentType, quantity decimal.Decimal) error {
	switch adjType {
	case models.AdjustmentPurchase, models.AdjustmentDamage, models.AdjustmentLoss:
		if quantity.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("quantity must be greater than zero")
		}
	case models.AdjustmentCount:
		if quantity.LessThan(decimal.Zero) {
			return fmt.Errorf("quantity cannot be negative for stock count")
		}
	case models.AdjustmentAdjustment:
		if quantity.Equal(decimal.Zero) {
			return fmt.Errorf("quantity cannot be zero")
		}
	}
	return nil
}

func parseInventoryThresholdUpdateRequest(c *fiber.Ctx) (*inventoryThresholdUpdateRequest, error) {
	var req struct {
		LowStockThreshold string  `json:"low_stock_threshold"`
		ExpectedUpdatedAt *string `json:"expected_updated_at"`
	}

	if err := parseStrictJSONBody(c, &req, productMaxJSONBodyBytes); err != nil {
		return nil, productJSONError(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	if strings.TrimSpace(req.LowStockThreshold) == "" {
		return nil, productJSONError(c, fiber.StatusBadRequest, "LOW_STOCK_THRESHOLD_REQUIRED", "low_stock_threshold is required")
	}

	threshold, err := parseDecimalField(req.LowStockThreshold, "low_stock_threshold", false)
	if err != nil {
		return nil, productJSONError(c, fiber.StatusBadRequest, "INVALID_LOW_STOCK_THRESHOLD", "Invalid low_stock_threshold")
	}

	return &inventoryThresholdUpdateRequest{
		LowStockThreshold: threshold,
		ExpectedUpdatedAt: req.ExpectedUpdatedAt,
	}, nil
}
