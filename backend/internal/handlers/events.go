package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"dashpoint/backend/internal/auth"
)

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

// EventsHandler manages SSE connections and broadcasts user events
type EventsHandler struct {
	clients    map[string]*Client
	clientsMux sync.RWMutex
	jwtManager *auth.JWTManager
}

// NewEventsHandler creates a new events handler
func NewEventsHandler(jwtManager *auth.JWTManager) *EventsHandler {
	return &EventsHandler{
		clients:    make(map[string]*Client),
		jwtManager: jwtManager,
	}
}

// Subscribe handles GET /api/v1/events/subscribe
// This establishes an SSE connection for the authenticated user
func (h *EventsHandler) Subscribe(c *fiber.Ctx) error {
	// Get the token from query parameter (SSE can't use headers)
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "MISSING_TOKEN",
			"message": "Token is required",
		})
	}

	// Validate the token
	claims, err := h.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_TOKEN",
			"message": "Invalid or expired access token",
		})
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

	// Register the client
	h.clientsMux.Lock()
	h.clients[clientID] = client
	h.clientsMux.Unlock()

	log.Debug().
		Str("client_id", clientID).
		Str("user_id", claims.UserID.String()).
		Int("total_clients", len(h.clients)).
		Msg("SSE client connected")

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("X-Accel-Buffering", "no")

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

		for {
			select {
			case event, ok := <-client.Channel:
				if !ok {
					// Channel closed
					return
				}
				log.Info().
					Str("client_id", clientID).
					Str("event_type", string(event.Type)).
					Str("user_id", event.UserID).
					Msg("Sending SSE event to client")
				if err := h.sendEvent(w, event); err != nil {
					log.Error().Err(err).Msg("Failed to send SSE event")
					return
				}
				log.Info().
					Str("client_id", clientID).
					Str("event_type", string(event.Type)).
					Msg("SSE event sent successfully")
			case <-keepaliveTicker.C:
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
				log.Info().
					Str("client_id", client.ID).
					Str("user_id", userID.String()).
					Str("event_type", string(event.Type)).
					Msg("Event queued for client")
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
	} else {
		log.Info().
			Str("user_id", userID.String()).
			Str("event_type", string(event.Type)).
			Int("client_count", clientCount).
			Msg("Event broadcast to user's clients")
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
