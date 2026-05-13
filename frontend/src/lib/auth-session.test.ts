import { beforeEach, describe, expect, test, vi } from "vitest";

import { AccountManager } from "@/lib/account-manager";

import { persistAuthPayload } from "./auth-session";

describe("persistAuthPayload", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
    vi.restoreAllMocks();
  });

  test("does not save the account when saveAccount is false", () => {
    const saveAccountSpy = vi.spyOn(AccountManager, "saveAccount");

    persistAuthPayload(
      {
        access_token: "access",
        refresh_token: "refresh",
        user: {
          id: "user-1",
          email: "cashier@example.com",
          name: "Cashier",
          role_name: "cashier",
          is_active: true,
          has_pin: true,
        },
      },
      { saveAccount: false },
    );

    expect(saveAccountSpy).not.toHaveBeenCalled();
  });

  test("saves the account when saveAccount is true", () => {
    const saveAccountSpy = vi.spyOn(AccountManager, "saveAccount");

    persistAuthPayload({
      access_token: "access",
      refresh_token: "refresh",
      user: {
        id: "user-2",
        email: "manager@example.com",
        name: "Manager",
        role_name: "manager",
        is_active: true,
        has_pin: true,
      },
    });

    expect(saveAccountSpy).toHaveBeenCalledWith({
      id: "user-2",
      name: "Manager",
      email: "manager@example.com",
      role_name: "manager",
      has_pin: true,
    });
  });
});
