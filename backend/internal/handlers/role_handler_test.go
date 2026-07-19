package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/models"
)

type fakeRoleReader struct {
	roles       []*models.Role
	roleByID    *models.Role
	updatedID   uuid.UUID
	updatedKeys []string
	userIDs     []uuid.UUID
	listErr     error
	getByIDErr  error
	updateErr   error
}

func (f *fakeRoleReader) List(context.Context) ([]*models.Role, error) {
	return f.roles, f.listErr
}

func (f *fakeRoleReader) GetByID(context.Context, uuid.UUID) (*models.Role, error) {
	return f.roleByID, f.getByIDErr
}

func (f *fakeRoleReader) ListActiveUserIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return f.userIDs, nil
}

func (f *fakeRoleReader) UpdatePermissions(_ context.Context, id uuid.UUID, permissionKeys []string) error {
	f.updatedID = id
	f.updatedKeys = permissionKeys
	return f.updateErr
}

type fakeRoleEventBroadcaster struct {
	events []UserEvent
}

func (f *fakeRoleEventBroadcaster) BroadcastToUser(_ uuid.UUID, event UserEvent) {
	f.events = append(f.events, event)
}

func (f *fakeRoleEventBroadcaster) DisconnectUser(uuid.UUID) {}

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

func TestRoleHandlerGetRoleReturnsStoredPermissions(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	handler := &RoleHandler{
		roleRepo: &fakeRoleReader{
			roleByID: &models.Role{
				ID:          roleID,
				Name:        "manager",
				Permissions: []string{"access_users_page", "manage_users_page"},
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
	if len(body.Permissions) != 2 {
		t.Fatalf("expected stored permissions, got %v", body.Permissions)
	}
}

func TestRoleHandlerUpdatePermissionsAddsAccessParent(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo := &fakeRoleReader{
		roleByID: &models.Role{
			ID:   roleID,
			Name: "manager",
		},
	}
	handler := &RoleHandler{roleRepo: repo}
	app := fiber.New()
	app.Patch("/roles/:id/permissions", handler.UpdateRolePermissions)

	req := httptest.NewRequest(
		"PATCH",
		"/roles/"+roleID.String()+"/permissions",
		bytes.NewBufferString(`{"permissions":["manage_users_page"],"expected_permissions":[]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if repo.updatedID != roleID {
		t.Fatalf("expected role %s to be updated, got %s", roleID, repo.updatedID)
	}
	if len(repo.updatedKeys) != 2 {
		t.Fatalf("expected manage permission and access parent, got %v", repo.updatedKeys)
	}
}

func TestRoleHandlerUpdatePermissionsRejectsInvalidPermission(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	handler := &RoleHandler{
		roleRepo: &fakeRoleReader{
			roleByID: &models.Role{ID: roleID, Name: "manager"},
		},
	}
	app := fiber.New()
	app.Patch("/roles/:id/permissions", handler.UpdateRolePermissions)

	req := httptest.NewRequest(
		"PATCH",
		"/roles/"+roleID.String()+"/permissions",
		bytes.NewBufferString(`{"permissions":["can_delete_everything"],"expected_permissions":[]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestRoleHandlerUpdatePermissionsRejectsStaleExpectedPermissions(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo := &fakeRoleReader{
		roleByID: &models.Role{
			ID:          roleID,
			Name:        "manager",
			Permissions: []string{"access_users_page", "manage_users_page"},
		},
	}
	handler := &RoleHandler{roleRepo: repo}
	app := fiber.New()
	app.Patch("/roles/:id/permissions", handler.UpdateRolePermissions)

	req := httptest.NewRequest(
		"PATCH",
		"/roles/"+roleID.String()+"/permissions",
		bytes.NewBufferString(`{"permissions":["access_users_page"],"expected_permissions":["access_users_page"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
	if repo.updatedID != uuid.Nil {
		t.Fatalf("expected stale permissions submit to be blocked")
	}
}

func TestRoleHandlerUpdatePermissionsAcceptsExpectedPermissionsInDifferentOrder(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo := &fakeRoleReader{
		roleByID: &models.Role{
			ID:          roleID,
			Name:        "manager",
			Permissions: []string{"manage_users_page", "access_users_page"},
		},
	}
	handler := &RoleHandler{roleRepo: repo}
	app := fiber.New()
	app.Patch("/roles/:id/permissions", handler.UpdateRolePermissions)

	req := httptest.NewRequest(
		"PATCH",
		"/roles/"+roleID.String()+"/permissions",
		bytes.NewBufferString(`{"permissions":["access_users_page"],"expected_permissions":["access_users_page","manage_users_page"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if repo.updatedID != roleID {
		t.Fatalf("expected role permissions update to continue")
	}
}

func TestRoleHandlerUpdatePermissionsDropsDeprecatedSettingsPermission(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo := &fakeRoleReader{
		roleByID: &models.Role{ID: roleID, Name: "manager"},
	}
	handler := &RoleHandler{roleRepo: repo}
	app := fiber.New()
	app.Patch("/roles/:id/permissions", handler.UpdateRolePermissions)

	req := httptest.NewRequest(
		"PATCH",
		"/roles/"+roleID.String()+"/permissions",
		bytes.NewBufferString(`{"permissions":["access_users_page","access_settings_page"],"expected_permissions":[]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if len(repo.updatedKeys) != 1 || repo.updatedKeys[0] != "access_users_page" {
		t.Fatalf("expected deprecated settings permission to be dropped, got %v", repo.updatedKeys)
	}
}

func TestRoleHandlerUpdatePermissionsLocksOwner(t *testing.T) {
	roleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	handler := &RoleHandler{
		roleRepo: &fakeRoleReader{
			roleByID: &models.Role{ID: roleID, Name: "owner"},
		},
	}
	app := fiber.New()
	app.Patch("/roles/:id/permissions", handler.UpdateRolePermissions)

	req := httptest.NewRequest(
		"PATCH",
		"/roles/"+roleID.String()+"/permissions",
		bytes.NewBufferString(`{"permissions":["access_users_page"],"expected_permissions":[]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestRoleHandlerUpdatePermissionsRequiresExpectedPermissions(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	handler := &RoleHandler{
		roleRepo: &fakeRoleReader{
			roleByID: &models.Role{ID: roleID, Name: "manager"},
		},
	}
	app := fiber.New()
	app.Patch("/roles/:id/permissions", handler.UpdateRolePermissions)

	req := httptest.NewRequest(
		"PATCH",
		"/roles/"+roleID.String()+"/permissions",
		bytes.NewBufferString(`{"permissions":["access_users_page"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestRoleHandlerUpdatePermissionsRejectsUnknownFields(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	handler := &RoleHandler{
		roleRepo: &fakeRoleReader{
			roleByID: &models.Role{ID: roleID, Name: "manager"},
		},
	}
	app := fiber.New()
	app.Patch("/roles/:id/permissions", handler.UpdateRolePermissions)

	req := httptest.NewRequest(
		"PATCH",
		"/roles/"+roleID.String()+"/permissions",
		bytes.NewBufferString(`{"permissions":["access_users_page"],"expected_permissions":[],"is_owner":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestRoleHandlerUpdatePermissionsAuditsAndBroadcastsRoleUsers(t *testing.T) {
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	auditRepo := &stubAuditRepository{}
	audit.Init(auditRepo)
	events := &fakeRoleEventBroadcaster{}
	repo := &fakeRoleReader{
		roleByID: &models.Role{
			ID:          roleID,
			Name:        "manager",
			Permissions: []string{"access_users_page"},
		},
		userIDs: []uuid.UUID{userID},
	}
	handler := &RoleHandler{roleRepo: repo}
	handler.SetEventsHandler(events)
	app := fiber.New()
	app.Patch("/roles/:id/permissions", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.MustParse("44444444-4444-4444-4444-444444444444"))
		return handler.UpdateRolePermissions(c)
	})

	req := httptest.NewRequest(
		"PATCH",
		"/roles/"+roleID.String()+"/permissions",
		bytes.NewBufferString(`{"permissions":["access_users_page","manage_users_page"],"expected_permissions":["access_users_page"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if auditRepo.last == nil || auditRepo.last.Action != models.AuditActionPermissionChange {
		t.Fatalf("expected permission change audit entry, got %#v", auditRepo.last)
	}
	if auditRepo.last.EntityType != models.AuditEntityRole {
		t.Fatalf("expected role audit entity, got %q", auditRepo.last.EntityType)
	}
	if len(events.events) != 1 || events.events[0].Type != EventPermissionsChanged {
		t.Fatalf("expected permissions_changed event, got %#v", events.events)
	}
	if events.events[0].UserID != userID.String() {
		t.Fatalf("expected event for role user %s, got %s", userID, events.events[0].UserID)
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
