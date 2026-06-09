import { describe, expect, test } from "vitest";

import type { CreateUserRequest, User } from "@/types";

import {
  canDeleteUser,
  canEditUser,
  findRoleIdByName,
  getAssignableUserRoles,
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
  permissions: [],
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

const cashierWithUserManagement: User = {
  ...cashier,
  id: "cashier-manager-1",
  permissions: ["manage_users_page"],
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

  test("cashier with manage-users permission can edit cashier users only", () => {
    expect(
      canEditUser({
        currentUser: cashierWithUserManagement,
        targetUser: cashier,
        canEditUserAny: true,
        hasPermission: makeHasPermission(cashierWithUserManagement),
      }),
    ).toBe(true);

    expect(
      canEditUser({
        currentUser: cashierWithUserManagement,
        targetUser: manager,
        canEditUserAny: true,
        hasPermission: makeHasPermission(cashierWithUserManagement),
      }),
    ).toBe(false);
  });

  test("delete protects self", () => {
    expect(
      canDeleteUser({
        currentUser: manager,
        targetUser: manager,
        canDeleteUserAny: true,
        hasPermission: makeHasPermission(manager),
      }),
    ).toBe(false);
  });

  test("cashier with manage-users permission can archive other cashier users", () => {
    expect(
      canDeleteUser({
        currentUser: cashierWithUserManagement,
        targetUser: cashier,
        canDeleteUserAny: true,
        hasPermission: makeHasPermission(cashierWithUserManagement),
      }),
    ).toBe(true);
  });
});

describe("users helper display and form helpers", () => {
  test("create-role assignment requires manage-users permission", () => {
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
          permissions: ["manage_users_page"],
        },
        isOwner: false,
        hasPermission: (permission) => permission === "manage_users_page",
      }),
    ).toEqual(["cashier"]);

    expect(
      getAssignableUserRoles({
        currentUser: cashierWithUserManagement,
        isOwner: false,
        hasPermission: makeHasPermission(cashierWithUserManagement),
      }),
    ).toEqual(["cashier"]);

    expect(
      getAssignableUserRoles({
        currentUser: owner,
        isOwner: true,
        hasPermission: makeHasPermission(owner),
      }),
    ).toEqual(["owner", "manager", "cashier"]);
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
      hasUserFormChanges(
        {
          ...formData,
          name: "Cashier Updated",
        },
        {
          ...cashier,
          email: "cashier@example.com",
        },
      ),
    ).toBe(true);

    expect(
      findRoleIdByName(
        [
          { id: "role-owner", name: "owner" },
          { id: "role-cashier", name: "cashier" },
        ],
        "cashier",
      ),
    ).toBe("role-cashier");
    expect(findRoleIdByName([], "cashier")).toBeUndefined();
  });
});
