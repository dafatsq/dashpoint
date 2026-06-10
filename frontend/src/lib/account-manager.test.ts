import { beforeEach, describe, expect, test } from "vitest";

import { AccountManager } from "./account-manager";

const STORAGE_KEY = "dashpoint_saved_accounts";

describe("AccountManager", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
  });

  describe("getSavedAccounts", () => {
    test("drops malformed legacy entries from localStorage", () => {
      window.localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify([
          {
            id: "cashier-1",
            name: "Cashier",
            email: "cashier@dashpoint.local",
            role_name: "cashier",
            has_pin: true,
            saved_at: "2026-05-14T01:00:00.000Z",
          },
          {
            id: "legacy-1",
            name: "User",
            role_name: "cashier",
            has_pin: true,
            saved_at: "2026-05-14T00:00:00.000Z",
          },
        ]),
      );

      expect(AccountManager.getSavedAccounts()).toEqual([
        {
          id: "cashier-1",
          name: "Cashier",
          email: "cashier@dashpoint.local",
          role_name: "cashier",
          has_pin: true,
          saved_at: "2026-05-14T01:00:00.000Z",
        },
      ]);

      expect(
        JSON.parse(window.localStorage.getItem(STORAGE_KEY) || "[]"),
      ).toHaveLength(1);
    });
  });

  describe("isAccountActive", () => {
    test("returns true when the active user is stored in localStorage", () => {
      window.localStorage.setItem("user", JSON.stringify({ id: "user-1" }));

      expect(AccountManager.isAccountActive("user-1")).toBe(true);
    });

    test("returns true when the active user is stored in sessionStorage", () => {
      window.sessionStorage.setItem("user", JSON.stringify({ id: "user-2" }));

      expect(AccountManager.isAccountActive("user-2")).toBe(true);
    });
  });
});
