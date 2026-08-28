package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

type eventUserReader interface {
	GetByID(context.Context, uuid.UUID) (*models.User, error)
}

// UserEventType represents different types of user management events
type UserEventType string

const (
	EventUserUpdated        UserEventType = "user_updated"
	EventUserDeactivated    UserEventType = "user_deactivated"
	EventUserActivated      UserEventType = "user_activated"
	EventUserDeleted        UserEventType = "user_deleted"
	EventPermissionsChanged UserEventType = "permissions_changed"
	EventRoleChanged        UserEventType = "role_changed"
	EventForceLogout        UserEventType = "force_logout"
)

// UserEvent represents an event that affects a user
type UserEvent struct {
	Type      UserEventType `json:"type"`
	UserID    string        `json:"user_id"`
	ChangedBy string        `json:"changed_by,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Details   interface{}   `json:"details,omitempty"`
}

// Client represents a connected SSE client
type Client struct {
	ID      string
	UserID  uuid.UUID
	Channel chan UserEvent
	Done    chan struct{}
}

const (
	// sseMaxConnectionsPerUser caps concurrent event streams per account so a
	// single leaked token cannot exhaust file descriptors (one tab opens one
	// stream, so three leaves headroom for a kiosk with a stuck reload).
	sseMaxConnectionsPerUser = 3
	// sseMaxTotalConnections caps the whole process. The endpoint serves
	// in-store staff; this is generous headroom, not a scaling knob.
	sseMaxTotalConnections = 500
	// sseConnectionTTL forces periodic reconnects so half-open sockets from
	// dead clients cannot accumulate forever (WriteTimeout stays 0 because
	// streaming requires it). Clients reconnect automatically.
	sseConnectionTTL = 30 * time.Minute
)

// EventsHandler manages SSE connections and broadcasts user events
type EventsHandler struct {
	clients    map[string]*Client
	clientsMux sync.RWMutex
	jwtManager authTokenManager
	userRepo   eventUserReader
	origins    []string
}

// NewEventsHandler creates a new events handler
func NewEventsHandler(jwtManager authTokenManager, userRepo eventUserReader, origins []string) *EventsHandler {
	return &EventsHandler{
		clients:    make(map[string]*Client),
		jwtManager: jwtManager,
		userRepo:   userRepo,
		origins:    origins,
	}
}

// countClientsFor returns how many active streams belong to a user.
// Callers must hold clientsMux (write lock).
func (h *EventsHandler) countClientsFor(userID uuid.UUID) int {
	count := 0
	for _, client := range h.clients {
		if client.UserID == userID {
			count++
		}
	}
	return count
}

// Subscribe handles GET /api/v1/events/subscribe
// This establishes an SSE connection for the authenticated user
func (h *EventsHandler) Subscribe(c *fiber.Ctx) error {
	token := eventStreamToken(c)
	if token == "" {
		return middleware.JSONError(c, fiber.StatusUnauthorized, "MISSING_TOKEN", "Token is required")
	}

	// Validate the token
	claims, err := h.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return middleware.JSONError(c, fiber.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired access token")
	}

	user, err := h.userRepo.GetByID(c.Context(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to verify SSE user status")
		return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify authentication")
	}
	if user == nil || !user.IsActive {
		return middleware.JSONError(c, fiber.StatusUnauthorized, "ACCOUNT_INACTIVE", "Your account has been deactivated")
	}

	// Create a unique client ID
	clientID := uuid.New().String()

	// Create the client
	client := &Client{
		ID:      clientID,
		UserID:  claims.UserID,
		Channel: make(chan UserEvent, 10),
		Done:    make(chan struct{}),
	}

	// Register the client, enforcing per-user and process-wide connection
	// caps in the same critical section so the check cannot race a register.
	h.clientsMux.Lock()
	if len(h.clients) >= sseMaxTotalConnections || h.countClientsFor(claims.UserID) >= sseMaxConnectionsPerUser {
		h.clientsMux.Unlock()
		log.Warn().
			Str("user_id", claims.UserID.String()).
			Int("total", len(h.clients)).
			Msg("SSE connection rejected: connection cap reached")
		return middleware.JSONError(c, fiber.StatusTooManyRequests, "TOO_MANY_CONNECTIONS", "Too many event stream connections; close unused tabs and retry")
	}
	h.clients[clientID] = client
	h.clientsMux.Unlock()

	log.Debug().
		Str("client_id", clientID).
		Str("user_id", claims.UserID.String()).
		Msg("SSE client connected")

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")
	middleware.ApplyCORSHeaders(c, h.origins)

	// Use streaming - the cleanup must happen INSIDE the StreamWriter
	// because SetBodyStreamWriter returns immediately
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// Ensure cleanup happens when this function exits
		defer func() {
			h.clientsMux.Lock()
			delete(h.clients, clientID)
			h.clientsMux.Unlock()

			// Close the Done channel to signal any pending operations
			select {
			case <-client.Done:
				// Already closed
			default:
				close(client.Done)
			}

			log.Debug().
				Str("client_id", clientID).
				Str("user_id", claims.UserID.String()).
				Msg("SSE client disconnected")
		}()

		// Send initial connection event
		initialEvent := UserEvent{
			Type:      "connected",
			UserID:    claims.UserID.String(),
			Timestamp: time.Now(),
		}
		if err := h.sendEvent(w, initialEvent); err != nil {
			log.Error().Err(err).Msg("Failed to send initial SSE event")
			return
		}

		// Create a ticker for keepalive
		keepaliveTicker := time.NewTicker(15 * time.Second)
		defer keepaliveTicker.Stop()

		connectedAt := time.Now()

		for {
			select {
			case event, ok := <-client.Channel:
				if !ok {
					// Channel closed
					return
				}
				if err := h.sendEvent(w, event); err != nil {
					log.Error().Err(err).Msg("Failed to send SSE event")
					return
				}
			case <-keepaliveTicker.C:
				// Recycle the connection once its TTL is spent so sockets from
				// vanished clients cannot linger indefinitely; the browser's
				// EventSource reconnects on close.
				if time.Since(connectedAt) >= sseConnectionTTL {
					log.Debug().Str("client_id", clientID).Msg("SSE connection TTL reached, recycling")
					return
				}
				// Send keepalive comment
				if _, err := fmt.Fprintf(w, ": keepalive %s\n\n", time.Now().Format(time.RFC3339)); err != nil {
					log.Debug().Err(err).Msg("SSE keepalive failed, client disconnected")
					return
				}
				if err := w.Flush(); err != nil {
					log.Debug().Err(err).Msg("SSE flush failed, client disconnected")
					return
				}
			case <-client.Done:
				// Client was forcefully disconnected
				return
			}
		}
	}))

	return nil
}

func eventStreamToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}

	return ""
}

// sendEvent sends an SSE event to the writer
func (h *EventsHandler) sendEvent(w *bufio.Writer, event UserEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal event")
		return err
	}

	if _, err := fmt.Fprintf(w, "event: %s\n", event.Type); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return w.Flush()
}

// BroadcastToUser sends an event to all clients connected for a specific user
func (h *EventsHandler) BroadcastToUser(userID uuid.UUID, event UserEvent) {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()

	clientCount := 0
	for _, client := range h.clients {
		if client.UserID == userID {
			clientCount++
			select {
			case client.Channel <- event:
			default:
				log.Warn().
					Str("client_id", client.ID).
					Msg("Client channel full, event dropped")
			}
		}
	}

	if clientCount == 0 {
		log.Debug().
			Str("user_id", userID.String()).
			Str("event_type", string(event.Type)).
			Msg("No connected clients for user, event not delivered")
	}
}

// BroadcastToAll sends an event to all connected clients
func (h *EventsHandler) BroadcastToAll(event UserEvent) {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Channel <- event:
		default:
			log.Warn().
				Str("client_id", client.ID).
				Msg("Client channel full, event dropped")
		}
	}
}

// GetConnectedClientCount returns the number of connected clients
func (h *EventsHandler) GetConnectedClientCount() int {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()
	return len(h.clients)
}

// DisconnectUser forcefully disconnects all clients for a user
func (h *EventsHandler) DisconnectUser(userID uuid.UUID) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	for clientID, client := range h.clients {
		if client.UserID == userID {
			close(client.Done)
			delete(h.clients, clientID)
			log.Info().
				Str("client_id", clientID).
				Str("user_id", userID.String()).
				Msg("Client forcefully disconnected")
		}
	}
}
