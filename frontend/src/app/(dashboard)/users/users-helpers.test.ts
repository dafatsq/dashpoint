import { describe, expect, test } from "vitest";

import type {
  CreateUserRequest,
  Permission,
  PermissionOverride,
  User,
} from "@/types";

import {
  canDeleteUser,
  canEditUser,
  canManageUserPermissions,
  createPermissionChangeSet,
  findRoleIdByName,
  getAssignableUserRoles,
  getPermissionDisplayName,
  getVisibleChangesCount,
  hasUserFormChanges,
} from "./users-helpers";

const owner: User = {
  id: "owner-1",
  name: "Owner",
  role_id: "role-owner",
  role_name: "owner",
  is_active: true,
  has_pin: true,
  permissions: [],
  created_at: "",
  updated_at: "",
};

const manager: User = {
  id: "manager-1",
  name: "Manager",
  role_id: "role-manager",
  role_name: "manager",
  is_active: true,
  has_pin: true,
  permissions: [
    "can_edit_cashier_users",
    "can_delete_cashier_users",
    "can_manage_cashier_permissions",
  ],
  created_at: "",
  updated_at: "",
};

const cashier: User = {
  id: "cashier-1",
  name: "Cashier",
  role_id: "role-cashier",
  role_name: "cashier",
  is_active: true,
  has_pin: true,
  permissions: [],
  created_at: "",
  updated_at: "",
};

const createUserPermission: Permission = {
  id: "perm-create-user",
  key: "can_create_user",
  name: "Create User",
  category: "users",
};

const createCashierUsersPermission: Permission = {
  id: "perm-create-cashier",
  key: "can_create_cashier_users",
  name: "Create Cashier Users",
  category: "users",
};

const managePermissionsPermission: Permission = {
  id: "perm-manage-permissions",
  key: "can_manage_permissions",
  name: "Manage Permissions",
  category: "users",
};

const manageCashierPermissionsPermission: Permission = {
  id: "perm-manage-cashier-permissions",
  key: "can_manage_cashier_permissions",
  name: "Manage Cashier Permissions",
  category: "users",
};

const viewSalesPermission: Permission = {
  id: "perm-view-sales",
  key: "can_view_sales",
  name: "View Sales",
  category: "sales",
};

const viewUsersPermission: Permission = {
  id: "perm-view-users",
  key: "can_view_users",
  name: "View Users",
  category: "users",
};

const voidSalePermission: Permission = {
  id: "perm-void-sale",
  key: "can_void_sale",
  name: "Void Sale",
  category: "sales",
};

const makeHasPermission =
  (user: User) =>
  (permission: string): boolean =>
    user.role_name === "owner" || (user.permissions || []).includes(permission);

describe("users helper policy rules", () => {
  test("manager can edit cashier but not manager", () => {
    expect(
      canEditUser({
        currentUser: manager,
        targetUser: cashier,
        canEditUserAny: true,
        hasPermission: makeHasPermission(manager),
      }),
    ).toBe(true);

    expect(
      canEditUser({
        currentUser: manager,
        targetUser: { ...manager, id: "manager-2" },
        canEditUserAny: true,
        hasPermission: makeHasPermission(manager),
      }),
    ).toBe(false);
  });

  test("delete and permission management protect self", () => {
    expect(
      canDeleteUser({
        currentUser: manager,
        targetUser: manager,
        canDeleteUserAny: true,
        hasPermission: makeHasPermission(manager),
      }),
    ).toBe(false);

    expect(
      canManageUserPermissions({
        currentUser: owner,
        targetUser: owner,
        canManagePermissions: true,
        hasPermission: makeHasPermission(owner),
      }),
    ).toBe(false);
  });
});

describe("users helper permission cascades", () => {
  const groupedPermissions: Record<string, Permission[]> = {
    sales: [viewSalesPermission, voidSalePermission],
    users: [
      viewUsersPermission,
      createUserPermission,
      createCashierUsersPermission,
      managePermissionsPermission,
      manageCashierPermissionsPermission,
    ],
  };

  const managerOverrides: PermissionOverride[] = [];

  test("disabling sales access also disables void sale", () => {
    const changes = createPermissionChangeSet({
      permission: viewSalesPermission,
      enabled: false,
      allPermissions: groupedPermissions,
      permissionChanges: {},
      userOverrides: managerOverrides,
      userEffectivePermissions: ["can_view_sales", "can_void_sale"],
      currentUser: owner,
      targetUser: manager,
    });

    expect(changes[viewSalesPermission.id]).toBe(false);
    expect(changes[voidSalePermission.id]).toBe(false);
  });

  test("disabling parent user permission cascades hidden cashier child", () => {
    const changes = createPermissionChangeSet({
      permission: managePermissionsPermission,
      enabled: false,
      allPermissions: groupedPermissions,
      permissionChanges: {},
      userOverrides: managerOverrides,
      userEffectivePermissions: [
        "can_manage_permissions",
        "can_manage_cashier_permissions",
      ],
      currentUser: owner,
      targetUser: cashier,
    });

    expect(changes[managePermissionsPermission.id]).toBe(false);
    expect(changes[manageCashierPermissionsPermission.id]).toBe(false);
  });

  test("re-enabling users access does not silently restore hidden user-management grants", () => {
    const changes = createPermissionChangeSet({
      permission: viewUsersPermission,
      enabled: true,
      allPermissions: groupedPermissions,
      permissionChanges: {
        [viewUsersPermission.id]: false,
        [createCashierUsersPermission.id]: false,
        [manageCashierPermissionsPermission.id]: false,
      },
      userOverrides: managerOverrides,
      userEffectivePermissions: [
        "can_view_users",
        "can_create_cashier_users",
        "can_manage_cashier_permissions",
      ],
      currentUser: owner,
      targetUser: manager,
    });

    expect(changes[viewUsersPermission.id]).toBeUndefined();
    expect(changes[createCashierUsersPermission.id]).toBe(false);
    expect(changes[manageCashierPermissionsPermission.id]).toBe(false);
  });
});

describe("users helper display and form helpers", () => {
  test("create-role assignment requires at least one explicit child role grant for managers", () => {
    expect(
      getAssignableUserRoles({
        currentUser: manager,
        isOwner: false,
        hasPermission: makeHasPermission(manager),
      }),
    ).toEqual([]);

    expect(
      getAssignableUserRoles({
        currentUser: {
          ...manager,
          permissions: ["can_create_cashier_users"],
        },
        isOwner: false,
        hasPermission: (permission) => permission === "can_create_cashier_users",
      }),
    ).toEqual(["cashier"]);
  });

  test("normalizes delete/archive display names", () => {
    expect(getPermissionDisplayName(voidSalePermission, "sales")).toBe(
      "Void Sale",
    );
    expect(
      getPermissionDisplayName(
        {
          id: "perm-delete-expense",
          key: "can_delete_expense",
          name: "Delete Expense",
          category: "expenses",
        },
        "expenses",
      ),
    ).toBe("Delete Expense");
    expect(
      getPermissionDisplayName(
        {
          id: "perm-delete-user",
          key: "can_delete_user",
          name: "Delete/User",
          category: "users",
        },
        "users",
      ),
    ).toBe("Delete/archive Users");
  });

  test("counts only visible permission changes", () => {
    expect(
      getVisibleChangesCount(
        {
          "perm-create-user": false,
          "perm-create-cashier": false,
        },
        {
          users: [createUserPermission, createCashierUsersPermission],
        },
        null
      ),
    ).toBe(1);
  });

  test("detects create/edit form changes and resolves role IDs", () => {
    const formData: CreateUserRequest = {
      email: "cashier@example.com",
      password: "",
      name: "Cashier",
      role: "cashier",
      pin: "",
    };

    expect(hasUserFormChanges(formData, null)).toBe(true);
    expect(
      hasUserFormChanges(
        {
          ...formData,
          password: "",
          pin: "",
        },
        {
          ...cashier,
          email: "cashier@example.com",
        },
      ),
    ).toBe(false);

    expect(
      findRoleIdByName(
        [
          { id: "owner-id", name: "owner" },
          { id: "cashier-id", name: "cashier" },
        ],
        "cashier",
      ),
    ).toBe("cashier-id");
  });
});
