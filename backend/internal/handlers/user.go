package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// normalizeEmail converts email to lowercase for case-insensitive storage
func normalizeEmail(email *string) *string {
	if email == nil || *email == "" {
		return email
	}
	normalized := strings.ToLower(strings.TrimSpace(*email))
	return &normalized
}

// UserHandler handles user management endpoints
type UserHandler struct {
	userRepo       *repository.UserRepository
	roleRepo       *repository.RoleRepository
	permissionRepo *repository.PermissionRepository
	eventsHandler  *EventsHandler
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	permissionRepo *repository.PermissionRepository,
) *UserHandler {
	return &UserHandler{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// SetEventsHandler sets the events handler for broadcasting user updates
func (h *UserHandler) SetEventsHandler(eventsHandler *EventsHandler) {
	h.eventsHandler = eventsHandler
}

// broadcastUserEvent broadcasts a user event if events handler is configured
// It sends the event multiple times with delays to ensure delivery even if the client
// is temporarily disconnected/reconnecting
func (h *UserHandler) broadcastUserEvent(userID uuid.UUID, eventType UserEventType, changedBy uuid.UUID, details interface{}) {
	if h.eventsHandler == nil {
		log.Warn().Msg("Events handler not configured, cannot broadcast user event")
		return
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("event_type", string(eventType)).
		Str("changed_by", changedBy.String()).
		Msg("Broadcasting user event")

	event := UserEvent{
		Type:      eventType,
		UserID:    userID.String(),
		ChangedBy: changedBy.String(),
		Timestamp: time.Now(),
		Details:   details,
	}

	// Send immediately
	h.eventsHandler.BroadcastToUser(userID, event)

	// Also send with delays to catch reconnecting clients
	// This runs in a goroutine so it doesn't block the response
	go func() {
		delays := []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond, 3000 * time.Millisecond}
		for i, delay := range delays {
			time.Sleep(delay)
			log.Debug().
				Str("user_id", userID.String()).
				Str("event_type", string(eventType)).
				Int("retry", i+1).
				Msg("Retry broadcasting user event")
			h.eventsHandler.BroadcastToUser(userID, event)
		}
	}()
}

// UserListResponse represents a paginated list of users
type UserListResponse struct {
	Users      []UserDetailResponse `json:"users"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	TotalPages int                  `json:"total_pages"`
}

// UserDetailResponse represents a user in API responses
type UserDetailResponse struct {
	ID          string        `json:"id"`
	Email       *string       `json:"email,omitempty"`
	Name        string        `json:"name"`
	RoleID      string        `json:"role_id"`
	RoleName    string        `json:"role_name"`
	IsActive    bool          `json:"is_active"`
	HasPIN      bool          `json:"has_pin"`
	LastLoginAt *string       `json:"last_login_at,omitempty"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
	Role        *RoleResponse `json:"role,omitempty"`
}

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email    *string `json:"email"`
	Name     string  `json:"name"`
	Password *string `json:"password"`
	PIN      *string `json:"pin"`
	RoleID   string  `json:"role_id"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email    *string `json:"email"`
	Name     *string `json:"name"`
	RoleID   *string `json:"role_id"`
	IsActive *bool   `json:"is_active"`
	PIN      *string `json:"pin"`
}

// UpdatePasswordRequest represents the request to update a password
type UpdatePasswordRequest struct {
	Password string `json:"password"`
}

// UpdatePINRequest represents the request to update a PIN
type UpdatePINRequest struct {
	PIN *string `json:"pin"` // nil to remove PIN
}

// SetPermissionsRequest represents the request to set user permissions
type SetPermissionsRequest struct {
	Permissions []PermissionOverride `json:"permissions"`
}

// PermissionOverride represents a single permission override
type PermissionOverride struct {
	PermissionID string `json:"permission_id"`
	Allowed      bool   `json:"allowed"`
}

// List handles GET /api/v1/users
func (h *UserHandler) List(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	activeOnlyStr := c.Query("active_only", "")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// Parse active filter: "true" = active only, "false" = inactive only, "" = all
	var isActive *bool
	if activeOnlyStr == "true" {
		active := true
		isActive = &active
	} else if activeOnlyStr == "false" {
		active := false
		isActive = &active
	}

	users, total, err := h.userRepo.ListWithFilter(c.Context(), perPage, offset, isActive)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve users",
		})
	}

	// Convert to response format
	userResponses := make([]UserDetailResponse, len(users))
	for i, user := range users {
		userResponses[i] = h.toUserDetailResponse(user)
	}

	totalPages := (total + perPage - 1) / perPage

	return c.JSON(UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// Get handles GET /api/v1/users/:id
func (h *UserHandler) Get(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid user ID format",
		})
	}

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "User not found",
		})
	}

	// Get user's permissions
	permissions, err := h.userRepo.GetUserPermissions(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user permissions")
	}

	response := h.toUserDetailResponse(user)

	return c.JSON(fiber.Map{
		"user":        response,
		"permissions": permissions,
	})
}

// Create handles POST /api/v1/users
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Name is required",
		})
	}

	if req.RoleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Role ID is required",
		})
	}

	// Parse role ID
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ROLE_ID",
			"message": "Invalid role ID format",
		})
	}

	// Verify role exists
	role, err := h.roleRepo.GetByID(c.Context(), roleID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get role")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to validate role",
		})
	}
	if role == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ROLE",
			"message": "Role not found",
		})
	}

	// Check if current user can assign this role
	currentRoleName := middleware.GetRoleName(c)
	if !canAssignRole(currentRoleName, role.Name) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "You cannot assign the " + role.Name + " role",
		})
	}

	// Check email uniqueness if provided
	if req.Email != nil && *req.Email != "" {
		exists, err := h.userRepo.EmailExists(c.Context(), *req.Email, nil)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check email")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to validate email",
			})
		}
		if exists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "EMAIL_EXISTS",
				"message": "Email is already in use",
			})
		}
	}

	// Create user model
	user := &models.User{
		Email:    normalizeEmail(req.Email),
		Name:     req.Name,
		RoleID:   roleID,
		IsActive: true,
	}

	// Hash password if provided
	if req.Password != nil && *req.Password != "" {
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash password")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to process password",
			})
		}
		user.PasswordHash = &hash
	}

	// Hash PIN if provided
	if req.PIN != nil && *req.PIN != "" {
		hash, err := auth.HashPIN(*req.PIN)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash PIN")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to process PIN",
			})
		}
		user.PINHash = &hash
	}

	// Create the user
	if err := h.userRepo.Create(c.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create user",
		})
	}

	// Fetch the created user with role
	createdUser, _ := h.userRepo.GetByID(c.Context(), user.ID)
	if createdUser != nil {
		user = createdUser
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user":    h.toUserDetailResponse(user),
	})
}

// Update handles PATCH /api/v1/users/:id
func (h *UserHandler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid user ID format",
		})
	}

	// Get existing user
	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve user",
		})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "User not found",
		})
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	// Update fields if provided
	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Email != nil {
		// Check email uniqueness
		if *req.Email != "" {
			exists, err := h.userRepo.EmailExists(c.Context(), *req.Email, &id)
			if err != nil {
				log.Error().Err(err).Msg("Failed to check email")
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to validate email",
				})
			}
			if exists {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"code":    "EMAIL_EXISTS",
					"message": "Email is already in use",
				})
			}
		}
		user.Email = normalizeEmail(req.Email)
	}

	if req.RoleID != nil {
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_ROLE_ID",
				"message": "Invalid role ID format",
			})
		}

		// Verify role exists
		role, err := h.roleRepo.GetByID(c.Context(), roleID)
		if err != nil || role == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_ROLE",
				"message": "Role not found",
			})
		}

		// Check if current user can assign this role
		currentRoleName := middleware.GetRoleName(c)
		if !canAssignRole(currentRoleName, role.Name) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    "FORBIDDEN",
				"message": "You cannot assign the " + role.Name + " role",
			})
		}

		user.RoleID = roleID
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Handle PIN update if provided
	if req.PIN != nil {
		var pinHash *string
		if *req.PIN != "" {
			hash, err := auth.HashPIN(*req.PIN)
			if err != nil {
				log.Error().Err(err).Msg("Failed to hash PIN")
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to process PIN",
				})
			}
			pinHash = &hash
		}
		if err := h.userRepo.UpdatePIN(c.Context(), id, pinHash); err != nil {
			log.Error().Err(err).Msg("Failed to update PIN")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to update PIN",
			})
		}
	}

	// Update the user
	if err := h.userRepo.Update(c.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to update user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update user",
		})
	}

	// Fetch updated user with Role join
	updatedUser, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch updated user")
		// Fallback to basic response if join fails
		return c.JSON(fiber.Map{
			"message": "User updated successfully",
			"user":    h.toUserDetailResponse(user),
		})
	}

	// Broadcast user update event
	currentUserID := middleware.GetUserID(c)
	if req.RoleID != nil {
		// Role was changed
		h.broadcastUserEvent(id, EventRoleChanged, currentUserID, map[string]interface{}{
			"new_role_id":   *req.RoleID,
			"new_role_name": updatedUser.Role.Name,
		})
	} else if req.IsActive != nil {
		// Active status was changed
		if *req.IsActive {
			h.broadcastUserEvent(id, EventUserActivated, currentUserID, nil)
		} else {
			h.broadcastUserEvent(id, EventUserDeactivated, currentUserID, nil)
			// Also disconnect the user if deactivated
			if h.eventsHandler != nil {
				h.eventsHandler.DisconnectUser(id)
			}
		}
	} else {
		// General update (name, email, etc.)
		h.broadcastUserEvent(id, EventUserUpdated, currentUserID, nil)
	}

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
		"user":    h.toUserDetailResponse(updatedUser),
	})
}

// UpdatePassword handles PATCH /api/v1/users/:id/password
func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid user ID format",
		})
	}

	var req UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Password is required",
		})
	}

	// Hash the new password
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process password",
		})
	}

	if err := h.userRepo.UpdatePassword(c.Context(), id, hash); err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password updated successfully",
	})
}

// UpdatePIN handles PATCH /api/v1/users/:id/pin
func (h *UserHandler) UpdatePIN(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid user ID format",
		})
	}

	var req UpdatePINRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	var pinHash *string
	if req.PIN != nil && *req.PIN != "" {
		hash, err := auth.HashPIN(*req.PIN)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash PIN")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to process PIN",
			})
		}
		pinHash = &hash
	}

	if err := h.userRepo.UpdatePIN(c.Context(), id, pinHash); err != nil {
		log.Error().Err(err).Msg("Failed to update PIN")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update PIN",
		})
	}

	message := "PIN updated successfully"
	if pinHash == nil {
		message = "PIN removed successfully"
	}

	return c.JSON(fiber.Map{
		"message": message,
	})
}

// SetPermissions handles PATCH /api/v1/users/:id/permissions
func (h *UserHandler) SetPermissions(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid user ID format",
		})
	}

	// Verify user exists
	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "User not found",
		})
	}

	var req SetPermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	grantedBy := middleware.GetUserID(c)

	// Set each permission
	for _, perm := range req.Permissions {
		permID, err := uuid.Parse(perm.PermissionID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_PERMISSION_ID",
				"message": "Invalid permission ID: " + perm.PermissionID,
			})
		}

		// Verify permission exists
		permission, err := h.permissionRepo.GetByID(c.Context(), permID)
		if err != nil || permission == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_PERMISSION",
				"message": "Permission not found: " + perm.PermissionID,
			})
		}

		if err := h.userRepo.SetUserPermission(c.Context(), id, permID, perm.Allowed, &grantedBy); err != nil {
			log.Error().Err(err).Msg("Failed to set user permission")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to set permissions",
			})
		}
	}

	// Get updated permissions
	permissions, _ := h.userRepo.GetUserPermissions(c.Context(), id)
	overrides, _ := h.userRepo.GetUserPermissionOverrides(c.Context(), id)

	// Broadcast permissions changed event
	h.broadcastUserEvent(id, EventPermissionsChanged, grantedBy, map[string]interface{}{
		"permissions": permissions,
	})

	return c.JSON(fiber.Map{
		"message":               "Permissions updated successfully",
		"effective_permissions": permissions,
		"overrides":             len(overrides),
	})
}

// Delete handles DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid user ID format",
		})
	}

	// Prevent self-deletion
	currentUserID := middleware.GetUserID(c)
	if currentUserID == id {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "CANNOT_DELETE_SELF",
			"message": "You cannot deactivate your own account",
		})
	}

	// Deactivate the user
	if err := h.userRepo.Deactivate(c.Context(), id); err != nil {
		log.Error().Err(err).Msg("Failed to deactivate user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to deactivate user",
		})
	}

	// Broadcast deactivation event and disconnect the user
	h.broadcastUserEvent(id, EventUserDeactivated, currentUserID, nil)
	if h.eventsHandler != nil {
		h.eventsHandler.DisconnectUser(id)
	}

	return c.JSON(fiber.Map{
		"message": "User deactivated successfully",
	})
}

// PermanentDelete handles DELETE /api/v1/users/:id/permanent
func (h *UserHandler) PermanentDelete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid user ID format",
		})
	}

	// Prevent self-deletion
	currentUserID := middleware.GetUserID(c)
	if currentUserID == id {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "CANNOT_DELETE_SELF",
			"message": "You cannot delete your own account",
		})
	}

	// Check if user has sales history
	hasSales, err := h.userRepo.HasSalesHistory(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check sales history")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to check user sales history",
		})
	}

	if hasSales {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"code":    "HAS_SALES_HISTORY",
			"message": "Cannot permanently delete user with sales history. The user must remain archived.",
		})
	}

	if err := h.userRepo.PermanentDelete(c.Context(), id); err != nil {
		log.Error().Err(err).Msg("Failed to permanently delete user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to permanently delete user",
		})
	}

	// Broadcast delete event and disconnect the user
	h.broadcastUserEvent(id, EventUserDeleted, currentUserID, nil)
	if h.eventsHandler != nil {
		h.eventsHandler.DisconnectUser(id)
	}

	return c.JSON(fiber.Map{
		"message": "User permanently deleted",
	})
}

// GetPermissions handles GET /api/v1/users/:id/permissions
func (h *UserHandler) GetPermissions(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid user ID format",
		})
	}

	// Get effective permissions
	permissions, err := h.userRepo.GetUserPermissions(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user permissions")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve permissions",
		})
	}

	// Get overrides
	overrides, err := h.userRepo.GetUserPermissionOverrides(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get permission overrides")
	}

	return c.JSON(fiber.Map{
		"effective_permissions": permissions,
		"overrides":             overrides,
	})
}

// Helper functions

func (h *UserHandler) toUserDetailResponse(user *models.User) UserDetailResponse {
	response := UserDetailResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		RoleID:    user.RoleID.String(),
		IsActive:  user.IsActive,
		HasPIN:    user.PINHash != nil,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if user.LastLoginAt != nil {
		formatted := user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		response.LastLoginAt = &formatted
	}

	if user.Role != nil {
		response.RoleName = user.Role.Name
		response.Role = &RoleResponse{
			ID:          user.Role.ID.String(),
			Name:        user.Role.Name,
			Description: user.Role.Description,
		}
	}

	return response
}

// canAssignRole checks if a user with the given role can assign another role
func canAssignRole(currentRole, targetRole string) bool {
	// Owner can assign any role
	if currentRole == "owner" {
		return true
	}

	// Manager can only assign manager or cashier
	if currentRole == "manager" {
		return targetRole == "manager" || targetRole == "cashier"
	}

	// Cashiers cannot assign roles
	return false
}
