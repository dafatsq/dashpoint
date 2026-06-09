import { describe, expect, test } from "vitest";

import {
  ROLE_PERMISSION_GROUPS,
  normalizeRolePermissionKeys,
  toggleRolePermissionKey,
} from "./users-role-permissions";

describe("users role permission helpers", () => {
  test("drops deprecated settings permission from submit payloads", () => {
    expect(
      normalizeRolePermissionKeys([
        "access_users_page",
        "access_settings_page",
      ]),
    ).toEqual(["access_users_page"]);
  });

  test("enabling manage permission also enables access permission", () => {
    expect(toggleRolePermissionKey([], "manage_users_page", true)).toEqual([
      "access_users_page",
      "manage_users_page",
    ]);
  });

  test("uses specific action labels for non-CRUD permissions", () => {
    const sales = ROLE_PERMISSION_GROUPS.find((group) => group.category === "sales");
    const reports = ROLE_PERMISSION_GROUPS.find((group) => group.category === "reports");

    expect(sales?.manageLabel).toBe("Void Sales");
    expect(reports?.manageLabel).toBe("Export Reports");
  });
});
