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
	NameExists(context.Context, string, *uuid.UUID) (bool, error)
}

type roleReader interface {
	GetByID(context.Context, uuid.UUID) (*models.Role, error)
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

	userActionCreate userAction = "create"
	userActionEdit   userAction = "edit"
	userActionDelete userAction = "delete"
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

func userArchivedConflict(c *fiber.Ctx, message string) error {
	return middleware.JSONError(c, fiber.StatusConflict, "USER_INACTIVE", message)
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
	if isRoleName(targetRoleName, roleCashier) && (isRoleName(currentRoleName, roleManager) || isRoleName(currentRoleName, roleCashier)) {
		return true
	}

	currentLevel := getRoleLevel(currentRoleName)
	targetLevel := getRoleLevel(targetRoleName)
	if (!isRoleName(currentRoleName, roleOwner) && currentLevel <= targetLevel) || currentLevel < targetLevel {
		_ = userForbidden(c, "You do not have permission to modify a user with the same or higher role level.")
		return false
	}
	return true
}

func (h *UserHandler) requireActionPermission(c *fiber.Ctx, targetRoleName string, action userAction) bool {
	_ = c
	_ = targetRoleName
	_ = action
	return true
}

func (h *UserHandler) enforceTargetUserAction(c *fiber.Ctx, targetRoleName string, action userAction) bool {
	if !h.requireRoleHierarchy(c, targetRoleName) {
		return false
	}
	return h.requireActionPermission(c, targetRoleName, action)
}
