import type { AuditLog, Shift } from "@/types";

export interface DashboardStats {
  todaySales: number;
  todayTransactions: number;
  averageSale: number;
  lowStockCount: number;
}

export const DASHBOARD_ACTION_LABELS: Record<string, string> = {
  create: "Created",
  update: "Updated",
  delete: "Deleted",
  void: "Voided",
  adjust: "Adjusted",
  count: "Counted",
  threshold_update: "Updated",
  permission_change: "Updated",
  start: "Started",
  close: "Closed",
  archive: "Archived",
  restore: "Restored",
};

const DASHBOARD_SKIP_FIELDS = new Set([
  "affected_product",
  "affected_category",
  "affected_expense",
  "affected_user",
  "affected_role",
  "product_name",
  "invoice_no",
  "category_id",
  "product_id",
]);

const DASHBOARD_PERMISSION_LABELS: Record<string, string> = {
  access_pos_page: "Access POS",
  manage_pos_page: "Process Sales",
  access_products_page: "Access Products",
  manage_products_page: "Manage Products",
  access_inventory_page: "Access Inventory",
  manage_inventory_page: "Manage Inventory",
  access_sales_page: "Access Sales",
  manage_sales_page: "Void Sales",
  access_reports_page: "Access Reports",
  manage_reports_page: "Export Reports",
  access_expenses_page: "Access Expenses",
  manage_expenses_page: "Manage Expenses",
  access_categories_page: "Access Categories",
  manage_categories_page: "Manage Categories",
  access_users_page: "Access Users",
  manage_users_page: "Manage Users",
  access_shifts_page: "Access Shifts",
  manage_shifts_page: "Operate Shifts",
  access_changes_page: "Access Recent Changes",
  access_audit_page: "Access Audit Logs",
};

function formatDashboardPermissionList(value: string[]): string {
  return value
    .map((permission) => DASHBOARD_PERMISSION_LABELS[permission] || permission)
    .join(", ");
}

export function getDashboardActionVerb(action: string): string {
  const parts = action.split(".");
  return parts[parts.length - 1];
}

export function getDashboardActionBadgeColor(action: string): string {
  switch (getDashboardActionVerb(action)) {
    case "create":
    case "start":
      return "bg-green-600 text-white";
    case "update":
    case "permission_change":
    case "close":
    case "restore":
      return "bg-yellow-600 text-white";
    case "archive":
      return "bg-orange-500 text-white";
    case "delete":
      return "bg-red-600 text-white";
    case "void":
      return "bg-orange-600 text-white";
    case "adjust":
      return "bg-purple-600 text-white";
    case "count":
      return "bg-blue-600 text-white";
    default:
      return "bg-gray-600 text-white";
  }
}

export function formatDashboardDate(dateString: string): string {
  return new Date(dateString).toLocaleString("id-ID", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatDashboardDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatDashboardCurrency(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(amount);
}

export function buildDashboardStats(totalSales: string | undefined, transactionCount: number | undefined, lowStockCount: number): DashboardStats {
  const salesAmount = totalSales ? Number.parseFloat(totalSales) : 0;
  const safeTransactionCount = transactionCount || 0;

  return {
    todaySales: salesAmount,
    todayTransactions: safeTransactionCount,
    averageSale: safeTransactionCount > 0 ? salesAmount / safeTransactionCount : 0,
    lowStockCount,
  };
}

export function getDashboardChangeDescription(log: AuditLog): string {
  const newVals = log.new_values || {};
  const oldVals = log.old_values || {};
  const verb = getDashboardActionVerb(log.action);

  if (log.entity_type === "product") {
    const name = String(newVals.name || oldVals.name || "");
    if (verb === "create") return name || "New product";
    if (verb === "update") return name || "Updated product";
    if (verb === "archive") return name || "Archived product";
    if (verb === "restore") return name || "Restored product";
    if (verb === "delete") return name || "Deleted product";
  }

  if (log.entity_type === "inventory") {
    const productName =
      newVals.product_name ||
      oldVals.product_name ||
      newVals.affected_product ||
      oldVals.affected_product ||
      "";
    const adjustmentType = newVals.adjustment_type as string | undefined;
    const qty = newVals.quantity;
    const newQty = newVals.new_quantity;
    const oldThreshold = oldVals.low_stock_threshold;
    const newThreshold = newVals.low_stock_threshold;
    const adjustmentLabel = adjustmentType
      ? adjustmentType.charAt(0).toUpperCase() + adjustmentType.slice(1)
      : "Adjusted";

    if (verb === "threshold_update") {
      if (productName && oldThreshold !== undefined && newThreshold !== undefined) {
        return `Updated low stock threshold: ${productName} (${oldThreshold} → ${newThreshold})`;
      }
      if (productName) {
        return `Updated low stock threshold: ${productName}`;
      }
      return "Updated low stock threshold";
    }

    if (productName && qty !== undefined && newQty !== undefined) {
      return `${adjustmentLabel}: ${productName} → ${newQty} (Δ${qty})`;
    }
    if (productName && qty !== undefined) {
      return `${adjustmentLabel} ${productName}: ${qty}`;
    }
    if (productName) {
      return `Stock change: ${productName}`;
    }
    return "Stock adjustment";
  }

  if (log.entity_type === "sale") {
    const invoice = String(newVals.invoice_no || "");
    if (verb === "create") return invoice || "New sale created";
    if (verb === "void") return invoice || "Voided sale";
    if (verb === "delete") return invoice || "Deleted sale";
  }

  if (log.entity_type === "expense") {
    const description = String(newVals.description || newVals.affected_expense || oldVals.affected_expense || "");
    const amount = newVals.amount;
    if (verb === "create") return `${description}${amount ? ` — ${formatDashboardCurrency(Number(amount))}` : ""}`;
    if (verb === "update") return description || "Updated expense";
    if (verb === "delete") return description || "Deleted expense";
  }

  if (log.entity_type === "user") {
    const name = String(newVals.name || oldVals.name || oldVals.affected_user || newVals.affected_user || "");
    if (verb === "create") return name || "New user";
    if (verb === "update") return name || "Updated user";
    if (verb === "archive") return name || "Archived user";
    if (verb === "restore") return name || "Restored user";
    if (verb === "delete") return name || "Deleted user";
  }

  if (log.entity_type === "role") {
    const roleName = String(
      newVals.affected_role || oldVals.affected_role || newVals.name || oldVals.name || "",
    );
    if (verb === "permission_change") {
      return roleName ? `Updated role permissions: ${roleName}` : "Updated role permissions";
    }
  }

  return `${DASHBOARD_ACTION_LABELS[verb] || verb} ${log.entity_type}`;
}

function buildDashboardPermissionFieldChanges(
  oldVal: unknown,
  newVal: unknown,
): { key: string; oldVal: unknown; newVal: unknown; label: string }[] {
  const oldPermissions = Array.isArray(oldVal) ? oldVal.map(String) : [];
  const newPermissions = Array.isArray(newVal) ? newVal.map(String) : [];

  const removed = oldPermissions
    .filter((permission) => !newPermissions.includes(permission))
    .map((permission) => ({
      key: "permissions",
      label: DASHBOARD_PERMISSION_LABELS[permission] || permission,
      oldVal: "Enabled",
      newVal: "Disabled",
    }));

  const added = newPermissions
    .filter((permission) => !oldPermissions.includes(permission))
    .map((permission) => ({
      key: "permissions",
      label: DASHBOARD_PERMISSION_LABELS[permission] || permission,
      oldVal: "Disabled",
      newVal: "Enabled",
    }));

  return [...removed, ...added];
}

export function getDashboardFieldChanges(log: AuditLog): {
  key: string;
  oldVal: unknown;
  newVal: unknown;
  label?: string;
}[] {
  const oldVals = log.old_values || {};
  const newVals = log.new_values || {};
  const verb = getDashboardActionVerb(log.action);
  const allKeys = new Set([...Object.keys(oldVals), ...Object.keys(newVals)]);
  const changes: { key: string; oldVal: unknown; newVal: unknown }[] = [];

  allKeys.forEach((key) => {
    if (DASHBOARD_SKIP_FIELDS.has(key)) return;

    const oldVal = oldVals[key];
    const newVal = newVals[key];

    if (key === "permissions" && verb === "permission_change") {
      changes.push(...buildDashboardPermissionFieldChanges(oldVal, newVal));
      return;
    }

    if (
      verb === "update" ||
      verb === "permission_change" ||
      verb === "close" ||
      verb === "restore"
    ) {
      if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
        changes.push({ key, oldVal, newVal });
      }
      return;
    }

    if (verb === "create" || verb === "start" || verb === "adjust" || verb === "count" || verb === "void") {
      if (newVal !== undefined && newVal !== null && newVal !== "") {
        changes.push({ key, oldVal: undefined, newVal });
      }
      return;
    }

    if ((verb === "delete" || verb === "archive") && oldVal !== undefined && oldVal !== null && oldVal !== "") {
      changes.push({ key, oldVal, newVal: undefined });
    }
  });

  return changes;
}

export function formatDashboardFieldName(key: string, action: string): string {
  const verb = getDashboardActionVerb(action);
  if (key === "image_url") return "Photo";
  if (key === "is_active") return "Active";
  if (key === "permissions") return "Permissions";
  if (key === "adjustment_type") return "Type";
  if (key === "new_quantity") return "New Stock";
  if (key === "quantity") return "Change";
  if (key === "reason") return verb === "void" ? "Reason" : "Notes";
  return key.replace(/_/g, " ").replace(/\b\w/g, (char) => char.toUpperCase());
}

export function formatDashboardFieldValue(key: string, value: unknown): string {
  if (value === null || value === undefined) return "—";
  if (typeof value === "boolean") return value ? "Yes" : "No";
  if (key === "permissions" && Array.isArray(value)) {
    return formatDashboardPermissionList(value.map(String));
  }
  if (typeof value === "object") return JSON.stringify(value);
  if (key === "tax_rate") return `${String(value)}%`;
  if (key === "total" || key === "amount") return formatDashboardCurrency(Number(value));
  return String(value);
}

export function isDashboardImageField(key: string): boolean {
  return key === "image_url";
}

export function getShiftPreview(shift: Shift) {
  const isOpen = shift.status === "open";
  const openingCash = Number.parseFloat(shift.opening_cash || "0");
  const closingCash = shift.closing_cash ? Number.parseFloat(shift.closing_cash) : null;
  const cashDifference = shift.cash_difference ? Number.parseFloat(shift.cash_difference) : null;
  return { isOpen, openingCash, closingCash, cashDifference };
}
