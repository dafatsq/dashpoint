package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/models"
)

// List handles GET /api/v1/users.
func (h *UserHandler) List(c *fiber.Ctx) error {
	page, perPage, isActive, search, role := parseUserPagination(c)
	offset := (page - 1) * perPage

	users, total, err := h.userRepo.ListWithFilter(c.Context(), perPage, offset, isActive, search, role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users")
		return userInternalError(c, "Failed to retrieve users")
	}

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

// ListBasic handles GET /api/v1/users/basic.
func (h *UserHandler) ListBasic(c *fiber.Ctx) error {
	// Get up to 100 active users for dropdowns
	isActive := true
	users, _, err := h.userRepo.ListWithFilter(c.Context(), 100, 0, &isActive, "", "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users")
		return userInternalError(c, "Failed to retrieve users")
	}

	type basicUser struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	responses := make([]basicUser, len(users))
	for i, user := range users {
		responses[i] = basicUser{ID: user.ID.String(), Name: user.Name}
	}

	return c.JSON(fiber.Map{"data": responses})
}

// Get handles GET /api/v1/users/:id.
func (h *UserHandler) Get(c *fiber.Ctx) error {
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

	permissions, err := h.userRepo.GetUserPermissions(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user permissions")
	}

	return c.JSON(fiber.Map{
		"user":        h.toUserDetailResponse(user),
		"permissions": permissions,
	})
}

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
