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

func parseCreateProductInput(req CreateProductRequest) (*productCreateInput, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
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
		SKU:                req.SKU,
		Barcode:            req.Barcode,
		Name:               req.Name,
		Description:        req.Description,
		CategoryID:         categoryID,
		Price:              price,
		Cost:               decimal.Zero,
		TaxRate:            decimal.Zero,
		IsActive:           true,
		ImageURL:           req.ImageURL,
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
		ProductID      string `json:"product_id"`
		AdjustmentType string `json:"adjustment_type"`
		Quantity       string `json:"quantity"`
		Reason         string `json:"reason"`
	}

	if err := c.BodyParser(&req); err != nil {
		return nil, productJSONError(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, productJSONError(c, fiber.StatusBadRequest, "INVALID_PRODUCT_ID", "Invalid product ID format")
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
		return nil, productJSONError(c, fiber.StatusBadRequest, "INVALID_ADJUSTMENT_TYPE", "Invalid adjustment type. Valid types: purchase, adjustment, damage, loss, count")
	}

	quantity, err := parseDecimalField(req.Quantity, "quantity", false)
	if err != nil {
		return nil, productJSONError(c, fiber.StatusBadRequest, "INVALID_QUANTITY", "Invalid quantity")
	}

	if (adjType == models.AdjustmentDamage || adjType == models.AdjustmentLoss || (adjType == models.AdjustmentAdjustment && quantity.LessThan(decimal.Zero))) && strings.TrimSpace(req.Reason) == "" {
		return nil, productJSONError(c, fiber.StatusBadRequest, "REASON_REQUIRED", "Reason is required for damaged, lost, or correction removals")
	}

	var reason *string
	if req.Reason != "" {
		reason = &req.Reason
	}

	return &stockAdjustmentRequest{
		ProductID:      productID,
		AdjustmentType: adjType,
		Quantity:       quantity,
		Reason:         reason,
	}, nil
}
