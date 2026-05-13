package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

func normalizeEmail(email *string) *string {
	if email == nil || *email == "" {
		return email
	}
	normalized := strings.ToLower(strings.TrimSpace(*email))
	return &normalized
}

// UserHandler handles user management endpoints.
type UserHandler struct {
	userRepo       userRepository
	roleRepo       roleReader
	permissionRepo permissionReader
	eventsHandler  userEventBroadcaster
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository, permissionRepo *repository.PermissionRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo, roleRepo: roleRepo, permissionRepo: permissionRepo}
}

// SetEventsHandler sets the events handler for broadcasting user updates.
func (h *UserHandler) SetEventsHandler(eventsHandler userEventBroadcaster) {
	h.eventsHandler = eventsHandler
}

func (h *UserHandler) broadcastUserEvent(userID uuid.UUID, eventType UserEventType, changedBy uuid.UUID, details interface{}) {
	if h.eventsHandler == nil {
		log.Warn().Msg("Events handler not configured, cannot broadcast user event")
		return
	}

	log.Info().Str("user_id", userID.String()).Str("event_type", string(eventType)).Str("changed_by", changedBy.String()).Msg("Broadcasting user event")
	h.eventsHandler.BroadcastToUser(userID, UserEvent{
		Type:      eventType,
		UserID:    userID.String(),
		ChangedBy: changedBy.String(),
		Timestamp: time.Now(),
		Details:   details,
	})
}

type UserListResponse struct {
	Users      []UserDetailResponse `json:"users"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	TotalPages int                  `json:"total_pages"`
}

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

type RoleResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type CreateUserRequest struct {
	Email    *string `json:"email"`
	Name     string  `json:"name"`
	Password *string `json:"password"`
	PIN      *string `json:"pin"`
	RoleID   string  `json:"role_id"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email"`
	Name     *string `json:"name"`
	RoleID   *string `json:"role_id"`
	IsActive *bool   `json:"is_active"`
	PIN      *string `json:"pin"`
	Password *string `json:"password"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password"`
}

type UpdatePINRequest struct {
	PIN *string `json:"pin"`
}

type SetPermissionsRequest struct {
	Permissions []PermissionOverride `json:"permissions"`
}

type PermissionOverride struct {
	PermissionID string `json:"permission_id"`
	Allowed      bool   `json:"allowed"`
}

// Create handles POST /api/v1/users.
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}
	if req.Name == "" {
		return badUserRequest(c, "VALIDATION_ERROR", "Name is required")
	}
	if req.RoleID == "" {
		return badUserRequest(c, "VALIDATION_ERROR", "Role ID is required")
	}
	if req.Email == nil || *req.Email == "" {
		return badUserRequest(c, "VALIDATION_ERROR", "Email is required")
	}
	if req.Password == nil || *req.Password == "" {
		return badUserRequest(c, "VALIDATION_ERROR", "Password is required")
	}
	if req.PIN == nil || *req.PIN == "" {
		return badUserRequest(c, "VALIDATION_ERROR", "PIN is required")
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return badUserRequest(c, "INVALID_ROLE_ID", "Invalid role ID format")
	}
	role, err := h.roleRepo.GetByID(c.Context(), roleID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get role")
		return userInternalError(c, "Failed to validate role")
	}
	if role == nil {
		return badUserRequest(c, "INVALID_ROLE", "Role not found")
	}

	if !canAssignRole(middleware.GetRoleName(c), role.Name) {
		return userForbidden(c, "You cannot assign the "+role.Name+" role")
	}
	if exists, err := h.userRepo.EmailExists(c.Context(), *req.Email, nil); err != nil {
		log.Error().Err(err).Msg("Failed to check email")
		return userInternalError(c, "Failed to validate email")
	} else if exists {
		return userConflict(c, "EMAIL_EXISTS", "Email is already in use")
	}
	if err := h.requireActionPermission(c, role.Name, userActionCreate); err != nil {
		return err
	}

	user := &models.User{Email: normalizeEmail(req.Email), Name: req.Name, RoleID: roleID, IsActive: true}
	if hash, err := auth.HashPassword(*req.Password); err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return userInternalError(c, "Failed to process password")
	} else {
		user.PasswordHash = &hash
	}
	if hash, err := auth.HashPIN(*req.PIN); err != nil {
		log.Error().Err(err).Msg("Failed to hash PIN")
		return userInternalError(c, "Failed to process PIN")
	} else {
		user.PINHash = &hash
	}

	if err := h.userRepo.Create(c.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return userInternalError(c, "Failed to create user")
	}

	createdUser, _ := h.userRepo.GetByID(c.Context(), user.ID)
	if createdUser != nil {
		user = createdUser
	}

	newValues := map[string]interface{}{"affected_user": user.Name, "name": user.Name}
	if user.Role != nil {
		newValues["role"] = user.Role.Name
	}
	if user.Email != nil {
		newValues["email"] = *user.Email
	}
	audit.LogWithValues(c, models.AuditActionUserCreate, models.AuditEntityUser, user.ID.String(), "Created user: "+user.Name, nil, newValues)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User created successfully", "user": h.toUserDetailResponse(user)})
}

// Update handles PATCH /api/v1/users/:id.
func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return userInternalError(c, "Failed to retrieve user")
	}
	if user == nil {
		return userNotFound(c)
	}
	if user.Role == nil {
		userWithRole, err := h.userRepo.GetByID(c.Context(), id)
		if err == nil && userWithRole != nil {
			user.Role = userWithRole.Role
		}
	}

	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}
	if err := h.enforceTargetUserAction(c, targetRoleName, userActionEdit); err != nil {
		return err
	}

	oldValues := map[string]interface{}{"affected_user": user.Name, "name": user.Name, "is_active": user.IsActive}
	if user.Role != nil {
		oldValues["role"] = user.Role.Name
	}
	if user.Email != nil {
		oldValues["email"] = *user.Email
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		if *req.Email != "" {
			exists, err := h.userRepo.EmailExists(c.Context(), *req.Email, &id)
			if err != nil {
				log.Error().Err(err).Msg("Failed to check email")
				return userInternalError(c, "Failed to validate email")
			}
			if exists {
				return userConflict(c, "EMAIL_EXISTS", "Email is already in use")
			}
		}
		user.Email = normalizeEmail(req.Email)
	}

	if req.RoleID != nil {
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			return badUserRequest(c, "INVALID_ROLE_ID", "Invalid role ID format")
		}
		if roleID != user.RoleID {
			if middleware.GetUserID(c) == id {
				return middleware.JSONError(c, fiber.StatusForbidden, "CANNOT_CHANGE_OWN_ROLE", "You cannot change your own role")
			}
			role, err := h.roleRepo.GetByID(c.Context(), roleID)
			if err != nil || role == nil {
				return badUserRequest(c, "INVALID_ROLE", "Role not found")
			}
			if !canAssignRole(middleware.GetRoleName(c), role.Name) {
				return userForbidden(c, "You cannot assign the "+role.Name+" role")
			}
			if err := h.userRepo.ClearUserPermissionOverrides(c.Context(), id); err != nil {
				log.Error().Err(err).Msg("Failed to clear permission overrides")
			}
			user.RoleID = roleID
		}
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Password != nil && *req.Password != "" {
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash password")
			return userInternalError(c, "Failed to process password")
		}
		if err := h.userRepo.UpdatePassword(c.Context(), id, hash); err != nil {
			log.Error().Err(err).Msg("Failed to update password")
			return userInternalError(c, "Failed to update password")
		}
	}
	if req.PIN != nil {
		var pinHash *string
		if *req.PIN != "" {
			hash, err := auth.HashPIN(*req.PIN)
			if err != nil {
				log.Error().Err(err).Msg("Failed to hash PIN")
				return userInternalError(c, "Failed to process PIN")
			}
			pinHash = &hash
		}
		if err := h.userRepo.UpdatePIN(c.Context(), id, pinHash); err != nil {
			log.Error().Err(err).Msg("Failed to update PIN")
			return userInternalError(c, "Failed to update PIN")
		}
	}

	if err := h.userRepo.Update(c.Context(), user); err != nil {
		log.Error().Err(err).Msg("Failed to update user")
		return userInternalError(c, "Failed to update user")
	}

	updatedForAudit, _ := h.userRepo.GetByID(c.Context(), id)
	newValues := map[string]interface{}{"affected_user": user.Name, "name": user.Name, "is_active": user.IsActive}
	if updatedForAudit != nil && updatedForAudit.Role != nil {
		newValues["role"] = updatedForAudit.Role.Name
	}
	if user.Email != nil {
		newValues["email"] = *user.Email
	}
	if req.PIN != nil {
		oldValues["pin"] = "[set]"
		if *req.PIN == "" {
			newValues["pin"] = "[removed]"
		} else {
			newValues["pin"] = "[changed]"
		}
	}
	if req.Password != nil && *req.Password != "" {
		oldValues["password"] = "[set]"
		newValues["password"] = "[changed]"
	}
	oldIsActive, ok := oldValues["is_active"].(bool)
	isRestore := req.IsActive != nil && *req.IsActive && ok && !oldIsActive

	action := models.AuditActionUserUpdate
	actionMsg := "Updated user: " + user.Name
	if isRestore {
		action = models.AuditActionUserRestore
		actionMsg = "Restored user: " + user.Name
	}
	audit.LogWithValues(c, action, models.AuditEntityUser, id.String(), actionMsg, oldValues, newValues)

	updatedUser, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch updated user")
		return c.JSON(fiber.Map{"message": "User updated successfully", "user": h.toUserDetailResponse(user)})
	}

	currentUserID := middleware.GetUserID(c)
	if req.RoleID != nil {
		h.broadcastUserEvent(id, EventRoleChanged, currentUserID, map[string]interface{}{"new_role_id": *req.RoleID, "new_role_name": updatedUser.Role.Name})
	} else if req.IsActive != nil {
		if *req.IsActive {
			h.broadcastUserEvent(id, EventUserActivated, currentUserID, nil)
		} else {
			h.broadcastUserEvent(id, EventUserDeactivated, currentUserID, nil)
			if h.eventsHandler != nil {
				h.eventsHandler.DisconnectUser(id)
			}
		}
	} else {
		h.broadcastUserEvent(id, EventUserUpdated, currentUserID, nil)
	}

	return c.JSON(fiber.Map{"message": "User updated successfully", "user": h.toUserDetailResponse(updatedUser)})
}

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

func canAssignRole(currentRole, targetRole string) bool {
	return getRoleLevel(currentRole) >= getRoleLevel(targetRole)
}
