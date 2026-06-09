export interface RolePermissionGroup {
  category: string;
  label: string;
  accessKey: string;
  manageKey?: string;
  manageLabel?: string;
  description: string;
}

export const ROLE_PERMISSION_GROUPS: RolePermissionGroup[] = [
  {
    category: "pos",
    label: "POS",
    accessKey: "access_pos_page",
    manageKey: "manage_pos_page",
    manageLabel: "Process Sales",
    description: "Access opens the POS page. Process Sales allows checkout and cart validation.",
  },
  {
    category: "products",
    label: "Products",
    accessKey: "access_products_page",
    manageKey: "manage_products_page",
    manageLabel: "Manage Products",
    description: "Access opens products. Manage Products allows creating, editing, archiving, and deleting products.",
  },
  {
    category: "inventory",
    label: "Inventory",
    accessKey: "access_inventory_page",
    manageKey: "manage_inventory_page",
    manageLabel: "Manage Inventory",
    description: "Access opens inventory. Manage Inventory allows stock adjustments and threshold edits.",
  },
  {
    category: "sales",
    label: "Sales",
    accessKey: "access_sales_page",
    manageKey: "manage_sales_page",
    manageLabel: "Void Sales",
    description: "Access opens sales history and details. Void Sales allows reversing completed sales.",
  },
  {
    category: "reports",
    label: "Reports",
    accessKey: "access_reports_page",
    manageKey: "manage_reports_page",
    manageLabel: "Export Reports",
    description: "Access opens reports. Export Reports allows downloading CSV report files.",
  },
  {
    category: "expenses",
    label: "Expenses",
    accessKey: "access_expenses_page",
    manageKey: "manage_expenses_page",
    manageLabel: "Manage Expenses",
    description: "Access opens expenses. Manage Expenses allows creating, editing, archiving, and deleting expenses.",
  },
  {
    category: "categories",
    label: "Categories",
    accessKey: "access_categories_page",
    manageKey: "manage_categories_page",
    manageLabel: "Manage Categories",
    description: "Access opens categories. Manage Categories allows changing product and expense categories.",
  },
  {
    category: "users",
    label: "Users",
    accessKey: "access_users_page",
    manageKey: "manage_users_page",
    manageLabel: "Manage Users",
    description: "Access opens users and roles. Manage Users allows staff and role permission changes.",
  },
  {
    category: "shifts",
    label: "Shifts",
    accessKey: "access_shifts_page",
    manageKey: "manage_shifts_page",
    manageLabel: "Operate Shifts",
    description: "Access opens shift history. Operate Shifts allows starting, closing, and cash drawer operations.",
  },
  {
    category: "changes",
    label: "Recent Changes",
    accessKey: "access_changes_page",
    description: "Access opens recent dashboard activity.",
  },
  {
    category: "audit",
    label: "Audit Logs",
    accessKey: "access_audit_page",
    description: "Access opens detailed audit logs.",
  },
];

const ROLE_PERMISSION_KEYS = new Set(
  ROLE_PERMISSION_GROUPS.flatMap((group) =>
    group.manageKey ? [group.accessKey, group.manageKey] : [group.accessKey],
  ),
);

const ACCESS_KEY_BY_MANAGE_KEY = ROLE_PERMISSION_GROUPS.reduce<Record<string, string>>(
  (acc, group) => {
    if (group.manageKey) {
      acc[group.manageKey] = group.accessKey;
    }
    return acc;
  },
  {},
);

const MANAGE_KEY_BY_ACCESS_KEY = ROLE_PERMISSION_GROUPS.reduce<Record<string, string>>(
  (acc, group) => {
    if (group.manageKey) {
      acc[group.accessKey] = group.manageKey;
    }
    return acc;
  },
  {},
);

export function normalizeRolePermissionKeys(keys: string[]): string[] {
  const next = new Set(keys.filter((key) => ROLE_PERMISSION_KEYS.has(key)));
  for (const [manageKey, accessKey] of Object.entries(ACCESS_KEY_BY_MANAGE_KEY)) {
    if (next.has(manageKey)) {
      next.add(accessKey);
    }
  }
  return [...next].sort();
}

export function toggleRolePermissionKey(keys: string[], key: string, enabled: boolean): string[] {
  const next = new Set(keys);

  if (enabled) {
    next.add(key);
    const accessKey = ACCESS_KEY_BY_MANAGE_KEY[key];
    if (accessKey) {
      next.add(accessKey);
    }
    return [...next].sort();
  }

  next.delete(key);
  const manageKey = MANAGE_KEY_BY_ACCESS_KEY[key];
  if (manageKey) {
    next.delete(manageKey);
  }

  return [...next].sort();
}
