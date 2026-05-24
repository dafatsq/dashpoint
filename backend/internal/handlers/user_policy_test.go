package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

type fakeUserRepo struct {
	getByIDUser         *models.User
	getByIDErr          error
	permissions         []string
	getPermissionsErr   error
	hasSalesHistory     bool
	hasExpenseHistory   bool
}

func (f *fakeUserRepo) GetByID(context.Context, uuid.UUID) (*models.User, error) {
	return f.getByIDUser, f.getByIDErr
}
func (f *fakeUserRepo) ListWithFilter(context.Context, int, int, *bool, string, string) ([]*models.User, int, error) {
	return nil, 0, nil
}
func (f *fakeUserRepo) Create(context.Context, *models.User) error              { return nil }
func (f *fakeUserRepo) Update(context.Context, *models.User) error              { return nil }
func (f *fakeUserRepo) UpdatePassword(context.Context, uuid.UUID, string) error { return nil }
func (f *fakeUserRepo) UpdatePIN(context.Context, uuid.UUID, *string) error     { return nil }
func (f *fakeUserRepo) Deactivate(context.Context, uuid.UUID) error             { return nil }
func (f *fakeUserRepo) PermanentDelete(context.Context, uuid.UUID) error        { return nil }
func (f *fakeUserRepo) HasSalesHistory(context.Context, uuid.UUID) (bool, error) {
	return f.hasSalesHistory, nil
}
func (f *fakeUserRepo) HasExpenseHistory(context.Context, uuid.UUID) (bool, error) {
	return f.hasExpenseHistory, nil
}
func (f *fakeUserRepo) EmailExists(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}
func (f *fakeUserRepo) NameExists(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}
type fakeRoleRepo struct{}

func (f *fakeRoleRepo) GetByID(context.Context, uuid.UUID) (*models.Role, error) { return nil, nil }

func TestUserHandlerEnforceTargetUserActionRequiresManagerGrant(t *testing.T) {
	handler := &UserHandler{
		userRepo: &fakeUserRepo{permissions: []string{}},
	}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("role_name", "manager")
		c.Locals("user_id", uuid.New())
		if !handler.enforceTargetUserAction(c, "manager", userActionEdit) {
			return nil
		}
		return nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestUserHandlerPermanentDeleteRejectsExpenseHistory(t *testing.T) {
	targetUserID := uuid.New()
	handler := &UserHandler{
		userRepo: &fakeUserRepo{
			getByIDUser:       &models.User{ID: targetUserID, Name: "Cashier", Role: &models.Role{Name: "cashier"}},
			permissions:       []string{"can_delete_cashier_users"},
			hasExpenseHistory: true,
		},
	}

	app := fiber.New()
	app.Delete("/users/:id/permanent", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role_name", "manager")
		return handler.PermanentDelete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+targetUserID.String()+"/permanent", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
}
