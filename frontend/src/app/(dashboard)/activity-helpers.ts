import type { AuditLog } from "@/types";

export type ActivityChangeTab = "product" | "inventory" | "sale" | "expense";

export interface ActivityDateRange {
  start: string;
  end: string;
}

export interface ActivityFieldChange {
  key: string;
  oldVal: unknown;
  newVal: unknown;
}

export const ACTIVITY_ACTION_LABELS: Record<string, string> = {
  login: "Login",
  login_failed: "Login Failed",
  logout: "Logout",
  create: "Created",
  update: "Updated",
  delete: "Deleted",
  void: "Voided",
  adjust: "Adjusted",
  count: "Counted",
  start: "Started",
  close: "Closed",
  archive: "Archived",
  restore: "Restored",
  "auth.login": "Login",
  "auth.login_failed": "Login Failed",
  "auth.logout": "Logout",
  "auth.pin_login": "PIN Login",
  "user.create": "Create User",
  "user.update": "Update User",
  "user.delete": "Delete User",
  "user.archive": "Archive User",
  "user.restore": "Restore User",
  "product.create": "Create Product",
  "product.update": "Update Product",
  "product.delete": "Delete Product",
  "product.archive": "Archive Product",
  "product.restore": "Restore Product",
  "inventory.adjust": "Adjust Stock",
  "inventory.count": "Stock Count",
  "category.create": "Create Category",
  "category.update": "Update Category",
  "category.delete": "Delete Category",
  "category.archive": "Archive Category",
  "category.restore": "Restore Category",
  "expense.create": "Create Expense",
  "expense.update": "Update Expense",
  "expense.delete": "Delete Expense",
  "expense.archive": "Archive Expense",
  "expense.restore": "Restore Expense",
  "sale.create": "Sale",
  "sale.void": "Void Sale",
  "sale.refund": "Refund",
  "shift.start": "Start Shift",
  "shift.close": "Close Shift",
};

export const ACTIVITY_ENTITY_LABELS: Record<string, string> = {
  auth: "Authentication",
  user: "User",
  product: "Product",
  category: "Category",
  sale: "Sale",
  shift: "Shift",
  inventory: "Inventory",
  expense: "Expense",
};

const ACTIVITY_SKIP_FIELDS = new Set([
  "affected_user",
  "affected_product",
  "affected_category",
  "affected_expense",
  "product_name",
  "invoice_no",
  "category_id",
  "product_id",
]);

export function getActivityActionVerb(action: string): string {
  const parts = action.split(".");
  return parts[parts.length - 1];
}

export function getActivityActionLabel(action: string): string {
  return ACTIVITY_ACTION_LABELS[action] || ACTIVITY_ACTION_LABELS[getActivityActionVerb(action)] || action;
}

export function getActivityEntityLabel(entityType: string): string {
  return ACTIVITY_ENTITY_LABELS[entityType] || entityType;
}

export function getActivityBadgeColor(action: string): string {
  const verb = getActivityActionVerb(action);

  switch (verb) {
    case "create":
    case "start":
      return "bg-green-600 text-white";
    case "update":
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
      if (action.includes("login")) {
        return action.includes("failed")
          ? "bg-red-600 text-white"
          : "bg-blue-600 text-white";
      }
      return "bg-gray-600 text-white";
  }
}

export function formatActivityDate(dateString: string): string {
  return new Date(dateString).toLocaleString("id-ID", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatActivityDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString("id-ID", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export function formatActivityCurrency(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(amount);
}

export function formatActivityFieldName(key: string, action: string): string {
  const verb = getActivityActionVerb(action);

  if (key === "image_url") return "Photo";
  if (key === "is_active") return "Active";
  if (key === "adjustment_type") return "Type";
  if (key === "new_quantity") return "New Stock";
  if (key === "quantity") return "Change";
  if (key === "reason") return verb === "void" ? "Reason" : "Notes";

  return key.replace(/_/g, " ").replace(/\b\w/g, (char) => char.toUpperCase());
}

export function formatActivityFieldValue(key: string, value: unknown): string {
  if (value === null || value === undefined) return "—";
  if (typeof value === "boolean") return value ? "Yes" : "No";
  if (typeof value === "object") return JSON.stringify(value);
  if (key === "tax_rate") return `${String(value)}%`;
  if (key === "total" || key === "amount") return formatActivityCurrency(Number(value));

  return String(value);
}

export function isActivityImageField(key: string): boolean {
  return key === "image_url";
}

export function buildActivityFieldChanges(log: AuditLog): ActivityFieldChange[] {
  const oldVals = log.old_values || {};
  const newVals = log.new_values || {};
  const verb = getActivityActionVerb(log.action);
  const allKeys = new Set([...Object.keys(oldVals), ...Object.keys(newVals)]);
  const changes: ActivityFieldChange[] = [];

  allKeys.forEach((key) => {
    if (ACTIVITY_SKIP_FIELDS.has(key)) return;

    const oldVal = oldVals[key];
    const newVal = newVals[key];

    if (verb === "update" || verb === "close" || verb === "restore") {
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

export function buildActivityDescription(log: AuditLog): string {
  const newVals = log.new_values || {};
  const oldVals = log.old_values || {};
  const verb = getActivityActionVerb(log.action);

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
    const quantity = newVals.quantity;
    const newQuantity = newVals.new_quantity;
    const adjustmentLabel = adjustmentType
      ? adjustmentType.charAt(0).toUpperCase() + adjustmentType.slice(1)
      : "Adjusted";

    if (productName && quantity !== undefined && newQuantity !== undefined) {
      return `${adjustmentLabel}: ${productName} -> ${newQuantity} (Δ${quantity})`;
    }
    if (productName && quantity !== undefined) {
      return `${adjustmentLabel} ${productName}: ${quantity}`;
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
    if (verb === "create") {
      return `${description}${amount ? ` — ${formatActivityCurrency(Number(amount))}` : ""}`;
    }
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

  return `${ACTIVITY_ACTION_LABELS[verb] || verb} ${log.entity_type}`;
}

export function buildAuditLogParams(input: {
  page: number;
  limit: number;
  selectedAction: string;
  selectedEntity: string;
  dateRange: ActivityDateRange;
}) {
  const params: {
    action?: string;
    entity_type?: string;
    from?: string;
    to?: string;
    limit: number;
    offset: number;
  } = {
    limit: input.limit,
    offset: (input.page - 1) * input.limit,
  };

  if (input.selectedAction !== "all") params.action = input.selectedAction;
  if (input.selectedEntity !== "all") params.entity_type = input.selectedEntity;
  if (input.dateRange.start) params.from = input.dateRange.start;
  if (input.dateRange.end) params.to = input.dateRange.end;

  return params;
}

export function buildDashboardChangeParams(input: {
  entityType: ActivityChangeTab;
  page: number;
  limit: number;
  dateRange: ActivityDateRange;
  selectedUser: string;
}) {
  const params: {
    entity_type: ActivityChangeTab;
    limit: number;
    offset: number;
    user_id?: string;
    from?: string;
    to?: string;
  } = {
    entity_type: input.entityType,
    limit: input.limit,
    offset: (input.page - 1) * input.limit,
  };

  if (input.dateRange.start) params.from = input.dateRange.start;
  if (input.dateRange.end) params.to = input.dateRange.end;
  if (input.selectedUser !== "all") params.user_id = input.selectedUser;

  return params;
}

export function buildChangesListResetKey(input: {
  entityType: ActivityChangeTab;
  dateRange: ActivityDateRange;
  selectedUser: string;
}): string {
  return `${input.entityType}|${input.selectedUser}|${input.dateRange.start}|${input.dateRange.end}`;
}

export function incrementActivityRefreshKey(currentKey: number): number {
  return currentKey + 1;
}
