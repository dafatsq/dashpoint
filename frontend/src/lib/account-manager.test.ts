import { beforeEach, describe, expect, test } from "vitest";

import { AccountManager } from "./account-manager";

describe("AccountManager.isAccountActive", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
  });

  test("returns true when the active user is stored in localStorage", () => {
    window.localStorage.setItem("user", JSON.stringify({ id: "user-1" }));

    expect(AccountManager.isAccountActive("user-1")).toBe(true);
  });

  test("returns true when the active user is stored in sessionStorage", () => {
    window.sessionStorage.setItem("user", JSON.stringify({ id: "user-2" }));

    expect(AccountManager.isAccountActive("user-2")).toBe(true);
  });
});
