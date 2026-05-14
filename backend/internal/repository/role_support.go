package repository

import "dashpoint/backend/internal/models"

func roleSortWeight(name string) int {
	switch name {
	case "owner":
		return 1
	case "manager":
		return 2
	case "cashier":
		return 3
	default:
		return 4
	}
}

func groupPermissionsByCategory(permissions []*models.Permission) map[string][]*models.Permission {
	grouped := make(map[string][]*models.Permission)
	for _, permission := range permissions {
		grouped[permission.Category] = append(grouped[permission.Category], permission)
	}
	return grouped
}
