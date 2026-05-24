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
	roles      []*models.Role
	roleByID   *models.Role
	listErr    error
	getByIDErr error
}

func (f *fakeRoleReader) List(context.Context) ([]*models.Role, error) {
	return f.roles, f.listErr
}

func (f *fakeRoleReader) GetByID(context.Context, uuid.UUID) (*models.Role, error) {
	return f.roleByID, f.getByIDErr
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

func TestRoleHandlerGetRoleReturnsDerivedPermissions(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	handler := &RoleHandler{
		roleRepo: &fakeRoleReader{
			roleByID: &models.Role{
				ID:   roleID,
				Name: "manager",
			},
		},
	}

	app := fiber.New()
	app.Get("/roles/:id", handler.GetRole)

	resp, err := app.Test(httptest.NewRequest("GET", "/roles/"+roleID.String(), nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body RoleDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode returned error: %v", err)
	}
	if len(body.Permissions) == 0 {
		t.Fatal("expected derived permissions")
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
