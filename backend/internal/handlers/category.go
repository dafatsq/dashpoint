package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// CategoryHandler handles category endpoints
type CategoryHandler struct {
	categoryRepo *repository.CategoryRepository
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(categoryRepo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		categoryRepo: categoryRepo,
	}
}

// CategoryResponse represents a category in API responses
type CategoryResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ParentID     *string `json:"parent_id,omitempty"`
	SortOrder    int     `json:"sort_order"`
	IsActive     bool      `json:"is_active"`
	ProductCount *int      `json:"product_count,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateCategoryRequest represents the request to create a category
type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id"`
	SortOrder   *int    `json:"sort_order"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

// List handles GET /api/v1/categories
func (h *CategoryHandler) List(c *fiber.Ctx) error {
	status := c.Query("status", "")
	if status == "" {
		if c.Query("active_only", "true") == "true" {
			status = "active"
		} else {
			status = "all"
		}
	}

	categories, err := h.categoryRepo.List(c.Context(), status)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list categories")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve categories",
		})
	}

	responses := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = h.toCategoryResponse(cat)
		// Get product count
		count, _ := h.categoryRepo.GetProductCount(c.Context(), cat.ID)
		responses[i].ProductCount = &count
	}

	return c.JSON(fiber.Map{
		"categories": responses,
	})
}

// Get handles GET /api/v1/categories/:id
func (h *CategoryHandler) Get(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid category ID format",
		})
	}

	category, err := h.categoryRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve category",
		})
	}

	if category == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Category not found",
		})
	}

	response := h.toCategoryResponse(category)
	count, _ := h.categoryRepo.GetProductCount(c.Context(), id)
	response.ProductCount = &count

	return c.JSON(fiber.Map{
		"category": response,
	})
}

// Create handles POST /api/v1/categories
func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Name is required",
		})
	}

	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
		SortOrder:   0,
	}

	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	if req.ParentID != nil && *req.ParentID != "" {
		parentID, err := uuid.Parse(*req.ParentID)
		if err == nil {
			category.ParentID = &parentID
		}
	}

	if err := h.categoryRepo.Create(c.Context(), category); err != nil {
		log.Error().Err(err).Msg("Failed to create category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create category",
		})
	}

	// Audit log with new values
	newValues := map[string]interface{}{
		"affected_category": category.Name,
		"name":              category.Name,
	}
	audit.LogWithValues(c, models.AuditActionCategoryCreate, models.AuditEntityCategory, category.ID.String(), "Created category: "+category.Name, nil, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Category created successfully",
		"category": h.toCategoryResponse(category),
	})
}

// Update handles PATCH /api/v1/categories/:id
func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid category ID format",
		})
	}

	category, err := h.categoryRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve category",
		})
	}
	if category == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Category not found",
		})
	}

	// Capture old values for audit
	oldValues := map[string]interface{}{
		"affected_category": category.Name,
		"name":              category.Name,
		"is_active":         category.IsActive,
	}

	var req UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	if req.Name != nil {
		category.Name = *req.Name
	}

	if req.Description != nil {
		category.Description = req.Description
	}

	if req.ParentID != nil {
		if *req.ParentID == "" {
			category.ParentID = nil
		} else {
			parentID, err := uuid.Parse(*req.ParentID)
			if err == nil {
				category.ParentID = &parentID
			}
		}
	}

	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := h.categoryRepo.Update(c.Context(), category); err != nil {
		log.Error().Err(err).Msg("Failed to update category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update category",
		})
	}

	// Audit log with old/new values
	newValues := map[string]interface{}{
		"affected_category": category.Name,
		"name":              category.Name,
		"is_active":         category.IsActive,
	}
	audit.LogWithValues(c, models.AuditActionCategoryUpdate, models.AuditEntityCategory, id.String(), "Updated category: "+category.Name, oldValues, newValues)

	return c.JSON(fiber.Map{
		"message":  "Category updated successfully",
		"category": h.toCategoryResponse(category),
	})
}

// Delete handles DELETE /api/v1/categories/:id
func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid category ID format",
		})
	}
	// Category can be archived even if it has products (soft-delete)
	// Fetch category for audit
	category, _ := h.categoryRepo.GetByID(c.Context(), id)
	categoryName := "Unknown"
	if category != nil {
		categoryName = category.Name
	}

	if err := h.categoryRepo.Delete(c.Context(), id); err != nil {
		log.Error().Err(err).Msg("Failed to delete category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to delete category",
		})
	}

	// Audit log
	audit.LogWithValues(c, models.AuditActionCategoryArchive, models.AuditEntityCategory, id.String(), "Archived category: "+categoryName, map[string]interface{}{
		"affected_category": categoryName,
	}, nil)

	return c.JSON(fiber.Map{
		"message": "Category deleted successfully",
	})
}

// PermanentDelete handles DELETE /api/v1/categories/:id/permanent
func (h *CategoryHandler) PermanentDelete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid category ID format",
		})
	}

	// Check if category has products
	count, _ := h.categoryRepo.GetProductCount(c.Context(), id)
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "HAS_PRODUCTS",
			"message": "Cannot delete category with products. Move or delete products first.",
		})
	}

	// Fetch category for audit
	category, _ := h.categoryRepo.GetByID(c.Context(), id)
	categoryName := "Unknown"
	if category != nil {
		categoryName = category.Name
	}

	if err := h.categoryRepo.PermanentDelete(c.Context(), id); err != nil {
		log.Error().Err(err).Msg("Failed to permanently delete category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to permanently delete category",
		})
	}

	// Audit log
	audit.LogWithValues(c, models.AuditActionCategoryDelete, models.AuditEntityCategory, id.String(), "Permanently deleted category: "+categoryName, map[string]interface{}{
		"affected_category": categoryName,
	}, nil)

	return c.JSON(fiber.Map{
		"message": "Category permanently deleted successfully",
	})
}

func (h *CategoryHandler) toCategoryResponse(cat *models.Category) CategoryResponse {
	response := CategoryResponse{
		ID:          cat.ID.String(),
		Name:        cat.Name,
		Description: cat.Description,
		SortOrder:   cat.SortOrder,
		IsActive:    cat.IsActive,
		UpdatedAt:   cat.UpdatedAt,
	}

	if cat.ParentID != nil {
		parentIDStr := cat.ParentID.String()
		response.ParentID = &parentIDStr
	}

	return response
}
