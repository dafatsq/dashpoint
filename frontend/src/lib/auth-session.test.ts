import { beforeEach, describe, expect, test, vi } from "vitest";

import { AccountManager } from "@/lib/account-manager";

import { persistAuthPayload, refreshSessionTokens } from "./auth-session";

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

    persistAuthPayload(
      {
        access_token: "access",
        user: {
          id: "user-2",
          email: "manager@example.com",
          name: "Manager",
          role_name: "manager",
          is_active: true,
          has_pin: true,
        },
      },
      { saveAccount: true },
    );

    expect(saveAccountSpy).toHaveBeenCalledWith({
      id: "user-2",
      name: "Manager",
      email: "manager@example.com",
      role_name: "manager",
      has_pin: true,
    });
  });

  test("refresh does not save an account that was intentionally left unsaved", async () => {
    const saveAccountSpy = vi.spyOn(AccountManager, "saveAccount");
    const fetchSpy = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(
        new Response(
          JSON.stringify({
            access_token: "next-access",
            user: {
              id: "user-3",
              email: "owner@example.com",
              name: "Owner",
              role_name: "owner",
              is_active: true,
              has_pin: true,
            },
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        ),
      );

    persistAuthPayload(
      {
        access_token: "access",
        user: {
          id: "user-3",
          email: "owner@example.com",
          name: "Owner",
          role_name: "owner",
          is_active: true,
          has_pin: true,
        },
      },
      { saveAccount: false },
    );

    saveAccountSpy.mockClear();

    await expect(refreshSessionTokens()).resolves.toBe(true);
    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining("/auth/refresh"),
      expect.objectContaining({
        method: "POST",
        credentials: "include",
      }),
    );
    expect(saveAccountSpy).not.toHaveBeenCalled();
  });

  test("does not store refresh tokens in browser storage", () => {
    window.localStorage.setItem("refresh_token", "legacy-refresh");
    window.sessionStorage.setItem("refresh_token", "legacy-refresh");

    persistAuthPayload({
      access_token: "access",
      user: {
        id: "user-4",
        email: "cashier@example.com",
        name: "Cashier",
        role_name: "cashier",
        is_active: true,
        has_pin: false,
      },
    });

    expect(window.localStorage.getItem("access_token")).toBe("access");
    expect(window.localStorage.getItem("refresh_token")).toBeNull();
    expect(window.sessionStorage.getItem("refresh_token")).toBeNull();
  });
});
