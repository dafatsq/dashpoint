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
	userRepo         userRepository
	roleRepo         roleReader
	eventsHandler    userEventBroadcaster
	refreshTokenRepo userRefreshTokenRevoker
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo, roleRepo: roleRepo}
}

// SetEventsHandler sets the events handler for broadcasting user updates.
func (h *UserHandler) SetEventsHandler(eventsHandler userEventBroadcaster) {
	h.eventsHandler = eventsHandler
}

// SetRefreshTokenRevoker sets the repository used to invalidate user sessions after credential changes.
func (h *UserHandler) SetRefreshTokenRevoker(refreshTokenRepo userRefreshTokenRevoker) {
	h.refreshTokenRepo = refreshTokenRepo
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

func (h *UserHandler) revokeUserRefreshTokens(c *fiber.Ctx, userID uuid.UUID, reason string) error {
	if h.refreshTokenRepo == nil {
		return nil
	}
	if err := h.refreshTokenRepo.RevokeAllForUser(c.Context(), userID, reason); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to revoke user refresh tokens")
		return userInternalError(c, "Failed to revoke user sessions")
	}
	return nil
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
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type CreateUserRequest struct {
	Email    *string `json:"email"`
	Name     string  `json:"name"`
	Password *string `json:"password"`
	PIN      *string `json:"pin"`
	RoleID   string  `json:"role_id"`
}

type UpdateUserRequest struct {
	Email             *string `json:"email"`
	Name              *string `json:"name"`
	RoleID            *string `json:"role_id"`
	IsActive          *bool   `json:"is_active"`
	PIN               *string `json:"pin"`
	Password          *string `json:"password"`
	CurrentPassword   *string `json:"current_password"`
	CurrentPIN        *string `json:"current_pin"`
	ExpectedUpdatedAt *string `json:"expected_updated_at"`
}

type UpdatePasswordRequest struct {
	Password          string  `json:"password"`
	CurrentPassword   *string `json:"current_password"`
	CurrentPIN        *string `json:"current_pin"`
	ExpectedUpdatedAt *string `json:"expected_updated_at"`
}

type UpdatePINRequest struct {
	PIN               *string `json:"pin"`
	CurrentPassword   *string `json:"current_password"`
	CurrentPIN        *string `json:"current_pin"`
	ExpectedUpdatedAt *string `json:"expected_updated_at"`
}

// Create handles POST /api/v1/users.
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := parseStrictUserJSON(c, &req); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.RoleID = strings.TrimSpace(req.RoleID)
	if req.Email != nil {
		trimmed := strings.TrimSpace(*req.Email)
		req.Email = &trimmed
	}
	if req.PIN != nil {
		trimmed := strings.TrimSpace(*req.PIN)
		req.PIN = &trimmed
	}
	if message := validateUserName(req.Name, true); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
	}
	if req.RoleID == "" {
		return badUserRequest(c, "VALIDATION_ERROR", "Role ID is required")
	}
	if message := validateUserEmail(req.Email, true); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
	}
	if message := validateUserPassword(req.Password, true); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
	}
	if message := validateUserPIN(req.PIN, true); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
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
	if exists, err := h.userRepo.NameExists(c.Context(), req.Name, nil); err != nil {
		log.Error().Err(err).Msg("Failed to check name")
		return userInternalError(c, "Failed to validate name")
	} else if exists {
		return userConflict(c, "NAME_EXISTS", "A user with this name already exists. Please use an initial (e.g., 'John S.')")
	}
	if !h.requireActionPermission(c, role.Name, userActionCreate) {
		return nil
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

	if createdUser, _ := h.userRepo.GetByID(c.Context(), user.ID); createdUser != nil {
		user = createdUser
	}
	newValues := baseUserAuditValues(user)
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

	var req UpdateUserRequest
	if err := parseStrictUserJSON(c, &req); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}
	if ok, err := requireExpectedUpdatedAt(c, req.ExpectedUpdatedAt, user.UpdatedAt); !ok {
		return err
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		req.Name = &trimmed
		if message := validateUserName(trimmed, false); message != "" {
			return badUserRequest(c, "VALIDATION_ERROR", message)
		}
	}
	if req.Email != nil {
		trimmed := strings.TrimSpace(*req.Email)
		req.Email = &trimmed
		if message := validateUserEmail(req.Email, false); message != "" {
			return badUserRequest(c, "VALIDATION_ERROR", message)
		}
	}
	if req.RoleID != nil {
		trimmed := strings.TrimSpace(*req.RoleID)
		req.RoleID = &trimmed
	}
	if message := validateUserPassword(req.Password, false); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
	}
	if req.PIN != nil {
		trimmed := strings.TrimSpace(*req.PIN)
		req.PIN = &trimmed
		if message := validateUserPIN(req.PIN, false); message != "" {
			return badUserRequest(c, "VALIDATION_ERROR", message)
		}
	}
	if !user.IsActive && (req.IsActive == nil || !*req.IsActive) {
		return userArchivedConflict(c, "Archived users cannot be changed")
	}

	isSelfUpdate := middleware.GetUserID(c) == id
	if isSelfUpdate {
		if req.RoleID != nil {
			roleID, err := uuid.Parse(*req.RoleID)
			if err != nil {
				return badUserRequest(c, "INVALID_ROLE_ID", "Invalid role ID format")
			}
			if roleID != user.RoleID {
				return middleware.JSONError(c, fiber.StatusForbidden, "CANNOT_CHANGE_OWN_ROLE", "You cannot change your own role")
			}
		}
		if req.IsActive != nil && *req.IsActive != user.IsActive {
			return middleware.JSONError(c, fiber.StatusForbidden, "CANNOT_CHANGE_OWN_STATUS", "You cannot change your own account status")
		}
	} else {
		targetRoleName := roleNameOfUser(user)
		if !h.enforceTargetUserAction(c, targetRoleName, userActionEdit) {
			return nil
		}
	}

	oldValues := baseUserAuditValues(user)
	oldValues["is_active"] = user.IsActive

	if req.Name != nil {
		if *req.Name != "" {
			if exists, err := h.userRepo.NameExists(c.Context(), *req.Name, &id); err != nil {
				log.Error().Err(err).Msg("Failed to check name")
				return userInternalError(c, "Failed to validate name")
			} else if exists {
				return userConflict(c, "NAME_EXISTS", "A user with this name already exists. Please use an initial (e.g., 'John S.')")
			}
		}
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
			user.RoleID = roleID
		}
	}
	// A user changing their own credentials must prove knowledge of an
	// existing one; admins resetting others' credentials stay permission-gated.
	if middleware.GetUserID(c) == id && ((req.Password != nil && *req.Password != "") || req.PIN != nil) {
		if ok, err := verifySelfCredentialProof(c, user, req.CurrentPassword, req.CurrentPIN); !ok {
			return err
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
	if credentialsChangedByUpdate(req) {
		if err := h.revokeUserRefreshTokens(c, id, "user_credentials_changed"); err != nil {
			return err
		}
	}

	updatedForAudit, _ := h.userRepo.GetByID(c.Context(), id)
	newValues := baseUserAuditValues(user)
	newValues["is_active"] = user.IsActive
	if updatedForAudit != nil && updatedForAudit.Role != nil {
		newValues["role"] = updatedForAudit.Role.Name
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

func credentialsChangedByUpdate(req UpdateUserRequest) bool {
	if req.Email != nil {
		return true
	}
	if req.Password != nil && *req.Password != "" {
		return true
	}
	return req.PIN != nil
}

func getRoleLevel(roleName string) int {
	switch {
	case isRoleName(roleName, roleOwner):
		return 3
	case isRoleName(roleName, roleManager):
		return 2
	case isRoleName(roleName, roleCashier):
		return 1
	default:
		return 0
	}
}

func canAssignRole(currentRole, targetRole string) bool {
	if isRoleName(targetRole, roleCashier) && (isRoleName(currentRole, roleManager) || isRoleName(currentRole, roleCashier)) {
		return true
	}

	currentLevel := getRoleLevel(currentRole)
	targetLevel := getRoleLevel(targetRole)
	if isRoleName(currentRole, roleOwner) {
		return currentLevel >= targetLevel
	}
	return currentLevel > targetLevel
}
