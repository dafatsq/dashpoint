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
  permissions: ["access_sales_page", "access_products_page"],
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
    expect(hasPermission(cashier, "access_sales_page")).toBe(true);
    expect(hasAnyPermission(cashier, ["manage_products_page", "access_sales_page"])).toBe(
      true,
    );
    expect(hasAllPermissions(cashier, ["access_sales_page", "access_products_page"])).toBe(
      true,
    );
    expect(hasAllPermissions(cashier, ["access_sales_page", "manage_products_page"])).toBe(
      false,
    );
  });
});
