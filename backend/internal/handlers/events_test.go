package handlers

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	authpkg "dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/models"
)

func TestEventsHandlerSubscribeRequiresToken(t *testing.T) {
	handler := NewEventsHandler(&fakeJWTManager{}, &fakeAuthUserRepo{}, []string{"http://localhost:3000"})
	app := fiber.New()
	app.Get("/events/subscribe", handler.Subscribe)

	req := httptest.NewRequest("GET", "/events/subscribe", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestEventsHandlerSubscribeRejectsInvalidToken(t *testing.T) {
	handler := NewEventsHandler(&fakeJWTManager{validateAccessErr: fiber.ErrUnauthorized}, &fakeAuthUserRepo{}, []string{"http://localhost:3000"})
	app := fiber.New()
	app.Get("/events/subscribe", handler.Subscribe)

	req := httptest.NewRequest("GET", "/events/subscribe", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestEventStreamTokenPrefersAuthorizationHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		token := eventStreamToken(c)
		if token != "valid-token" {
			t.Fatalf("expected bearer token, got %q", token)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/?token=query-token", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}
}

func TestEventStreamTokenRejectsQueryParam(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		token := eventStreamToken(c)
		if token != "" {
			t.Fatalf("expected empty token, got %q", token)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/?token=query-token", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}
}

func TestEventsHandlerSubscribeRejectsInactiveUser(t *testing.T) {
	userID := uuid.New()
	handler := NewEventsHandler(
		&fakeJWTManager{validateAccessClaims: &authpkg.Claims{UserID: userID}},
		&fakeAuthUserRepo{userByID: &models.User{ID: userID, IsActive: false}},
		[]string{"http://localhost:3000"},
	)
	app := fiber.New()
	app.Get("/events/subscribe", handler.Subscribe)

	req := httptest.NewRequest("GET", "/events/subscribe", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestEventsHandlerBroadcastToMatchingUser(t *testing.T) {
	userID := uuid.New()
	otherID := uuid.New()
	handler := NewEventsHandler(&fakeJWTManager{
		validateAccessClaims: &authpkg.Claims{UserID: userID},
	}, &fakeAuthUserRepo{}, []string{"http://localhost:3000"})

	matchClient := &Client{ID: "1", UserID: userID, Channel: make(chan UserEvent, 1), Done: make(chan struct{})}
	otherClient := &Client{ID: "2", UserID: otherID, Channel: make(chan UserEvent, 1), Done: make(chan struct{})}
	handler.clients["1"] = matchClient
	handler.clients["2"] = otherClient

	handler.BroadcastToUser(userID, UserEvent{Type: EventUserUpdated, UserID: userID.String(), Timestamp: time.Now()})

	select {
	case <-matchClient.Channel:
	default:
		t.Fatalf("expected matching client to receive event")
	}

	select {
	case <-otherClient.Channel:
		t.Fatalf("did not expect other client to receive event")
	default:
	}
}

func TestEventsHandlerDisconnectUserRemovesClients(t *testing.T) {
	userID := uuid.New()
	handler := NewEventsHandler(&fakeJWTManager{}, &fakeAuthUserRepo{}, []string{"http://localhost:3000"})
	handler.clients["1"] = &Client{ID: "1", UserID: userID, Channel: make(chan UserEvent, 1), Done: make(chan struct{})}
	handler.clients["2"] = &Client{ID: "2", UserID: uuid.New(), Channel: make(chan UserEvent, 1), Done: make(chan struct{})}

	handler.DisconnectUser(userID)

	if handler.GetConnectedClientCount() != 1 {
		t.Fatalf("expected one remaining client, got %d", handler.GetConnectedClientCount())
	}
	if _, exists := handler.clients["1"]; exists {
		t.Fatalf("expected target user client to be removed")
	}
}
