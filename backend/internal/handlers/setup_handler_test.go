package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	authpkg "dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/models"
)

type fakeSetupUserStore struct {
	hasActiveUser bool
	hasActiveErr  error
	createErr     error
	createWins    bool
	created       *models.User
}

func (f *fakeSetupUserStore) HasActiveUser(context.Context) (bool, error) {
	return f.hasActiveUser, f.hasActiveErr
}

func (f *fakeSetupUserStore) CreateInitialOwner(_ context.Context, user *models.User) (bool, error) {
	if f.createErr != nil {
		return false, f.createErr
	}
	if !f.createWins {
		return false, nil
	}
	f.created = user
	return true, nil
}

func (f *fakeSetupUserStore) EmailExists(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}

type fakeSetupRoleStore struct {
	role *models.Role
	err  error
}

func (f *fakeSetupRoleStore) GetByName(context.Context, string) (*models.Role, error) {
	return f.role, f.err
}

func newSetupTestApp(userRepo *fakeSetupUserStore, roleRepo *fakeSetupRoleStore) *fiber.App {
	handler := NewSetupHandler(userRepo, roleRepo)
	app := fiber.New()
	app.Get("/setup/status", handler.Status)
	app.Post("/setup/owner", handler.CreateOwner)
	return app
}

func TestSetupStatusRequiresSetupWhenNoActiveUsers(t *testing.T) {
	app := newSetupTestApp(&fakeSetupUserStore{}, &fakeSetupRoleStore{})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/setup/status", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body struct {
		SetupRequired bool `json:"setup_required"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !body.SetupRequired {
		t.Fatalf("expected setup_required=true when no active users exist")
	}
}

func TestSetupStatusReportsCompletedWhenActiveUsersExist(t *testing.T) {
	app := newSetupTestApp(
		&fakeSetupUserStore{hasActiveUser: true},
		&fakeSetupRoleStore{},
	)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/setup/status", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var body struct {
		SetupRequired bool `json:"setup_required"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.SetupRequired {
		t.Fatalf("expected setup_required=false when an active user exists")
	}
}

func TestSetupCreateOwnerSuccess(t *testing.T) {
	ownerRoleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userStore := &fakeSetupUserStore{createWins: true}
	roleStore := &fakeSetupRoleStore{
		role: &models.Role{ID: ownerRoleID, Name: "owner"},
	}
	app := newSetupTestApp(userStore, roleStore)

	body := `{"name":"Store Owner","email":"Owner@Example.com","password":"correct-horse-battery","pin":"2468"}`
	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/setup/owner", bytes.NewBufferString(body)))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	created := userStore.created
	if created == nil {
		t.Fatalf("expected CreateInitialOwner to be called with a user")
	}
	if created.RoleID != ownerRoleID {
		t.Fatalf("expected owner role ID %s, got %s", ownerRoleID, created.RoleID)
	}
	if created.Email == nil || *created.Email != "owner@example.com" {
		t.Fatalf("expected normalized email owner@example.com, got %v", created.Email)
	}
	if created.PasswordHash == nil || authpkg.CheckPassword("correct-horse-battery", *created.PasswordHash) == false {
		t.Fatalf("expected password to be hashed from request value")
	}
	if created.PINHash == nil || authpkg.CheckPIN("2468", *created.PINHash) == false {
		t.Fatalf("expected PIN to be hashed from request value")
	}
	if !created.IsActive {
		t.Fatalf("expected created owner to be active")
	}
}

func TestSetupCreateOwnerRefusesWhenSetupCompleted(t *testing.T) {
	app := newSetupTestApp(
		&fakeSetupUserStore{createWins: false},
		&fakeSetupRoleStore{role: &models.Role{ID: uuid.New(), Name: "owner"}},
	)

	body := `{"name":"Late Owner","email":"late@example.com","password":"long-enough-password","pin":"1357"}`
	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/setup/owner", bytes.NewBufferString(body)))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}

	var result struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Code != "SETUP_ALREADY_COMPLETED" {
		t.Fatalf("expected code SETUP_ALREADY_COMPLETED, got %q", result.Code)
	}
}

func TestSetupCreateOwnerValidatesInput(t *testing.T) {
	cases := []struct {
		name string
		body string
		code string
	}{
		{
			name: "missing name",
			body: `{"name":"","email":"a@example.com","password":"long-enough","pin":"1234"}`,
			code: "VALIDATION_ERROR",
		},
		{
			name: "invalid email",
			body: `{"name":"Owner","email":"not-an-email","password":"long-enough","pin":"1234"}`,
			code: "VALIDATION_ERROR",
		},
		{
			name: "password below minimum length",
			body: `{"name":"Owner","email":"a@example.com","password":"short","pin":"1234"}`,
			code: "VALIDATION_ERROR",
		},
		{
			name: "PIN with non-digits",
			body: `{"name":"Owner","email":"a@example.com","password":"long-enough","pin":"12ab"}`,
			code: "VALIDATION_ERROR",
		},
		{
			name: "unknown field rejected",
			body: `{"name":"Owner","email":"a@example.com","password":"long-enough","pin":"1234","role_id":"surprise"}`,
			code: "INVALID_REQUEST",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := newSetupTestApp(
				&fakeSetupUserStore{createWins: true},
				&fakeSetupRoleStore{role: &models.Role{Name: "owner"}},
			)

			resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/setup/owner", bytes.NewBufferString(tc.body)))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", resp.StatusCode)
			}

			var result struct {
				Code string `json:"code"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if result.Code != tc.code {
				t.Fatalf("expected code %q, got %q", tc.code, result.Code)
			}
		})
	}
}

func TestSetupCreateOwnerFailsWithoutOwnerRole(t *testing.T) {
	app := newSetupTestApp(
		&fakeSetupUserStore{createWins: true},
		&fakeSetupRoleStore{},
	)

	body := `{"name":"Owner","email":"a@example.com","password":"long-enough","pin":"1234"}`
	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/setup/owner", bytes.NewBufferString(body)))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.StatusCode)
	}
}
