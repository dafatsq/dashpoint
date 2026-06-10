import { describe, expect, test, vi } from "vitest";

import type { ApiTransport } from "./transport";
import { createUserApi } from "./users";

describe("users API", () => {
  test("strips UI-only role fields from create and update payloads", async () => {
    const request = vi.fn().mockResolvedValue({ data: { user: undefined } });
    const api = createUserApi({ request } as unknown as ApiTransport);

    await api.createUser({
      email: "cashier@example.com",
      password: "new-secret",
      name: "Cashier",
      role: "cashier",
      role_id: "role-cashier",
      pin: "1234",
      permissions: ["manage_users_page"],
    });

    await api.updateUser("user-1", {
      name: "Cashier One",
      role: "cashier",
      role_id: "role-cashier",
      permissions: ["manage_users_page"],
      expected_updated_at: "2026-06-05T10:00:00Z",
    });

    expect(request).toHaveBeenNthCalledWith(
      1,
      "/users",
      expect.objectContaining({
        method: "POST",
        body: expect.not.objectContaining({
          role: expect.any(String),
          permissions: expect.any(Array),
        }),
      }),
    );
    expect(request).toHaveBeenNthCalledWith(
      2,
      "/users/user-1",
      expect.objectContaining({
        method: "PATCH",
        body: expect.not.objectContaining({
          role: expect.any(String),
          permissions: expect.any(Array),
        }),
      }),
    );
  });
});
