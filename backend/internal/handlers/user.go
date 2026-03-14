package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"AL-Zauk/backend/internal/audit"
	"AL-Zauk/backend/internal/auth"
	"AL-Zauk/backend/internal/middleware"
	"AL-Zauk/backend/internal/models"
	"AL-Zauk/backend/internal/repository"
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

	// Send the event once - the frontend handles deduplication via localStorage
	h.eventsHandler.BroadcastToUser(userID, event)
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
	Password *string `json:"password"`
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

	if req.Email == nil || *req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Email is required",
		})
	}

	if req.Password == nil || *req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Password is required",
		})
	}

	if req.PIN == nil || *req.PIN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "PIN is required",
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

	// Audit log with new values
	newValues := map[string]interface{}{
		"affected_user": user.Name,
		"name":          user.Name,
	}
	if user.Role != nil {
		newValues["role"] = user.Role.Name
	}
	if user.Email != nil {
		newValues["email"] = *user.Email
	}
	audit.LogWithValues(c, models.AuditActionUserCreate, models.AuditEntityUser, user.ID.String(), "Created user: "+user.Name, nil, newValues)

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

	// Ensure Role is fetched for hierarchy check
	if user.Role == nil {
		userWithRole, err := h.userRepo.GetByID(c.Context(), id)
		if err == nil && userWithRole != nil {
			user.Role = userWithRole.Role
		}
	}

	currentRoleName := middleware.GetRoleName(c)
	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}

	// Enforce strict role hierarchy: Cannot modify users with a strictly higher role
	if getRoleLevel(currentRoleName) < getRoleLevel(targetRoleName) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "You do not have permission to modify a user with a higher role level.",
		})
	}

	// Capture old values for audit (include affected user name at top)
	oldValues := map[string]interface{}{
		"affected_user": user.Name,
		"name":          user.Name,
		"is_active":     user.IsActive,
	}
	if user.Role != nil {
		oldValues["role"] = user.Role.Name
	}
	if user.Email != nil {
		oldValues["email"] = *user.Email
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

		if roleID != user.RoleID {
			// Prevent self-role change
			if middleware.GetUserID(c) == id {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"code":    "CANNOT_CHANGE_OWN_ROLE",
					"message": "You cannot change your own role",
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

			// Clear all permission overrides when role changes
			if err := h.userRepo.ClearUserPermissionOverrides(c.Context(), id); err != nil {
				log.Error().Err(err).Msg("Failed to clear permission overrides")
				// Don't fail the role change if clearing overrides fails
			}
			user.RoleID = roleID
		}
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Handle Password update if provided
	if req.Password != nil && *req.Password != "" {
		hash, err := auth.HashPassword(*req.Password)
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

	// Fetch updated user with Role join for audit
	updatedForAudit, _ := h.userRepo.GetByID(c.Context(), id)

	// Audit log with old and new values
	newValues := map[string]interface{}{
		"affected_user": user.Name,
		"name":          user.Name,
		"is_active":     user.IsActive,
	}
	if updatedForAudit != nil && updatedForAudit.Role != nil {
		newValues["role"] = updatedForAudit.Role.Name
	}
	if user.Email != nil {
		newValues["email"] = *user.Email
	}
	// Track PIN change (don't log actual value for security)
	if req.PIN != nil {
		oldValues["pin"] = "[set]"
		if *req.PIN == "" {
			newValues["pin"] = "[removed]"
		} else {
			newValues["pin"] = "[changed]"
		}
	}
	// Track Password change (don't log actual value for security)
	if req.Password != nil && *req.Password != "" {
		oldValues["password"] = "[set]"
		newValues["password"] = "[changed]"
	}
	// Check if this is a restore action
	oldIsActive, ok := oldValues["is_active"].(bool)
	isRestore := req.IsActive != nil && *req.IsActive && ok && !oldIsActive

	action := models.AuditActionUserUpdate
	actionMsg := "Updated user: " + user.Name
	if isRestore {
		action = models.AuditActionUserRestore
		actionMsg = "Restored user: " + user.Name
	}

	audit.LogWithValues(c, action, models.AuditEntityUser, id.String(), actionMsg, oldValues, newValues)

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

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "User not found",
		})
	}

	currentRoleName := middleware.GetRoleName(c)
	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}
	if getRoleLevel(currentRoleName) < getRoleLevel(targetRoleName) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "You do not have permission to modify a user with a higher role level.",
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

	// Get user name for audit
	user, _ = h.userRepo.GetByID(c.Context(), id)
	userName := "Unknown"
	if user != nil {
		userName = user.Name
	}

	// Audit log (don't log actual password for security)
	audit.LogWithValues(c, models.AuditActionUserUpdate, models.AuditEntityUser, id.String(), "Updated password for: "+userName,
		map[string]interface{}{"affected_user": userName, "password": "[set]"},
		map[string]interface{}{"affected_user": userName, "password": "[changed]"})

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

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "User not found",
		})
	}

	currentRoleName := middleware.GetRoleName(c)
	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}
	if getRoleLevel(currentRoleName) < getRoleLevel(targetRoleName) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "You do not have permission to modify a user with a higher role level.",
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

	// Get user name for audit
	user, _ = h.userRepo.GetByID(c.Context(), id)
	userName := "Unknown"
	if user != nil {
		userName = user.Name
	}

	// Audit log (don't log actual PIN for security)
	newPinStatus := "[changed]"
	if pinHash == nil {
		newPinStatus = "[removed]"
	}
	audit.LogWithValues(c, models.AuditActionUserUpdate, models.AuditEntityUser, id.String(), "Updated PIN for: "+userName,
		map[string]interface{}{"affected_user": userName, "pin": "[set]"},
		map[string]interface{}{"affected_user": userName, "pin": newPinStatus})

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

	// Prevent managing own permissions
	currentUserID := middleware.GetUserID(c)
	if currentUserID == id {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "CANNOT_MODIFY_SELF",
			"message": "You cannot modify your own permissions",
		})
	}

	// Check role hierarchy - can only modify permissions of users with lower roles
	currentRoleName := middleware.GetRoleName(c)
	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}

	currentRoleLevel := getRoleLevel(currentRoleName)
	targetRoleLevel := getRoleLevel(targetRoleName)

	// Can only modify permissions of users with same or lower role level
	if targetRoleLevel > currentRoleLevel {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "You can only modify permissions of users with roles the same as or lower than yours",
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

	// Fetch current user's effective permissions to enforce grant restrictions.
	// A user can only grant permissions they themselves have (owners are exempt).
	var currentUserPermSet map[string]bool
	if strings.ToLower(currentRoleName) != "owner" {
		currentUserPerms, err := h.userRepo.GetUserPermissions(c.Context(), currentUserID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to fetch current user permissions")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to validate permissions",
			})
		}
		currentUserPermSet = make(map[string]bool, len(currentUserPerms))
		for _, p := range currentUserPerms {
			currentUserPermSet[p] = true
		}
	}

	// Get old permission overrides for audit trail (indexed by permission_id)
	oldOverrides, _ := h.userRepo.GetUserPermissionOverrides(c.Context(), id)
	oldOverrideMap := make(map[string]string) // permission_id -> "granted"/"denied"
	for _, override := range oldOverrides {
		if override.Permission != nil {
			status := "denied"
			if override.Allowed {
				status = "granted"
			}
			oldOverrideMap[override.PermissionID.String()] = status
		}
	}

	// Set each permission
	oldValues := make(map[string]interface{})
	newValues := make(map[string]interface{})
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

		// Non-owners can only grant permissions they themselves have.
		// Denying a permission is always allowed (it only restricts access).
		if perm.Allowed && currentUserPermSet != nil && !currentUserPermSet[permission.Key] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    "CANNOT_GRANT_UNOWNED_PERMISSION",
				"message": "You cannot grant the '" + permission.Name + "' permission because you do not have it yourself",
			})
		}

		if err := h.userRepo.SetUserPermission(c.Context(), id, permID, perm.Allowed, &grantedBy); err != nil {
			log.Error().Err(err).Msg("Failed to set user permission")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to set permissions",
			})
		}

		// Track old value for this specific permission
		if oldStatus, exists := oldOverrideMap[perm.PermissionID]; exists {
			oldValues[permission.Name] = oldStatus
		} else {
			oldValues[permission.Name] = "from role"
		}

		// Track new value for audit
		newStatus := "denied"
		if perm.Allowed {
			newStatus = "granted"
		}
		newValues[permission.Name] = newStatus
	}

	// Get updated permissions
	permissions, _ := h.userRepo.GetUserPermissions(c.Context(), id)
	overrides, _ := h.userRepo.GetUserPermissionOverrides(c.Context(), id)

	// Add affected user context
	oldValues["affected_user"] = user.Name
	newValues["affected_user"] = user.Name

	// Audit log with detailed changes
	audit.LogWithValues(c, models.AuditActionUserUpdate, models.AuditEntityUser, id.String(),
		"Updated permissions for: "+user.Name, oldValues, newValues)

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

	// Get user information for audit log and hierarchy check
	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "User not found",
		})
	}

	// Enforce strict role hierarchy: Cannot modify users with a strictly higher role
	currentRoleName := middleware.GetRoleName(c)
	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}
	if getRoleLevel(currentRoleName) < getRoleLevel(targetRoleName) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "You do not have permission to modify a user with a higher role level.",
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

	// Audit log
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}
	audit.LogWithValues(c, models.AuditActionUserArchive, models.AuditEntityUser, id.String(),
		"Archived user: "+user.Name,
		map[string]interface{}{"affected_user": user.Name, "name": user.Name, "role": roleName, "status": "active"},
		nil,
	)

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

	// Get user information for audit log and hierarchy check
	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "User not found",
		})
	}

	// Enforce strict role hierarchy: Cannot modify users with a strictly higher role
	currentRoleName := middleware.GetRoleName(c)
	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}
	if getRoleLevel(currentRoleName) < getRoleLevel(targetRoleName) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "You do not have permission to modify a user with a higher role level.",
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

	// Audit log
	delRoleName := ""
	if user.Role != nil {
		delRoleName = user.Role.Name
	}
	delEmail := ""
	if user.Email != nil {
		delEmail = *user.Email
	}
	audit.LogWithValues(c, models.AuditActionUserDelete, models.AuditEntityUser, id.String(),
		"Permanently deleted user: "+user.Name,
		map[string]interface{}{"affected_user": user.Name, "name": user.Name, "role": delRoleName, "email": delEmail, "status": "archived"},
		nil,
	)

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

// getRoleLevel returns an integer representing the hierarchy of the role
func getRoleLevel(roleName string) int {
	switch strings.ToLower(roleName) {
	case "owner":
		return 3
	case "manager":
		return 2
	case "cashier":
		return 1
	default:
		return 0
	}
}

// canAssignRole checks if a user with the given role can assign another role
func canAssignRole(currentRole, targetRole string) bool {
	return getRoleLevel(currentRole) >= getRoleLevel(targetRole)
}
