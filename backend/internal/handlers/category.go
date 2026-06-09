package handlers

import (
	"context"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

type categoryHandlerStore interface {
	List(context.Context, string) ([]*models.Category, error)
	GetByID(context.Context, uuid.UUID) (*models.Category, error)
	Create(context.Context, *models.Category) error
	Update(context.Context, *models.Category) error
	Delete(context.Context, uuid.UUID) error
	PermanentDelete(context.Context, uuid.UUID) error
	GetProductCount(context.Context, uuid.UUID) (int, error)
	GetProductCounts(context.Context, []uuid.UUID) (map[uuid.UUID]int, error)
	DuplicateSiblingExists(context.Context, string, *uuid.UUID) (bool, error)
}

type CategoryHandler struct {
	categoryRepo categoryHandlerStore
}

func NewCategoryHandler(categoryRepo categoryHandlerStore) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

type CategoryResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ProductCount *int    `json:"product_count,omitempty"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateCategoryRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
	ExpectedUpdatedAt *string `json:"expected_updated_at,omitempty"`
}
