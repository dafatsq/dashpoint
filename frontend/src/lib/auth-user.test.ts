import { describe, expect, test } from "vitest";

import { hydrateStoredUser, normalizeUser } from "./auth-user";

describe("normalizeUser", () => {
  test("fills frontend defaults from a backend auth payload", () => {
    const user = normalizeUser({
      id: "user-1",
      name: "Cashier",
      role_name: "cashier",
      is_active: true,
    });

    expect(user).toEqual({
      id: "user-1",
      name: "Cashier",
      email: undefined,
      role_id: "",
      role_name: "cashier",
      is_active: true,
      has_pin: false,
      permissions: [],
      created_at: "",
      updated_at: "",
    });
  });
});

describe("hydrateStoredUser", () => {
  test("migrates legacy role to role_name", () => {
    const user = hydrateStoredUser(
      JSON.stringify({
        id: "user-2",
        name: "Manager",
        role: "manager",
        is_active: true,
        has_pin: true,
        created_at: "2024-01-01",
        updated_at: "2024-01-01",
      }),
    );

    expect(user?.role_name).toBe("manager");
  });
});
