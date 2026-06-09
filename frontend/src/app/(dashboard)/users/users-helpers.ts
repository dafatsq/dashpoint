import type { CreateUserRequest, User, UserRole } from "@/types";

export const roleHierarchy: Record<UserRole, number> = {
  owner: 3,
  manager: 2,
  cashier: 1,
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

export interface AssignableUserRolesInput {
  currentUser: User | null;
  isOwner: boolean;
  hasPermission: (permission: string) => boolean;
  editingUser?: User | null;
}

export function canEditUser({
  currentUser,
  targetUser,
  canEditUserAny,
  hasPermission,
}: UserEditPolicyInput): boolean {
  if (!currentUser || !canEditUserAny) return false;
  if (targetUser.role_name === "cashier" && hasPermission("manage_users_page")) {
    return true;
  }

  const currentLevel = roleHierarchy[currentUser.role_name] || 0;
  const targetLevel = roleHierarchy[targetUser.role_name] || 0;
  if (currentUser.role_name !== "owner" && currentLevel <= targetLevel) return false;
  if (currentLevel < targetLevel) return false;

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
  if (targetUser.role_name === "cashier" && hasPermission("manage_users_page")) {
    return true;
  }

  const currentLevel = roleHierarchy[currentUser.role_name] || 0;
  const targetLevel = roleHierarchy[targetUser.role_name] || 0;
  if (currentUser.role_name !== "owner" && currentLevel <= targetLevel) return false;
  if (currentLevel < targetLevel) return false;

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
  void editingUser;

  if (roleHierarchy[currentUserRole] >= roleHierarchy.owner) {
    availableRoles.push("owner");
  }

  if (roleHierarchy[currentUserRole] >= roleHierarchy.manager && isOwner) {
    availableRoles.push("manager");
  }

  if (isOwner || hasPermission("manage_users_page")) {
    availableRoles.push("cashier");
  }

  return availableRoles;
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
