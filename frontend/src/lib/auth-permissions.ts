import type { User } from "@/types";

export const PERMISSIONS = {
  USERS_VIEW: "can_view_users",
  USERS_CREATE: "can_create_user",
  USERS_EDIT: "can_edit_user",
  USERS_DELETE: "can_delete_user",
  USERS_PERMISSIONS: "can_manage_permissions",
  PRODUCTS_VIEW: "can_view_products",
  PRODUCTS_CREATE: "can_create_product",
  PRODUCTS_EDIT: "can_edit_product",
  PRODUCTS_DELETE: "can_delete_product",
  INVENTORY_VIEW: "can_view_inventory",
  INVENTORY_EDIT: "can_edit_inventory",
  INVENTORY_ADD_STOCK: "can_add_stock",
  INVENTORY_REMOVE_STOCK: "can_remove_stock",
  INVENTORY_ADJUST_STOCK: "can_adjust_stock",
  SALES_CREATE: "can_create_sale",
  SALES_VIEW: "can_view_sales",
  SALES_VOID: "can_void_sale",
  REPORTS_VIEW: "can_view_reports",
  REPORTS_EXPORT: "can_export_data",
  EXPENSES_VIEW: "can_view_expenses",
  EXPENSES_MANAGE: "can_manage_expenses",
  EXPENSES_CREATE: "can_create_expenses",
  EXPENSES_EDIT: "can_edit_expenses",
  EXPENSES_DELETE: "can_delete_expenses",
  AUDIT_VIEW: "can_view_audit_logs",
  POS_VIEW: "can_view_pos",
  POS_SHIFT_START: "can_start_shift",
  POS_SHIFT_END: "can_end_shift",
  CATEGORIES_VIEW: "can_view_categories",
  CATEGORIES_MANAGE: "can_manage_categories",
  CATEGORIES_CREATE: "can_create_categories",
  CATEGORIES_EDIT: "can_edit_categories",
  CATEGORIES_DELETE: "can_delete_categories",
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
