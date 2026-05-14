import { describe, expect, test } from "vitest";

import {
  getRequiredRoutePermission,
  hasRouteAccess,
  navItems,
  filterVisibleNavItems,
} from "./nav-config";

describe("nav config helpers", () => {
  test("resolves exact and parent route permissions", () => {
    expect(getRequiredRoutePermission("/")).toBeUndefined();
    expect(getRequiredRoutePermission("/products")).toBe("can_view_products");
    expect(getRequiredRoutePermission("/expenses")).toBe("can_view_expenses");
    expect(getRequiredRoutePermission("/users/123")).toBe("can_view_users");
    expect(getRequiredRoutePermission("/settings")).toBeUndefined();
  });

  test("checks multi-permission routes through hasAnyPermission", () => {
    expect(
      hasRouteAccess("/shifts", {
        hasPermission: () => false,
        hasAnyPermission: (permissions) => permissions.includes("can_view_sales"),
      }),
    ).toBe(true);
    expect(
      hasRouteAccess("/shifts", {
        hasPermission: () => false,
        hasAnyPermission: () => false,
      }),
    ).toBe(false);
  });

  test("filters visible nav items from the shared config", () => {
    const visible = filterVisibleNavItems(navItems, {
      hasPermission: (permission) =>
        permission === "can_view_sales" || permission === "can_view_products",
      hasAnyPermission: (permissions) => permissions.includes("can_view_sales"),
    });

    expect(visible.some((item) => item.href === "/sales")).toBe(true);
    expect(visible.some((item) => item.href === "/products")).toBe(true);
    expect(visible.some((item) => item.href === "/users")).toBe(false);
    expect(visible.some((item) => item.href === "/shifts")).toBe(true);
  });
});
