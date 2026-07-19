import { describe, expect, test } from "vitest";

import {
  filterSwitchableAccounts,
  shouldRemoveSavedAccountAfterPINFailure,
} from "./account-switcher-utils";
import type { SavedAccount } from "@/lib/account-manager";

const accounts: SavedAccount[] = [
  {
    id: "owner-1",
    name: "Owner",
    email: "owner@dashpoint.local",
    role_name: "owner",
    has_pin: true,
    saved_at: "2026-05-14T00:00:00.000Z",
  },
  {
    id: "cashier-1",
    name: "Cashier",
    email: "cashier@dashpoint.local",
    role_name: "cashier",
    has_pin: true,
    saved_at: "2026-05-14T01:00:00.000Z",
  },
];

describe("account switcher utils", () => {
  test("filters out the active account when requested", () => {
    expect(filterSwitchableAccounts(accounts, "owner-1")).toEqual([accounts[1]]);
  });

  test("returns all accounts when no exclude id is provided", () => {
    expect(filterSwitchableAccounts(accounts)).toEqual(accounts);
  });

  test("removes saved account shortcuts for invalid account credentials", () => {
    expect(shouldRemoveSavedAccountAfterPINFailure("Invalid credentials")).toBe(true);
    expect(shouldRemoveSavedAccountAfterPINFailure("Account is inactive")).toBe(true);
    expect(
      shouldRemoveSavedAccountAfterPINFailure("Your account has been deactivated"),
    ).toBe(true);
    expect(shouldRemoveSavedAccountAfterPINFailure("Invalid PIN")).toBe(false);
  });
});
