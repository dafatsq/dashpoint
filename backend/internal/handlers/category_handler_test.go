package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

type fakeManagedCategoryStore struct {
	listFunc             func(context.Context, string) ([]*models.Category, error)
	getByIDFunc          func(context.Context, uuid.UUID) (*models.Category, error)
	createFunc           func(context.Context, *models.Category) error
	updateFunc           func(context.Context, *models.Category) error
	deleteFunc           func(context.Context, uuid.UUID) error
	permanentDeleteFunc  func(context.Context, uuid.UUID) error
	getProductCountFunc  func(context.Context, uuid.UUID) (int, error)
	getProductCountsFunc func(context.Context, []uuid.UUID) (map[uuid.UUID]int, error)
}

func (f *fakeManagedCategoryStore) List(ctx context.Context, status string) ([]*models.Category, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, status)
	}
	return nil, nil
}
func (f *fakeManagedCategoryStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return nil, nil
}
func (f *fakeManagedCategoryStore) Create(ctx context.Context, category *models.Category) error {
	if f.createFunc != nil {
		return f.createFunc(ctx, category)
	}
	return nil
}
func (f *fakeManagedCategoryStore) Update(ctx context.Context, category *models.Category) error {
	if f.updateFunc != nil {
		return f.updateFunc(ctx, category)
	}
	return nil
}
func (f *fakeManagedCategoryStore) Delete(ctx context.Context, id uuid.UUID) error {
	if f.deleteFunc != nil {
		return f.deleteFunc(ctx, id)
	}
	return nil
}

func (f *fakeManagedCategoryStore) DuplicateSiblingExists(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func (f *fakeManagedCategoryStore) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	if f.permanentDeleteFunc != nil {
		return f.permanentDeleteFunc(ctx, id)
	}
	return nil
}
func (f *fakeManagedCategoryStore) GetProductCount(ctx context.Context, id uuid.UUID) (int, error) {
	if f.getProductCountFunc != nil {
		return f.getProductCountFunc(ctx, id)
	}
	return 0, nil
}
func (f *fakeManagedCategoryStore) GetProductCounts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error) {
	if f.getProductCountsFunc != nil {
		return f.getProductCountsFunc(ctx, ids)
	}
	return map[uuid.UUID]int{}, nil
}

func TestCategoryHandlerListUsesBatchedProductCounts(t *testing.T) {
	categoryID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := &fakeManagedCategoryStore{
		listFunc: func(context.Context, string) ([]*models.Category, error) {
			return []*models.Category{{
				ID:        categoryID,
				Name:      "Coffee",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}}, nil
		},
		getProductCountsFunc: func(context.Context, []uuid.UUID) (map[uuid.UUID]int, error) {
			return map[uuid.UUID]int{categoryID: 9}, nil
		},
	}

	handler := NewCategoryHandler(repo)
	app := fiber.New()
	app.Get("/categories", handler.List)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/categories?status=active", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body struct {
		Categories []CategoryResponse `json:"categories"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode returned error: %v", err)
	}
	if len(body.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(body.Categories))
	}
	if body.Categories[0].ProductCount == nil || *body.Categories[0].ProductCount != 9 {
		t.Fatalf("expected product count 9, got %+v", body.Categories[0].ProductCount)
	}
}


func TestCategoryHandlerGetRejectsInvalidID(t *testing.T) {
	handler := NewCategoryHandler(&fakeManagedCategoryStore{})
	app := fiber.New()
	app.Get("/categories/:id", handler.Get)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/categories/not-a-uuid", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestCategoryHandlerGetReturnsNotFound(t *testing.T) {
	handler := NewCategoryHandler(&fakeManagedCategoryStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Category, error) {
			return nil, errors.New("category not found")
		},
	})
	app := fiber.New()
	app.Get("/categories/:id", handler.Get)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/categories/11111111-1111-1111-1111-111111111111", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestCategoryHandlerCreateIncludesSuccessMessage(t *testing.T) {
	handler := NewCategoryHandler(&fakeManagedCategoryStore{
		createFunc: func(_ context.Context, category *models.Category) error {
			category.ID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
			category.CreatedAt = time.Now()
			category.UpdatedAt = category.CreatedAt
			return nil
		},
	})
	app := fiber.New()
	app.Post("/categories", handler.Create)

	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"Coffee"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode returned error: %v", err)
	}
	if body.Message != "Category created successfully" {
		t.Fatalf("unexpected message %q", body.Message)
	}
}

func TestCategoryHandlerListReturnsInternalErrorOnFailure(t *testing.T) {
	handler := NewCategoryHandler(&fakeManagedCategoryStore{
		listFunc: func(context.Context, string) ([]*models.Category, error) {
			return nil, errors.New("db down")
		},
	})
	app := fiber.New()
	app.Get("/categories", handler.List)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/categories", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.StatusCode)
	}
}
