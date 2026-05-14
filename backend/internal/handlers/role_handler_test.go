package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

type fakeRoleReader struct {
	roles       []*models.Role
	roleByID    *models.Role
	permissions []*models.Permission
	listErr     error
	getByIDErr  error
	permErr     error
}

func (f *fakeRoleReader) List(context.Context) ([]*models.Role, error) {
	return f.roles, f.listErr
}

func (f *fakeRoleReader) GetByID(context.Context, uuid.UUID) (*models.Role, error) {
	return f.roleByID, f.getByIDErr
}

func (f *fakeRoleReader) GetRolePermissions(context.Context, uuid.UUID) ([]*models.Permission, error) {
	return f.permissions, f.permErr
}

type fakePermissionReader struct {
	permissions       []*models.Permission
	permissionsByCate map[string][]*models.Permission
	listErr           error
	listByCatErr      error
}

func (f *fakePermissionReader) List(context.Context) ([]*models.Permission, error) {
	return f.permissions, f.listErr
}

func (f *fakePermissionReader) ListByCategory(context.Context) (map[string][]*models.Permission, error) {
	return f.permissionsByCate, f.listByCatErr
}

func TestRoleHandlerListRolesSuccess(t *testing.T) {
	handler := &RoleHandler{
		roleRepo: &fakeRoleReader{
			roles: []*models.Role{
				{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Name: "owner"},
				{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Name: "cashier"},
			},
		},
	}

	app := fiber.New()
	app.Get("/roles", handler.ListRoles)

	resp, err := app.Test(httptest.NewRequest("GET", "/roles", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body struct {
		Roles []RoleResponse `json:"roles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if len(body.Roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(body.Roles))
	}
}

func TestRoleHandlerGetRoleRejectsInvalidID(t *testing.T) {
	handler := &RoleHandler{}
	app := fiber.New()
	app.Get("/roles/:id", handler.GetRole)

	resp, err := app.Test(httptest.NewRequest("GET", "/roles/not-a-uuid", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestRoleHandlerGetRoleNotFound(t *testing.T) {
	handler := &RoleHandler{
		roleRepo: &fakeRoleReader{},
	}
	app := fiber.New()
	app.Get("/roles/:id", handler.GetRole)

	resp, err := app.Test(httptest.NewRequest("GET", "/roles/11111111-1111-1111-1111-111111111111", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestRoleHandlerListPermissionsGroupedAndUngrouped(t *testing.T) {
	handler := &RoleHandler{
		permissionRepo: &fakePermissionReader{
			permissions: []*models.Permission{
				{ID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Key: "can_view_users", Name: "View Users", Category: "users"},
			},
			permissionsByCate: map[string][]*models.Permission{
				"users": {
					{ID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), Key: "can_view_users", Name: "View Users", Category: "users"},
				},
			},
		},
	}

	app := fiber.New()
	app.Get("/permissions", handler.ListPermissions)

	groupedResp, err := app.Test(httptest.NewRequest("GET", "/permissions?grouped=true", nil))
	if err != nil {
		t.Fatalf("grouped app.Test returned error: %v", err)
	}
	if groupedResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected grouped status 200, got %d", groupedResp.StatusCode)
	}

	var groupedBody struct {
		Permissions map[string][]PermissionResponse `json:"permissions"`
	}
	if err := json.NewDecoder(groupedResp.Body).Decode(&groupedBody); err != nil {
		t.Fatalf("grouped decode returned error: %v", err)
	}
	if len(groupedBody.Permissions["users"]) != 1 {
		t.Fatalf("expected one grouped permission")
	}

	plainResp, err := app.Test(httptest.NewRequest("GET", "/permissions", nil))
	if err != nil {
		t.Fatalf("plain app.Test returned error: %v", err)
	}
	if plainResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected plain status 200, got %d", plainResp.StatusCode)
	}

	var plainBody struct {
		Permissions []PermissionResponse `json:"permissions"`
	}
	if err := json.NewDecoder(plainResp.Body).Decode(&plainBody); err != nil {
		t.Fatalf("plain decode returned error: %v", err)
	}
	if len(plainBody.Permissions) != 1 {
		t.Fatalf("expected one permission, got %d", len(plainBody.Permissions))
	}
}

func TestRoleHandlerReturnsInternalErrorOnRoleListFailure(t *testing.T) {
	handler := &RoleHandler{
		roleRepo: &fakeRoleReader{listErr: errors.New("db down")},
	}
	app := fiber.New()
	app.Get("/roles", handler.ListRoles)

	resp, err := app.Test(httptest.NewRequest("GET", "/roles", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.StatusCode)
	}
}
