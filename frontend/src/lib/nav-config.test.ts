import { describe, expect, test } from "vitest";

import {
  getRequiredRoutePermission,
  hasRouteAccess,
  navItems,
  filterVisibleNavItems,
  routePermissions,
} from "./nav-config";

describe("nav config helpers", () => {
  test("resolves exact and parent route permissions", () => {
    expect(getRequiredRoutePermission("/")).toBeUndefined();
    expect(getRequiredRoutePermission("/products")).toBe("access_products_page");
    expect(getRequiredRoutePermission("/expenses")).toBe("access_expenses_page");
    expect(getRequiredRoutePermission("/users/123")).toBe("access_users_page");
    expect(getRequiredRoutePermission("/settings")).toBeUndefined();
  });

  test("checks multi-permission routes through hasAnyPermission", () => {
    routePermissions["/test-multi"] = ["access_sales_page", "access_products_page"];
    try {
      expect(
        hasRouteAccess("/test-multi", {
          hasPermission: () => false,
          hasAnyPermission: (permissions) =>
            permissions.includes("access_sales_page"),
        }),
      ).toBe(true);
      expect(
        hasRouteAccess("/test-multi", {
          hasPermission: () => false,
          hasAnyPermission: () => false,
        }),
      ).toBe(false);
    } finally {
      delete routePermissions["/test-multi"];
    }
  });

  test("filters visible nav items from the shared config", () => {
    const visible = filterVisibleNavItems(navItems, {
      hasPermission: (permission) =>
        permission === "access_sales_page" ||
        permission === "access_products_page" ||
        permission === "access_shifts_page",
      hasAnyPermission: (permissions) => permissions.includes("access_sales_page"),
    });

    expect(visible.some((item) => item.href === "/sales")).toBe(true);
    expect(visible.some((item) => item.href === "/products")).toBe(true);
    expect(visible.some((item) => item.href === "/users")).toBe(false);
    expect(visible.some((item) => item.href === "/shifts")).toBe(true);
  });
});
