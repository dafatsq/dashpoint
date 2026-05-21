package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

type userRepository interface {
	GetByID(context.Context, uuid.UUID) (*models.User, error)
	ListWithFilter(context.Context, int, int, *bool, string, string) ([]*models.User, int, error)
	Create(context.Context, *models.User) error
	Update(context.Context, *models.User) error
	UpdatePassword(context.Context, uuid.UUID, string) error
	UpdatePIN(context.Context, uuid.UUID, *string) error
	Deactivate(context.Context, uuid.UUID) error
	PermanentDelete(context.Context, uuid.UUID) error
	HasSalesHistory(context.Context, uuid.UUID) (bool, error)
	HasExpenseHistory(context.Context, uuid.UUID) (bool, error)
	EmailExists(context.Context, string, *uuid.UUID) (bool, error)
	GetUserPermissions(context.Context, uuid.UUID) ([]string, error)
	GetUserPermissionOverrides(context.Context, uuid.UUID) ([]*models.UserPermission, error)
	ClearUserPermissionOverrides(context.Context, uuid.UUID) error
	SetUserPermission(context.Context, uuid.UUID, uuid.UUID, bool, *uuid.UUID) error
	NameExists(context.Context, string, *uuid.UUID) (bool, error)
}

type roleReader interface {
	GetByID(context.Context, uuid.UUID) (*models.Role, error)
}

type permissionReader interface {
	GetByID(context.Context, uuid.UUID) (*models.Permission, error)
}

type userEventBroadcaster interface {
	BroadcastToUser(uuid.UUID, UserEvent)
	DisconnectUser(uuid.UUID)
}

type userAction string

const (
	roleOwner   = "owner"
	roleManager = "manager"
	roleCashier = "cashier"

	userActionCreate            userAction = "create"
	userActionEdit              userAction = "edit"
	userActionDelete            userAction = "delete"
	userActionManagePermissions userAction = "manage_permissions"
)

func badUserRequest(c *fiber.Ctx, code, message string) error {
	return middleware.JSONError(c, fiber.StatusBadRequest, code, message)
}

func userNotFound(c *fiber.Ctx) error {
	return middleware.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "User not found")
}

func userInternalError(c *fiber.Ctx, message string) error {
	return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func userForbidden(c *fiber.Ctx, message string) error {
	return middleware.JSONError(c, fiber.StatusForbidden, "FORBIDDEN", message)
}

func userConflict(c *fiber.Ctx, code, message string) error {
	return middleware.JSONError(c, fiber.StatusConflict, code, message)
}

func parseUserIDParam(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, badUserRequest(c, "INVALID_ID", "Invalid user ID format")
	}
	return id, nil
}

func parseUserPagination(c *fiber.Ctx) (int, int, *bool, string, string) {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	activeOnlyStr := c.Query("active_only", "")
	var isActive *bool
	if activeOnlyStr == "true" {
		active := true
		isActive = &active
	} else if activeOnlyStr == "false" {
		active := false
		isActive = &active
	}

	return page, perPage, isActive, strings.TrimSpace(c.Query("search", "")), strings.TrimSpace(c.Query("role", ""))
}

func (h *UserHandler) currentUserPermissionSet(c *fiber.Ctx) (map[string]bool, error) {
	permissions, err := h.userRepo.GetUserPermissions(c.Context(), middleware.GetUserID(c))
	if err != nil {
		return nil, err
	}

	permissionSet := make(map[string]bool, len(permissions))
	for _, permission := range permissions {
		permissionSet[permission] = true
	}

	return permissionSet, nil
}

func roleNameOfUser(user *models.User) string {
	if user != nil && user.Role != nil {
		return user.Role.Name
	}
	return ""
}

func userNameOrUnknown(user *models.User) string {
	if user != nil && user.Name != "" {
		return user.Name
	}
	return "Unknown"
}

func baseUserAuditValues(user *models.User) map[string]interface{} {
	values := map[string]interface{}{
		"affected_user": user.Name,
		"name":          user.Name,
	}
	if user.Role != nil {
		values["role"] = user.Role.Name
	}
	if user.Email != nil {
		values["email"] = *user.Email
	}
	return values
}

func isRoleName(roleName, expected string) bool {
	return strings.EqualFold(roleName, expected)
}

func (h *UserHandler) requireRoleHierarchy(c *fiber.Ctx, targetRoleName string) bool {
	currentRoleName := middleware.GetRoleName(c)
	if getRoleLevel(currentRoleName) < getRoleLevel(targetRoleName) {
		_ = userForbidden(c, "You do not have permission to modify a user with a higher role level.")
		return false
	}
	return true
}

func (h *UserHandler) requireActionPermission(c *fiber.Ctx, targetRoleName string, action userAction) bool {
	if !isRoleName(middleware.GetRoleName(c), roleManager) {
		return true
	}

	permissionSet, err := h.currentUserPermissionSet(c)
	if err != nil {
		_ = userInternalError(c, "Failed to validate permissions")
		return false
	}

	permissionKey, message, ok := managerPermissionForAction(action, targetRoleName)
	if !ok {
		return true
	}
	if permissionSet[permissionKey] {
		return true
	}

	_ = userForbidden(c, message)
	return false
}

func managerPermissionForAction(action userAction, targetRoleName string) (string, string, bool) {
	switch {
	case isRoleName(targetRoleName, roleManager):
		switch action {
		case userActionCreate:
			return "can_create_manager_users", "You do not have permission to create Managers", true
		case userActionEdit:
			return "can_edit_manager_users", "You do not have permission to edit details of Managers", true
		case userActionDelete:
			return "can_delete_manager_users", "You do not have permission to archive/delete Managers", true
		case userActionManagePermissions:
			return "can_manage_manager_permissions", "You do not have permission to manage Managers", true
		}
	case isRoleName(targetRoleName, roleCashier):
		switch action {
		case userActionCreate:
			return "can_create_cashier_users", "You do not have permission to create Cashiers", true
		case userActionEdit:
			return "can_edit_cashier_users", "You do not have permission to edit details of Cashiers", true
		case userActionDelete:
			return "can_delete_cashier_users", "You do not have permission to archive/delete Cashiers", true
		case userActionManagePermissions:
			return "can_manage_cashier_permissions", "You do not have permission to manage Cashiers", true
		}
	}

	return "", "", false
}

func (h *UserHandler) enforceTargetUserAction(c *fiber.Ctx, targetRoleName string, action userAction) bool {
	if !h.requireRoleHierarchy(c, targetRoleName) {
		return false
	}
	return h.requireActionPermission(c, targetRoleName, action)
}
