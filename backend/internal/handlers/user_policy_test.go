package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
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
	permissionOverrides []*models.UserPermission
}

func (f *fakeUserRepo) GetByID(context.Context, uuid.UUID) (*models.User, error) {
	return f.getByIDUser, f.getByIDErr
}
func (f *fakeUserRepo) ListWithFilter(context.Context, int, int, *bool, string, string) ([]*models.User, int, error) {
	return nil, 0, nil
}
func (f *fakeUserRepo) Create(context.Context, *models.User) error               { return nil }
func (f *fakeUserRepo) Update(context.Context, *models.User) error               { return nil }
func (f *fakeUserRepo) UpdatePassword(context.Context, uuid.UUID, string) error  { return nil }
func (f *fakeUserRepo) UpdatePIN(context.Context, uuid.UUID, *string) error      { return nil }
func (f *fakeUserRepo) Deactivate(context.Context, uuid.UUID) error              { return nil }
func (f *fakeUserRepo) PermanentDelete(context.Context, uuid.UUID) error         { return nil }
func (f *fakeUserRepo) HasSalesHistory(context.Context, uuid.UUID) (bool, error) { return false, nil }
func (f *fakeUserRepo) EmailExists(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}
func (f *fakeUserRepo) GetUserPermissions(context.Context, uuid.UUID) ([]string, error) {
	return f.permissions, f.getPermissionsErr
}
func (f *fakeUserRepo) GetUserPermissionOverrides(context.Context, uuid.UUID) ([]*models.UserPermission, error) {
	return f.permissionOverrides, nil
}
func (f *fakeUserRepo) ClearUserPermissionOverrides(context.Context, uuid.UUID) error { return nil }
func (f *fakeUserRepo) SetUserPermission(context.Context, uuid.UUID, uuid.UUID, bool, *uuid.UUID) error {
	return nil
}

type fakeRoleRepo struct{}

func (f *fakeRoleRepo) GetByID(context.Context, uuid.UUID) (*models.Role, error) { return nil, nil }

type fakePermissionRepo struct {
	permission *models.Permission
}

func (f *fakePermissionRepo) GetByID(context.Context, uuid.UUID) (*models.Permission, error) {
	return f.permission, nil
}

func TestUserHandlerEnforceTargetUserActionRequiresManagerGrant(t *testing.T) {
	handler := &UserHandler{
		userRepo: &fakeUserRepo{permissions: []string{}},
	}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("role_name", "manager")
		c.Locals("user_id", uuid.New())
		return handler.enforceTargetUserAction(c, "manager", userActionEdit)
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

func TestUserHandlerSetPermissionsRejectsSelfModification(t *testing.T) {
	userID := uuid.New()
	handler := &UserHandler{
		userRepo:       &fakeUserRepo{getByIDUser: &models.User{ID: userID, Role: &models.Role{Name: "cashier"}}},
		roleRepo:       &fakeRoleRepo{},
		permissionRepo: &fakePermissionRepo{},
	}

	app := fiber.New()
	app.Patch("/users/:id/permissions", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role_name", "owner")
		return handler.SetPermissions(c)
	})

	req := httptest.NewRequest("PATCH", "/users/"+userID.String()+"/permissions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestUserHandlerSetPermissionsRejectsUnownedPermissionGrant(t *testing.T) {
	targetUserID := uuid.New()
	permissionID := uuid.New()
	handler := &UserHandler{
		userRepo: &fakeUserRepo{
			getByIDUser: &models.User{ID: targetUserID, Name: "Cashier", Role: &models.Role{Name: "cashier"}},
			permissions: []string{},
		},
		roleRepo: &fakeRoleRepo{},
		permissionRepo: &fakePermissionRepo{
			permission: &models.Permission{ID: permissionID, Key: "can_manage_users", Name: "Manage Users"},
		},
	}

	app := fiber.New()
	app.Patch("/users/:id/permissions", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role_name", "manager")
		return handler.SetPermissions(c)
	})

	req := httptest.NewRequest("PATCH", "/users/"+targetUserID.String()+"/permissions",
		strings.NewReader(`{"permissions":[{"permission_id":"`+permissionID.String()+`","allowed":true}]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}
