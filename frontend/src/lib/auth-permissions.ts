import type { User } from "@/types";

export const PERMISSIONS = {
  USERS_VIEW: "access_users_page",
  USERS_CREATE: "manage_users_page",
  USERS_EDIT: "manage_users_page",
  USERS_DELETE: "manage_users_page",
  PRODUCTS_VIEW: "access_products_page",
  PRODUCTS_CREATE: "manage_products_page",
  PRODUCTS_EDIT: "manage_products_page",
  PRODUCTS_DELETE: "manage_products_page",
  INVENTORY_VIEW: "access_inventory_page",
  INVENTORY_EDIT: "manage_inventory_page",
  INVENTORY_ADD_STOCK: "manage_inventory_page",
  INVENTORY_REMOVE_STOCK: "manage_inventory_page",
  INVENTORY_ADJUST_STOCK: "manage_inventory_page",
  SALES_CREATE: "manage_pos_page",
  SALES_VIEW: "access_sales_page",
  SALES_VOID: "manage_sales_page",
  REPORTS_VIEW: "access_reports_page",
  REPORTS_EXPORT: "manage_reports_page",
  EXPENSES_VIEW: "access_expenses_page",
  EXPENSES_MANAGE: "manage_expenses_page",
  EXPENSES_CREATE: "manage_expenses_page",
  EXPENSES_EDIT: "manage_expenses_page",
  EXPENSES_DELETE: "manage_expenses_page",
  AUDIT_VIEW: "access_audit_page",
  POS_VIEW: "access_pos_page",
  POS_SHIFT_START: "manage_shifts_page",
  POS_SHIFT_END: "manage_shifts_page",
  CATEGORIES_VIEW: "access_categories_page",
  CATEGORIES_MANAGE: "manage_categories_page",
  CATEGORIES_CREATE: "manage_categories_page",
  CATEGORIES_EDIT: "manage_categories_page",
  CATEGORIES_DELETE: "manage_categories_page",
  SHIFTS_VIEW: "access_shifts_page",
  SHIFTS_MANAGE: "manage_shifts_page",
  CHANGES_VIEW: "access_changes_page",
} as const;

export function hasPermission(user: User | null, permission: string): boolean {
  if (!user) return false;
  if (user.role_name === "owner") return true;

  return user.permissions?.includes(permission) ?? false;
}

export function hasAnyPermission(
  user: User | null,
  permissions: string[],
): boolean {
  if (!user) return false;
  if (user.role_name === "owner") return true;

  return permissions.some((permission) => user.permissions?.includes(permission));
}

export function hasAllPermissions(
  user: User | null,
  permissions: string[],
): boolean {
  if (!user) return false;
  if (user.role_name === "owner") return true;

  return permissions.every((permission) => user.permissions?.includes(permission));
}
