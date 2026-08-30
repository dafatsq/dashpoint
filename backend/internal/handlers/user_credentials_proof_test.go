package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/models"
)

// Computed once per test binary: bcrypt cost 12 is slow enough that hashing
// inside every test would dominate the suite runtime.
var (
	selfProofPasswordHash, _ = auth.HashPassword("current-password")
	selfProofPINHash, _      = auth.HashPIN("1234")
)

func newSelfProofUser(id uuid.UUID, pinHash *string) *models.User {
	return &models.User{
		ID:           id,
		Name:         "Self User",
		PasswordHash: &selfProofPasswordHash,
		PINHash:      pinHash,
		IsActive:     true,
		UpdatedAt:    userPolicyUpdatedAt,
		Role:         &models.Role{Name: "cashier"},
	}
}

func runSelfProofRequest(t *testing.T, endpoint func(c *fiber.Ctx) error, method, target string, userID uuid.UUID, body string) *http.Response {
	t.Helper()

	app := fiber.New()
	app.Add(method, target, func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role_name", "cashier")
		return c.Next()
	}, func(c *fiber.Ctx) error {
		return endpoint(c)
	})

	url := strings.ReplaceAll(target, ":id", userID.String())
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	return resp
}

func TestUpdatePasswordSelfWithoutProofRejected(t *testing.T) {
	userID := uuid.New()
	handler := &UserHandler{
		userRepo:         &fakeUserRepo{getByIDUser: newSelfProofUser(userID, &selfProofPINHash)},
		refreshTokenRepo: &fakeRefreshTokenRevoker{},
	}

	resp := runSelfProofRequest(t, handler.UpdatePassword, fiber.MethodPatch, "/users/:id/password", userID,
		`{"password":"brand-new-password","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdatePasswordSelfWithWrongPasswordRejected(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{getByIDUser: newSelfProofUser(userID, &selfProofPINHash)}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: &fakeRefreshTokenRevoker{}}

	resp := runSelfProofRequest(t, handler.UpdatePassword, fiber.MethodPatch, "/users/:id/password", userID,
		`{"password":"brand-new-password","current_password":"wrong-password","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`)

	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
	if repo.updatePasswordCalled {
		t.Fatalf("password must not be updated when proof fails")
	}
}

func TestUpdatePasswordSelfWithCurrentPasswordSucceeds(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{getByIDUser: newSelfProofUser(userID, &selfProofPINHash)}
	revoker := &fakeRefreshTokenRevoker{}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: revoker}

	resp := runSelfProofRequest(t, handler.UpdatePassword, fiber.MethodPatch, "/users/:id/password", userID,
		`{"password":"brand-new-password","current_password":"current-password","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if !repo.updatePasswordCalled {
		t.Fatalf("expected password update to be called")
	}
	if revoker.calls != 1 {
		t.Fatalf("expected refresh tokens to be revoked once, got %d", revoker.calls)
	}
}

func TestUpdatePasswordSelfAcceptsCurrentPINAsProof(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{getByIDUser: newSelfProofUser(userID, &selfProofPINHash)}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: &fakeRefreshTokenRevoker{}}

	resp := runSelfProofRequest(t, handler.UpdatePassword, fiber.MethodPatch, "/users/:id/password", userID,
		`{"password":"brand-new-password","current_pin":"1234","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if !repo.updatePasswordCalled {
		t.Fatalf("expected password update to be called")
	}
}

func TestUpdatePINSelfRequiresProof(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{getByIDUser: newSelfProofUser(userID, &selfProofPINHash)}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: &fakeRefreshTokenRevoker{}}

	resp := runSelfProofRequest(t, handler.UpdatePIN, fiber.MethodPatch, "/users/:id/pin", userID,
		`{"pin":"5678","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	if repo.updatePINCalled {
		t.Fatalf("PIN must not be updated when proof is missing")
	}
}

func TestUpdatePINSelfWithWrongPINRejected(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{getByIDUser: newSelfProofUser(userID, &selfProofPINHash)}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: &fakeRefreshTokenRevoker{}}

	resp := runSelfProofRequest(t, handler.UpdatePIN, fiber.MethodPatch, "/users/:id/pin", userID,
		`{"pin":"5678","current_pin":"9999","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`)

	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestUpdatePINSelfWithCurrentPINSucceeds(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{getByIDUser: newSelfProofUser(userID, &selfProofPINHash)}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: &fakeRefreshTokenRevoker{}}

	resp := runSelfProofRequest(t, handler.UpdatePIN, fiber.MethodPatch, "/users/:id/pin", userID,
		`{"pin":"5678","current_pin":"1234","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if !repo.updatePINCalled {
		t.Fatalf("expected PIN update to be called")
	}
}

func TestUpdatePINSelfWithoutExistingPINUsesPasswordProof(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{getByIDUser: newSelfProofUser(userID, nil)}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: &fakeRefreshTokenRevoker{}}

	resp := runSelfProofRequest(t, handler.UpdatePIN, fiber.MethodPatch, "/users/:id/pin", userID,
		`{"pin":"5678","current_password":"current-password","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if !repo.updatePINCalled {
		t.Fatalf("expected PIN update to be called")
	}
}

func TestUpdatePasswordAdminResetNeedsNoProof(t *testing.T) {
	targetID := uuid.New()
	actorID := uuid.New()
	target := newSelfProofUser(targetID, &selfProofPINHash)
	target.Role = &models.Role{Name: "cashier"}
	handler := &UserHandler{
		userRepo:         &fakeUserRepo{getByIDUser: target, permissions: []string{"manage_users_page"}},
		refreshTokenRepo: &fakeRefreshTokenRevoker{},
	}

	app := fiber.New()
	app.Patch("/users/:id/password", func(c *fiber.Ctx) error {
		c.Locals("user_id", actorID)
		c.Locals("role_name", "manager")
		return c.Next()
	}, func(c *fiber.Ctx) error {
		return handler.UpdatePassword(c)
	})

	req := httptest.NewRequest(fiber.MethodPatch, "/users/"+targetID.String()+"/password",
		strings.NewReader(`{"password":"brand-new-password","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUpdateSelfCredentialChangeThroughUserUpdateRequiresProof(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{getByIDUser: newSelfProofUser(userID, &selfProofPINHash)}
	handler := &UserHandler{userRepo: repo, refreshTokenRepo: &fakeRefreshTokenRevoker{}}

	resp := runSelfProofRequest(t, handler.Update, fiber.MethodPatch, "/users/:id", userID,
		`{"name":"Self User","password":"brand-new-password","expected_updated_at":"`+userPolicyUpdatedAtString+`"}`)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	if repo.updateCalled {
		t.Fatalf("user must not be updated when credential proof is missing")
	}
}
