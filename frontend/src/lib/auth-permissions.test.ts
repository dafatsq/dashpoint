import { describe, expect, test } from "vitest";

import {
  hasAllPermissions,
  hasAnyPermission,
  hasPermission,
} from "./auth-permissions";
import type { User } from "@/types";

const cashier: User = {
  id: "cashier-1",
  name: "Cashier",
  role_id: "role-1",
  role_name: "cashier",
  is_active: true,
  has_pin: false,
  permissions: ["can_view_sales", "can_view_products"],
  created_at: "",
  updated_at: "",
};

const owner: User = {
  ...cashier,
  id: "owner-1",
  role_name: "owner",
  permissions: [],
};

describe("auth permission helpers", () => {
  test("grants all permissions to owners", () => {
    expect(hasPermission(owner, "anything")).toBe(true);
    expect(hasAnyPermission(owner, ["a", "b"])).toBe(true);
    expect(hasAllPermissions(owner, ["a", "b"])).toBe(true);
  });

  test("checks explicit permissions for non-owners", () => {
    expect(hasPermission(cashier, "can_view_sales")).toBe(true);
    expect(hasAnyPermission(cashier, ["can_edit_products", "can_view_sales"])).toBe(
      true,
    );
    expect(hasAllPermissions(cashier, ["can_view_sales", "can_view_products"])).toBe(
      true,
    );
    expect(hasAllPermissions(cashier, ["can_view_sales", "can_edit_products"])).toBe(
      false,
    );
  });
});
