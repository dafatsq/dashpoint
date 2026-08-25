package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

var userPolicyUpdatedAt = time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)

const userPolicyUpdatedAtString = "2026-06-05T10:00:00Z"

type fakeUserRepo struct {
	getByIDUser          *models.User
	getByIDErr           error
	permissions          []string
	getPermissionsErr    error
	hasSalesHistory      bool
	hasExpenseHistory    bool
	updateCalled         bool
	updatePasswordCalled bool
	updatePINCalled      bool
	deactivateCalled     bool
}

func (f *fakeUserRepo) GetByID(context.Context, uuid.UUID) (*models.User, error) {
	return f.getByIDUser, f.getByIDErr
}
func (f *fakeUserRepo) ListWithFilter(context.Context, int, int, *bool, string, string, string, string) ([]*models.User, int, error) {
	return nil, 0, nil
}
func (f *fakeUserRepo) Create(context.Context, *models.User) error { return nil }
func (f *fakeUserRepo) Update(context.Context, *models.User) error {
	f.updateCalled = true
	return nil
}
func (f *fakeUserRepo) UpdatePassword(context.Context, uuid.UUID, string) error {
	f.updatePasswordCalled = true
	return nil
}
func (f *fakeUserRepo) UpdatePIN(context.Context, uuid.UUID, *string) error {
	f.updatePINCalled = true
	return nil
}
func (f *fakeUserRepo) Deactivate(context.Context, uuid.UUID) error {
	f.deactivateCalled = true
	return nil
}
func (f *fakeUserRepo) PermanentDelete(context.Context, uuid.UUID) error { return nil }
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

type fakeRefreshTokenRevoker struct {
	calls  int
	userID uuid.UUID
	reason string
}

func (f *fakeRefreshTokenRevoker) RevokeAllForUser(_ context.Context, userID uuid.UUID, reason string) error {
	f.calls++
	f.userID = userID
	f.reason = reason
	return nil
}

func TestUserHandlerEnforceTargetUserActionRejectsPeerManager(t *testing.T) {
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

func TestUserHandlerEnforceTargetUserActionAllowsCashierManagingCashier(t *testing.T) {
	handler := &UserHandler{
		userRepo: &fakeUserRepo{permissions: []string{}},
	}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("role_name", "cashier")
		c.Locals("user_id", uuid.New())
		if !handler.enforceTargetUserAction(c, "cashier", userActionEdit) {
			return nil
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}
}

func TestCanAssignRoleAllowsCashierManagersToAssignCashierOnly(t *testing.T) {
	if !canAssignRole("cashier", "cashier") {
		t.Fatalf("expected cashier actor to assign cashier role after manage-users route permission passes")
	}
	if canAssignRole("cashier", "manager") {
		t.Fatalf("expected cashier actor not to assign manager role")
	}
}

func TestUserHandlerPermanentDeleteRejectsExpenseHistory(t *testing.T) {
	targetUserID := uuid.New()
	handler := &UserHandler{
		userRepo: &fakeUserRepo{
			getByIDUser:       &models.User{ID: targetUserID, Name: "Cashier", UpdatedAt: userPolicyUpdatedAt, Role: &models.Role{Name: "cashier"}},
			permissions:       []string{"manage_users_page"},
			hasExpenseHistory: true,
		},
	}

	app := fiber.New()
	app.Delete("/users/:id/permanent", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role_name", "manager")
		return handler.PermanentDelete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+targetUserID.String()+"/permanent?expected_updated_at="+userPolicyUpdatedAtString, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
}

func TestUserHandlerUpdateRejectsArchivedUserWithoutRestore(t *testing.T) {
	targetUserID := uuid.New()
	repo := &fakeUserRepo{
		getByIDUser: &models.User{ID: targetUserID, Name: "Cashier", IsActive: false, UpdatedAt: userPolicyUpdatedAt, Role: &models.Role{Name: "cashier"}},
		permissions: []string{"manage_users_page"},
	}
	handler := &UserHandler{userRepo: repo}

	app := fiber.New()
	app.Patch("/users/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role_name", "owner")
		return handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/"+targetUserID.String(), strings.NewReader(`{"name":"Renamed","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
	if repo.updateCalled {
		t.Fatalf("expected archived user update to be blocked")
	}
}

func TestUserHandlerUpdateAllowsSelfProfileWithoutManageUsers(t *testing.T) {
	targetUserID := uuid.New()
	roleID := uuid.New()
	repo := &fakeUserRepo{
		getByIDUser: &models.User{
			ID:        targetUserID,
			Name:      "Cashier",
			RoleID:    roleID,
			IsActive:  true,
			UpdatedAt: userPolicyUpdatedAt,
			Role:      &models.Role{ID: roleID, Name: "cashier"},
		},
		permissions: []string{},
	}
	handler := &UserHandler{userRepo: repo}

	app := fiber.New()
	app.Patch("/users/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", targetUserID)
		c.Locals("role_name", "cashier")
		return handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/"+targetUserID.String(), strings.NewReader(`{"name":"Cashier One","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if !repo.updateCalled {
		t.Fatalf("expected self profile update to be allowed")
	}
}

func TestUserHandlerUpdateRevokesRefreshTokensForCredentialChange(t *testing.T) {
	targetUserID := uuid.New()
	roleID := uuid.New()
	email := "cashier@example.com"
	repo := &fakeUserRepo{
		getByIDUser: &models.User{
			ID:        targetUserID,
			Email:     &email,
			Name:      "Cashier",
			RoleID:    roleID,
			IsActive:  true,
			UpdatedAt: userPolicyUpdatedAt,
			Role:      &models.Role{ID: roleID, Name: "cashier"},
		},
		permissions: []string{},
	}
	revoker := &fakeRefreshTokenRevoker{}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: revoker}

	app := fiber.New()
	app.Patch("/users/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", targetUserID)
		c.Locals("role_name", "cashier")
		return handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/"+targetUserID.String(), strings.NewReader(`{"email":"renamed@example.com","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if revoker.calls != 1 {
		t.Fatalf("expected one refresh token revocation, got %d", revoker.calls)
	}
	if revoker.userID != targetUserID || revoker.reason != "user_credentials_changed" {
		t.Fatalf("unexpected revocation target/reason: %s %q", revoker.userID, revoker.reason)
	}
}

func TestUserHandlerUpdateRequiresExpectedUpdatedAt(t *testing.T) {
	targetUserID := uuid.New()
	roleID := uuid.New()
	repo := &fakeUserRepo{
		getByIDUser: &models.User{
			ID:        targetUserID,
			Name:      "Cashier",
			RoleID:    roleID,
			IsActive:  true,
			UpdatedAt: userPolicyUpdatedAt,
			Role:      &models.Role{ID: roleID, Name: "cashier"},
		},
	}
	handler := &UserHandler{userRepo: repo}

	app := fiber.New()
	app.Patch("/users/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", targetUserID)
		c.Locals("role_name", "cashier")
		return handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/"+targetUserID.String(), strings.NewReader(`{"name":"Cashier One"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	if repo.updateCalled {
		t.Fatalf("expected missing expected_updated_at to be blocked")
	}
}

func TestUserHandlerCreateRejectsUnknownFields(t *testing.T) {
	handler := &UserHandler{userRepo: &fakeUserRepo{}, roleRepo: &fakeRoleRepo{}}

	app := fiber.New()
	app.Post("/users", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role_name", "owner")
		return handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Cashier","email":"cashier@example.com","password":"new-secret","pin":"1234","role_id":"11111111-1111-1111-1111-111111111111","is_admin":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserHandlerUpdateRejectsSelfStatusChange(t *testing.T) {
	targetUserID := uuid.New()
	roleID := uuid.New()
	repo := &fakeUserRepo{
		getByIDUser: &models.User{
			ID:        targetUserID,
			Name:      "Cashier",
			RoleID:    roleID,
			IsActive:  true,
			UpdatedAt: userPolicyUpdatedAt,
			Role:      &models.Role{ID: roleID, Name: "cashier"},
		},
		permissions: []string{},
	}
	handler := &UserHandler{userRepo: repo}

	app := fiber.New()
	app.Patch("/users/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", targetUserID)
		c.Locals("role_name", "cashier")
		return handler.Update(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/"+targetUserID.String(), strings.NewReader(`{"is_active":false,"expected_updated_at":"`+userPolicyUpdatedAtString+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
	if repo.updateCalled {
		t.Fatalf("expected self status change to be blocked")
	}
}

func TestUserHandlerUpdatePasswordRejectsArchivedUser(t *testing.T) {
	targetUserID := uuid.New()
	repo := &fakeUserRepo{
		getByIDUser: &models.User{ID: targetUserID, Name: "Cashier", IsActive: false, UpdatedAt: userPolicyUpdatedAt, Role: &models.Role{Name: "cashier"}},
		permissions: []string{"manage_users_page"},
	}
	handler := &UserHandler{userRepo: repo}

	app := fiber.New()
	app.Patch("/users/:id/password", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role_name", "owner")
		return handler.UpdatePassword(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/"+targetUserID.String()+"/password", strings.NewReader(`{"password":"new-secret","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
	if repo.updatePasswordCalled {
		t.Fatalf("expected archived user password update to be blocked")
	}
}

func TestUserHandlerUpdatePasswordRevokesRefreshTokens(t *testing.T) {
	targetUserID := uuid.New()
	roleID := uuid.New()
	repo := &fakeUserRepo{
		getByIDUser: &models.User{
			ID:           targetUserID,
			Name:         "Cashier",
			RoleID:       roleID,
			IsActive:     true,
			UpdatedAt:    userPolicyUpdatedAt,
			Role:         &models.Role{ID: roleID, Name: "cashier"},
			PasswordHash: &selfProofPasswordHash,
		},
		permissions: []string{},
	}
	revoker := &fakeRefreshTokenRevoker{}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: revoker}

	app := fiber.New()
	app.Patch("/users/:id/password", func(c *fiber.Ctx) error {
		c.Locals("user_id", targetUserID)
		c.Locals("role_name", "cashier")
		return handler.UpdatePassword(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/"+targetUserID.String()+"/password", strings.NewReader(`{"password":"new-secret","current_password":"current-password","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if revoker.calls != 1 {
		t.Fatalf("expected one refresh token revocation, got %d", revoker.calls)
	}
	if revoker.userID != targetUserID || revoker.reason != "user_password_changed" {
		t.Fatalf("unexpected revocation target/reason: %s %q", revoker.userID, revoker.reason)
	}
}

func TestUserHandlerUpdatePINAllowsSelfWithoutManageUsers(t *testing.T) {
	targetUserID := uuid.New()
	roleID := uuid.New()
	repo := &fakeUserRepo{
		getByIDUser: &models.User{
			ID:           targetUserID,
			Name:         "Cashier",
			RoleID:       roleID,
			IsActive:     true,
			UpdatedAt:    userPolicyUpdatedAt,
			Role:         &models.Role{ID: roleID, Name: "cashier"},
			PasswordHash: &selfProofPasswordHash,
		},
		permissions: []string{},
	}
	handler := &UserHandler{userRepo: repo}

	app := fiber.New()
	app.Patch("/users/:id/pin", func(c *fiber.Ctx) error {
		c.Locals("user_id", targetUserID)
		c.Locals("role_name", "cashier")
		return handler.UpdatePIN(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/"+targetUserID.String()+"/pin", strings.NewReader(`{"pin":"1234","current_password":"current-password","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if !repo.updatePINCalled {
		t.Fatalf("expected self PIN update to be allowed")
	}
}

func TestUserHandlerUpdatePINRevokesRefreshTokens(t *testing.T) {
	targetUserID := uuid.New()
	roleID := uuid.New()
	repo := &fakeUserRepo{
		getByIDUser: &models.User{
			ID:           targetUserID,
			Name:         "Cashier",
			RoleID:       roleID,
			IsActive:     true,
			UpdatedAt:    userPolicyUpdatedAt,
			Role:         &models.Role{ID: roleID, Name: "cashier"},
			PasswordHash: &selfProofPasswordHash,
		},
		permissions: []string{},
	}
	revoker := &fakeRefreshTokenRevoker{}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: revoker}

	app := fiber.New()
	app.Patch("/users/:id/pin", func(c *fiber.Ctx) error {
		c.Locals("user_id", targetUserID)
		c.Locals("role_name", "cashier")
		return handler.UpdatePIN(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/users/"+targetUserID.String()+"/pin", strings.NewReader(`{"pin":"1234","current_password":"current-password","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if revoker.calls != 1 {
		t.Fatalf("expected one refresh token revocation, got %d", revoker.calls)
	}
	if revoker.userID != targetUserID || revoker.reason != "user_pin_changed" {
		t.Fatalf("unexpected revocation target/reason: %s %q", revoker.userID, revoker.reason)
	}
}

func TestUserHandlerDeleteRejectsAlreadyArchivedUser(t *testing.T) {
	targetUserID := uuid.New()
	repo := &fakeUserRepo{
		getByIDUser: &models.User{ID: targetUserID, Name: "Cashier", IsActive: false, UpdatedAt: userPolicyUpdatedAt, Role: &models.Role{Name: "cashier"}},
		permissions: []string{"manage_users_page"},
	}
	handler := &UserHandler{userRepo: repo}

	app := fiber.New()
	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role_name", "owner")
		return handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+targetUserID.String()+"?expected_updated_at="+userPolicyUpdatedAtString, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
	if repo.deactivateCalled {
		t.Fatalf("expected archived user delete to be blocked")
	}
}
