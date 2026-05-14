package handlers

import (
	"dashpoint/backend/internal/repository"
)

// RoleHandler handles role and permission endpoints
type RoleHandler struct {
	roleRepo       roleEndpointReader
	permissionRepo rolePermissionEndpointReader
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(
	roleRepo *repository.RoleRepository,
	permissionRepo *repository.PermissionRepository,
) *RoleHandler {
	return &RoleHandler{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	ID          string  `json:"id"`
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Category    string  `json:"category"`
}

// RoleWithPermissionsResponse represents a role with its permissions
type RoleWithPermissionsResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	Permissions []PermissionResponse `json:"permissions"`
}
