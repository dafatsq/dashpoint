import type {
  CreateUserRequest,
  Permission,
  PermissionOverride,
  User,
  UserRole,
} from "@/types";

export const roleHierarchy: Record<UserRole, number> = {
  owner: 3,
  manager: 2,
  cashier: 1,
};

export const CATEGORY_ORDER = [
  "pos",
  "sales",
  "categories",
  "inventory",
  "reports",
  "expenses",
  "users",
  "system",
] as const;

export const HIDDEN_PERMS = [
  "can_manage_manager_permissions",
  "can_manage_cashier_permissions",
  "can_create_manager_users",
  "can_create_cashier_users",
  "can_edit_manager_users",
  "can_edit_cashier_users",
  "can_delete_manager_users",
  "can_delete_cashier_users",
  "can_manage_expenses",
  "can_manage_categories",
  "can_edit_inventory",
];

export const REPLACEMENT_PARENT_TOGGLES: Record<
  string,
  Array<{ label: string; description: string }>
> = {};

export const USER_PERMISSION_CASCADE: Record<string, string[]> = {
  can_create_user: ["can_create_manager_users", "can_create_cashier_users"],
  can_edit_user: ["can_edit_manager_users", "can_edit_cashier_users"],
  can_delete_user: ["can_delete_manager_users", "can_delete_cashier_users"],
  can_manage_permissions: [
    "can_manage_manager_permissions",
    "can_manage_cashier_permissions",
  ],
};

const MANAGER_USER_ACTION_KEYS = new Set([
  "can_create_manager_users",
  "can_create_cashier_users",
  "can_edit_manager_users",
  "can_edit_cashier_users",
  "can_delete_manager_users",
  "can_delete_cashier_users",
  "can_manage_manager_permissions",
  "can_manage_cashier_permissions",
]);

const FEATURE_NAME_BY_RESOURCE: Record<string, string> = {
  sale: "Sales",
  sales: "Sales",
  user: "Users",
  users: "Users",
  manager_users: "Manager Users",
  cashier_users: "Cashier Users",
  inventory: "Inventory",
  categories: "Categories",
  expenses: "Expenses",
  reports: "Reports",
  permissions: "Permissions",
  audit_logs: "Audit Logs",
};

const FEATURE_NAME_BY_CATEGORY: Record<string, string> = {
  pos: "Sales",
  sales: "Sales",
  system: "Audit",
};

export interface UserActionPolicyInput {
  currentUser: User | null;
  targetUser: User;
  hasPermission: (permission: string) => boolean;
}

export interface UserEditPolicyInput extends UserActionPolicyInput {
  canEditUserAny: boolean;
}

export interface UserDeletePolicyInput extends UserActionPolicyInput {
  canDeleteUserAny: boolean;
}

export interface UserPermissionPolicyInput extends UserActionPolicyInput {
  canManagePermissions: boolean;
}

export interface AssignableUserRolesInput {
  currentUser: User | null;
  isOwner: boolean;
  hasPermission: (permission: string) => boolean;
  editingUser?: User | null;
}

export interface PermissionChangeSetInput {
  permission: Permission;
  enabled: boolean;
  allPermissions: Record<string, Permission[]>;
  permissionChanges: Record<string, boolean | null>;
  userOverrides: PermissionOverride[];
  userEffectivePermissions: string[];
  currentUser: User | null;
  targetUser: User | null;
}

export function toTitleCase(value: string): string {
  return value
    .split(" ")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export function getPermissionFeatureName(
  category: string,
  resource?: string,
): string {
  if (resource) {
    const normalizedResource = resource.toLowerCase();
    if (FEATURE_NAME_BY_RESOURCE[normalizedResource]) {
      return FEATURE_NAME_BY_RESOURCE[normalizedResource];
    }

    return toTitleCase(normalizedResource.replace(/_/g, " "));
  }

  if (FEATURE_NAME_BY_CATEGORY[category]) {
    return FEATURE_NAME_BY_CATEGORY[category];
  }

  return toTitleCase(category.replace(/_/g, " "));
}

export function normalizeDeleteArchiveWording(text: string): string {
  return text
    .replace(/Delete\/Archive/g, "Delete/archive")
    .replace(/delete\/archive/g, "delete/archive")
    .replace(/archive\/delete/g, "delete/archive")
    .replace(/archive or delete/gi, "delete/archive")
    .replace(/delete or archive/gi, "delete/archive")
    .replace(/archive and delete/gi, "delete/archive")
    .replace(/delete and archive/gi, "delete/archive");
}

export function canEditUser({
  currentUser,
  targetUser,
  canEditUserAny,
  hasPermission,
}: UserEditPolicyInput): boolean {
  if (!currentUser || !canEditUserAny) return false;

  const currentLevel = roleHierarchy[currentUser.role_name] || 0;
  const targetLevel = roleHierarchy[targetUser.role_name] || 0;
  if (currentLevel < targetLevel) return false;

  if (currentUser.role_name === "manager") {
    if (
      targetUser.role_name === "manager" &&
      !hasPermission("can_edit_manager_users")
    ) {
      return false;
    }
    if (
      targetUser.role_name === "cashier" &&
      !hasPermission("can_edit_cashier_users")
    ) {
      return false;
    }
  }

  return true;
}

export function canDeleteUser({
  currentUser,
  targetUser,
  canDeleteUserAny,
  hasPermission,
}: UserDeletePolicyInput): boolean {
  if (!currentUser || !canDeleteUserAny) return false;
  if (targetUser.id === currentUser.id) return false;

  const currentLevel = roleHierarchy[currentUser.role_name] || 0;
  const targetLevel = roleHierarchy[targetUser.role_name] || 0;
  if (currentLevel < targetLevel) return false;

  if (currentUser.role_name === "manager") {
    if (
      targetUser.role_name === "manager" &&
      !hasPermission("can_delete_manager_users")
    ) {
      return false;
    }
    if (
      targetUser.role_name === "cashier" &&
      !hasPermission("can_delete_cashier_users")
    ) {
      return false;
    }
  }

  return true;
}

export function canManageUserPermissions({
  currentUser,
  targetUser,
  canManagePermissions,
  hasPermission,
}: UserPermissionPolicyInput): boolean {
  if (!currentUser || !canManagePermissions) return false;
  if (targetUser.id === currentUser.id) return false;

  const currentLevel = roleHierarchy[currentUser.role_name] || 0;
  const targetLevel = roleHierarchy[targetUser.role_name] || 0;
  if (currentLevel < targetLevel) return false;

  if (currentUser.role_name === "manager") {
    if (
      targetUser.role_name === "manager" &&
      !hasPermission("can_manage_manager_permissions")
    ) {
      return false;
    }
    if (
      targetUser.role_name === "cashier" &&
      !hasPermission("can_manage_cashier_permissions")
    ) {
      return false;
    }
  }

  return true;
}

export function getAssignableUserRoles({
  currentUser,
  isOwner,
  hasPermission,
  editingUser = null,
}: AssignableUserRolesInput): UserRole[] {
  const availableRoles: UserRole[] = [];
  const currentUserRole = (currentUser?.role_name || "cashier") as UserRole;
  const isEditing = Boolean(editingUser);

  if (roleHierarchy[currentUserRole] >= roleHierarchy.owner) {
    availableRoles.push("owner");
  }

  if (
    roleHierarchy[currentUserRole] >= roleHierarchy.manager &&
    (isOwner ||
      hasPermission(
        isEditing ? "can_edit_manager_users" : "can_create_manager_users",
      ))
  ) {
    availableRoles.push("manager");
  }

  if (
    isOwner ||
    currentUser?.role_name === "cashier" ||
    hasPermission(
      isEditing ? "can_edit_cashier_users" : "can_create_cashier_users",
    )
  ) {
    availableRoles.push("cashier");
  }

  return availableRoles;
}

export function getPermissionOverride(
  permissionId: string,
  overrides: PermissionOverride[],
): PermissionOverride | undefined {
  return overrides.find((override) => override.permission_id === permissionId);
}

export function isViewPermission(permission: Permission): boolean {
  if (permission.key === "can_create_sale") return false;
  if (permission.key === "can_view_sales") return true;
  if (permission.key === "can_void_sale") return false;
  return permission.key.startsWith("can_view_");
}

export function getViewPermissionForCategory(
  category: string,
  permissions: Permission[],
): Permission | undefined {
  if (category === "sales") {
    return permissions.find((permission) => permission.key === "can_view_sales");
  }
  return permissions.find((permission) => isViewPermission(permission));
}

export function isPermissionEnabled(
  permission: Permission,
  permissionChanges: Record<string, boolean | null>,
  userOverrides: PermissionOverride[],
  userEffectivePermissions: string[],
): boolean {
  if (permissionChanges[permission.id] !== undefined) {
    return permissionChanges[permission.id] === true;
  }

  const override = getPermissionOverride(permission.id, userOverrides);
  if (override) {
    return override.allowed;
  }

  return userEffectivePermissions.includes(permission.key);
}

export function currentUserCanGrantPermission(
  permission: Permission,
  currentUser: User | null,
  targetUser: User | null,
): boolean {
  if (!currentUser || !targetUser) return false;
  if (currentUser.role_name === "owner") return true;

  const userPermissions = currentUser.permissions || [];
  const targetRole = targetUser.role_name;
  const isUserPermission = [
    "can_create_user",
    "can_edit_user",
    "can_delete_user",
    "can_manage_permissions",
  ].includes(permission.key);

  if (isUserPermission) {
    if (targetRole === "cashier") {
      if (permission.key === "can_create_user") {
        return userPermissions.includes("can_create_cashier_users");
      }
      if (permission.key === "can_edit_user") {
        return userPermissions.includes("can_edit_cashier_users");
      }
      if (permission.key === "can_delete_user") {
        return userPermissions.includes("can_delete_cashier_users");
      }
      if (permission.key === "can_manage_permissions") {
        return userPermissions.includes("can_manage_cashier_permissions");
      }
    }

    if (targetRole === "manager") {
      if (permission.key === "can_create_user") {
        return userPermissions.includes("can_create_manager_users");
      }
      if (permission.key === "can_edit_user") {
        return userPermissions.includes("can_edit_manager_users");
      }
      if (permission.key === "can_delete_user") {
        return userPermissions.includes("can_delete_manager_users");
      }
      if (permission.key === "can_manage_permissions") {
        return userPermissions.includes("can_manage_manager_permissions");
      }
    }

    return false;
  }

  return userPermissions.includes(permission.key);
}

export function createPermissionChangeSet({
  permission,
  enabled,
  allPermissions,
  permissionChanges,
  userOverrides,
  userEffectivePermissions,
  currentUser,
  targetUser,
}: PermissionChangeSetInput): Record<string, boolean | null> {
  if (!currentUserCanGrantPermission(permission, currentUser, targetUser)) {
    return permissionChanges;
  }

  const nextChanges = { ...permissionChanges };
  const setChange = (nextPermission: Permission, value: boolean) => {
    const override = getPermissionOverride(nextPermission.id, userOverrides);
    const initialState =
      override !== undefined
        ? override.allowed
        : userEffectivePermissions.includes(nextPermission.key);

    if (value === initialState) {
      delete nextChanges[nextPermission.id];
    } else {
      nextChanges[nextPermission.id] = value;
    }
  };

  const trySetChange = (nextPermission: Permission, value: boolean) => {
    if (
      value &&
      !currentUserCanGrantPermission(nextPermission, currentUser, targetUser)
    ) {
      return;
    }
    setChange(nextPermission, value);
  };

  const cascadeHiddenChange = (nextPermission: Permission, value: boolean) => {
    trySetChange(nextPermission, value);
  };

  setChange(permission, enabled);

  const userManagementCascade: Record<string, string[]> = {
    can_create_user: ["can_create_manager_users", "can_create_cashier_users"],
    can_edit_user: ["can_edit_manager_users", "can_edit_cashier_users"],
    can_delete_user: ["can_delete_manager_users", "can_delete_cashier_users"],
    can_manage_permissions: [
      "can_manage_manager_permissions",
      "can_manage_cashier_permissions",
    ],
  };

  if (!enabled) {
    for (const [category, permissions] of Object.entries(allPermissions)) {
      const viewPermission = getViewPermissionForCategory(category, permissions);
      if (!viewPermission || viewPermission.id !== permission.id) continue;

      for (const childPermission of permissions) {
        if (childPermission.id === permission.id) continue;
        if (category === "sales" && childPermission.key !== "can_void_sale") {
          continue;
        }

        if (HIDDEN_PERMS.includes(childPermission.key)) {
          cascadeHiddenChange(childPermission, false);
        } else {
          trySetChange(childPermission, false);
        }
      }
      break;
    }

    const childKeysToDisable = userManagementCascade[permission.key];
    if (childKeysToDisable) {
      for (const childPermission of allPermissions.users || []) {
        if (childKeysToDisable.includes(childPermission.key)) {
          cascadeHiddenChange(childPermission, false);
        }
      }
    }
  }

  return nextChanges;
}

export function isPermissionDisabledByParent(
  permission: Permission,
  category: string,
  allPermissions: Record<string, Permission[]>,
  permissionChanges: Record<string, boolean | null>,
  userOverrides: PermissionOverride[],
  userEffectivePermissions: string[],
): boolean {
  if (permission.key === "can_void_sale") {
    const viewPermission = (allPermissions[category] || []).find(
      (entry) => entry.key === "can_view_sales",
    );
    if (
      viewPermission &&
      !isPermissionEnabled(
        viewPermission,
        permissionChanges,
        userOverrides,
        userEffectivePermissions,
      )
    ) {
      return true;
    }
  }

  if (category !== "sales") {
    const viewPermission = getViewPermissionForCategory(
      category,
      allPermissions[category] || [],
    );
    if (
      viewPermission &&
      !isViewPermission(permission) &&
      !isPermissionEnabled(
        viewPermission,
        permissionChanges,
        userOverrides,
        userEffectivePermissions,
      )
    ) {
      return true;
    }
  }

  return false;
}

export function sortPermissions(permissions: Permission[]): Permission[] {
  return [...permissions].sort((left, right) => {
    const leftIsView = isViewPermission(left);
    const rightIsView = isViewPermission(right);
    if (leftIsView && !rightIsView) return -1;
    if (!leftIsView && rightIsView) return 1;
    return 0;
  });
}

export function getPermissionDisplayName(
  permission: Permission,
  category: string,
): string {
  const crudMatch = permission.key.match(/^can_(create|view|edit|delete)_(.+)$/);
  if (crudMatch) {
    const [, action, resource] = crudMatch;
    const feature = getPermissionFeatureName(category, resource);

    if (action === "create") return `Create ${feature}`;
    if (action === "view") return `Access to ${feature}`;
    if (action === "edit") return `Edit ${feature}`;
    if (category === "expenses") return "Delete Expense";
    return `Delete/archive ${feature}`;
  }

  if (isViewPermission(permission)) {
    return `Access to ${getPermissionFeatureName(category)}`;
  }

  return normalizeDeleteArchiveWording(permission.name);
}

export function getVisibleChangesCount(
  permissionChanges: Record<string, boolean | null>,
  allPermissions: Record<string, Permission[]>,
  permissionsUser: User | null,
): number {
  return Object.keys(permissionChanges).filter((id) => {
    const permission = Object.values(allPermissions)
      .flat()
      .find((entry) => entry.id === id);
    if (!permission) return true;

    if (permissionsUser?.role_name === "manager") {
      if (MANAGER_USER_ACTION_KEYS.has(permission.key)) return true;
    }

    return !HIDDEN_PERMS.includes(permission.key);
  }).length;
}

export function hasUserFormChanges(
  formData: CreateUserRequest,
  editingUser: User | null,
): boolean {
  if (!editingUser) return true;

  return (
    formData.email !== (editingUser.email || "") ||
    formData.name !== editingUser.name ||
    formData.role !== editingUser.role_name ||
    formData.password !== "" ||
    formData.pin !== ""
  );
}

export function findRoleIdByName(
  roles: Array<{ id: string; name: string }>,
  roleName: string | undefined,
): string | undefined {
  return roles.find((role) => role.name === roleName)?.id;
}
