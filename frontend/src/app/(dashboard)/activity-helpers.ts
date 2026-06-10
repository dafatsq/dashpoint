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
  label?: string;
}

type ActivityValues = Record<string, unknown>;

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
  "user.permission_change": "Update Permissions",
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
  "inventory.threshold_update": "Update Threshold",
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
  role: "Role",
};

const ACTIVITY_SKIP_FIELDS = new Set([
  "affected_user",
  "affected_role",
  "affected_product",
  "affected_category",
  "affected_expense",
  "product_name",
  "invoice_no",
  "category_id",
  "product_id",
]);

const ACTIVITY_UPDATE_VERBS = new Set([
  "update",
  "close",
  "restore",
  "permission_change",
]);
const ACTIVITY_CREATE_LIKE_VERBS = new Set([
  "create",
  "start",
  "adjust",
  "count",
  "threshold_update",
  "void",
]);
const ACTIVITY_DELETE_LIKE_VERBS = new Set(["delete", "archive"]);

function getActivityValues(log: AuditLog): {
  oldVals: ActivityValues;
  newVals: ActivityValues;
} {
  return {
    oldVals: log.old_values || {},
    newVals: log.new_values || {},
  };
}

const ACTIVITY_PERMISSION_LABELS: Record<string, string> = {
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

function formatPermissionList(value: unknown): string {
  if (!Array.isArray(value)) {
    return String(value);
  }

  return value
    .map((permission) =>
      ACTIVITY_PERMISSION_LABELS[String(permission)] || String(permission),
    )
    .join(", ");
}

function buildPermissionFieldChanges(
  oldVal: unknown,
  newVal: unknown,
): ActivityFieldChange[] {
  const oldPermissions = Array.isArray(oldVal) ? oldVal.map(String) : [];
  const newPermissions = Array.isArray(newVal) ? newVal.map(String) : [];

  const removed = oldPermissions
    .filter((permission) => !newPermissions.includes(permission))
    .map((permission) => ({
      key: "permissions",
      label: ACTIVITY_PERMISSION_LABELS[permission] || permission,
      oldVal: "Enabled",
      newVal: "Disabled",
    }));

  const added = newPermissions
    .filter((permission) => !oldPermissions.includes(permission))
    .map((permission) => ({
      key: "permissions",
      label: ACTIVITY_PERMISSION_LABELS[permission] || permission,
      oldVal: "Disabled",
      newVal: "Enabled",
    }));

  return [...removed, ...added];
}

export function getActivityActionVerb(action: string): string {
  const parts = action.split(".");
  return parts[parts.length - 1];
}

export function getActivityActionLabel(action: string): string {
  return (
    ACTIVITY_ACTION_LABELS[action] ||
    ACTIVITY_ACTION_LABELS[getActivityActionVerb(action)] ||
    action
  );
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
  if (key === "permissions") return "Permissions";
  if (key === "adjustment_type") return "Type";
  if (key === "new_quantity") return "New Stock";
  if (key === "quantity") return "Change";
  if (key === "reason") return verb === "void" ? "Reason" : "Notes";

  return key.replace(/_/g, " ").replace(/\b\w/g, (char) => char.toUpperCase());
}

export function formatActivityFieldValue(key: string, value: unknown): string {
  if (value === null || value === undefined) return "—";
  if (typeof value === "boolean") return value ? "Yes" : "No";
  if (key === "permissions") return formatPermissionList(value);
  if (typeof value === "object") return JSON.stringify(value);
  if (key === "tax_rate") return `${String(value)}%`;
  if (key === "total" || key === "amount")
    return formatActivityCurrency(Number(value));

  return String(value);
}

export function isActivityImageField(key: string): boolean {
  return key === "image_url";
}

export function buildActivityFieldChanges(
  log: AuditLog,
): ActivityFieldChange[] {
  const { oldVals, newVals } = getActivityValues(log);
  const verb = getActivityActionVerb(log.action);
  const allKeys = new Set([...Object.keys(oldVals), ...Object.keys(newVals)]);
  const changes: ActivityFieldChange[] = [];

  allKeys.forEach((key) => {
    if (ACTIVITY_SKIP_FIELDS.has(key)) return;

    const oldVal = oldVals[key];
    const newVal = newVals[key];

    if (key === "permissions" && verb === "permission_change") {
      changes.push(...buildPermissionFieldChanges(oldVal, newVal));
      return;
    }

    if (ACTIVITY_UPDATE_VERBS.has(verb)) {
      if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
        changes.push({ key, oldVal, newVal });
      }
      return;
    }

    if (ACTIVITY_CREATE_LIKE_VERBS.has(verb)) {
      if (newVal !== undefined && newVal !== null && newVal !== "") {
        changes.push({ key, oldVal: undefined, newVal });
      }
      return;
    }

    if (
      ACTIVITY_DELETE_LIKE_VERBS.has(verb) &&
      oldVal !== undefined &&
      oldVal !== null &&
      oldVal !== ""
    ) {
      changes.push({ key, oldVal, newVal: undefined });
    }
  });

  return changes;
}

export function buildActivityDescription(log: AuditLog): string {
  const { oldVals, newVals } = getActivityValues(log);
  const verb = getActivityActionVerb(log.action);

  if (log.entity_type === "product") {
    const name = String(
      newVals.name ||
        oldVals.name ||
        newVals.affected_product ||
        oldVals.affected_product ||
        "",
    );
    if (verb === "create")
      return name ? `Created product: ${name}` : "New product";
    if (verb === "update")
      return name ? `Updated product: ${name}` : "Updated product";
    if (verb === "archive")
      return name ? `Archived product: ${name}` : "Archived product";
    if (verb === "restore")
      return name ? `Restored product: ${name}` : "Restored product";
    if (verb === "delete")
      return name ? `Deleted product: ${name}` : "Deleted product";
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

    if (productName && quantity !== undefined && newQuantity !== undefined) {
      return `${adjustmentLabel} stock: ${productName} -> ${newQuantity} (Δ${quantity})`;
    }
    if (productName && quantity !== undefined) {
      return `${adjustmentLabel} stock for ${productName}: ${quantity}`;
    }
    if (productName) {
      return `Stock change: ${productName}`;
    }
    return "Stock adjustment";
  }

  if (log.entity_type === "sale") {
    const invoice = String(
      newVals.invoice_no ||
        oldVals.invoice_no ||
        newVals.affected_sale ||
        oldVals.affected_sale ||
        "",
    );
    if (verb === "create")
      return invoice ? `Created sale: ${invoice}` : "New sale created";
    if (verb === "void")
      return invoice ? `Voided sale: ${invoice}` : "Voided sale";
    if (verb === "delete")
      return invoice ? `Deleted sale: ${invoice}` : "Deleted sale";
  }

  if (log.entity_type === "expense") {
    const description = String(
      newVals.description ||
        oldVals.description ||
        newVals.affected_expense ||
        oldVals.affected_expense ||
        "",
    );
    const amount = newVals.amount || oldVals.amount;
    if (verb === "create") {
      return description
        ? `Created expense: ${description}${amount ? ` — ${formatActivityCurrency(Number(amount))}` : ""}`
        : "New expense";
    }
    if (verb === "update")
      return description
        ? `Updated expense: ${description}`
        : "Updated expense";
    if (verb === "archive")
      return description
        ? `Archived expense: ${description}`
        : "Archived expense";
    if (verb === "restore")
      return description
        ? `Restored expense: ${description}`
        : "Restored expense";
    if (verb === "delete")
      return description
        ? `Deleted expense: ${description}`
        : "Deleted expense";
  }

  if (log.entity_type === "user") {
    const name = String(
      newVals.name ||
        oldVals.name ||
        oldVals.affected_user ||
        newVals.affected_user ||
        "",
    );
    if (verb === "create") return name ? `Created user: ${name}` : "New user";
    if (verb === "update")
      return name ? `Updated user: ${name}` : "Updated user";
    if (verb === "archive")
      return name ? `Archived user: ${name}` : "Archived user";
    if (verb === "restore")
      return name ? `Restored user: ${name}` : "Restored user";
    if (verb === "delete")
      return name ? `Deleted user: ${name}` : "Deleted user";
  }

  if (log.entity_type === "role") {
    const roleName = String(
      newVals.affected_role ||
        oldVals.affected_role ||
        newVals.name ||
        oldVals.name ||
        "",
    );
    if (verb === "permission_change") {
      return roleName
        ? `Updated role permissions: ${roleName}`
        : "Updated role permissions";
    }
  }

  if (log.entity_type === "category") {
    const name = String(
      newVals.name ||
        oldVals.name ||
        oldVals.affected_category ||
        newVals.affected_category ||
        "",
    );
    if (verb === "create")
      return name ? `Created category: ${name}` : "New category";
    if (verb === "update")
      return name ? `Updated category: ${name}` : "Updated category";
    if (verb === "archive")
      return name ? `Archived category: ${name}` : "Archived category";
    if (verb === "restore")
      return name ? `Restored category: ${name}` : "Restored category";
    if (verb === "delete")
      return name ? `Deleted category: ${name}` : "Deleted category";
  }

  if (log.entity_type === "auth") {
    if (verb === "login") return "User login";
    if (verb === "pin_login") return "User PIN login";
    if (verb === "login_failed") return "Failed login attempt";
    if (verb === "logout") return "User logout";
    return "Authentication event";
  }

  if (log.entity_type === "shift") {
    if (verb === "start") return "Started shift";
    if (verb === "close") return "Closed shift";
    if (verb === "pay_in" || verb === "pay_out") {
      const type = newVals.type || "";
      const amount = newVals.amount || "";
      const reason = newVals.reason || "";
      if (type && amount) return `${type}: ${amount} - ${reason}`;
    }
    return log.description || "Shift event";
  }

  return (
    log.description ||
    `${ACTIVITY_ACTION_LABELS[verb] || verb} ${log.entity_type}`
  );
}

export function buildAuditLogParams(input: {
  page: number;
  limit: number;
  selectedAction: string;
  selectedEntity: string;
  dateRange: ActivityDateRange;
  searchQuery?: string;
}) {
  const params: {
    action?: string;
    entity_type?: string;
    from?: string;
    to?: string;
    search?: string;
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
  if (input.searchQuery && input.searchQuery.trim() !== "")
    params.search = input.searchQuery.trim();

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
