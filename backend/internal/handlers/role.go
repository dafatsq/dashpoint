package handlers

import (
	"dashpoint/backend/internal/repository"
)

// RoleHandler handles role endpoints.
type RoleHandler struct {
	roleRepo roleEndpointReader
}

// NewRoleHandler creates a new role handler.
func NewRoleHandler(roleRepo *repository.RoleRepository) *RoleHandler {
	return &RoleHandler{
		roleRepo: roleRepo,
	}
}

// RoleDetailResponse represents a role with its derived capabilities.
type RoleDetailResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions"`
}

type UpdateRolePermissionsRequest struct {
	Permissions         []string `json:"permissions"`
	ExpectedPermissions []string `json:"expected_permissions"`
}
