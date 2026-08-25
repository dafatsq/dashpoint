// @vitest-environment jsdom

import { beforeEach, describe, expect, test } from "vitest";

import { AccountManager } from "@/lib/account-manager";
import { getRememberMeKey, migrateSession } from "@/lib/session";

import {
  captureDeviceSessionState,
  restoreDeviceSessionState,
} from "./verify-session-guard";

const USER_ID = "user-123";
const PREF_KEY = getRememberMeKey(USER_ID);

function seedSavedAccount() {
  AccountManager.saveAccount({
    id: USER_ID,
    name: "Self User",
    email: "self@dashpoint.local",
    role_name: "owner",
    has_pin: true,
  });
}

/**
 * Reproduces exactly what persistAuthUser(..., { saveAccount: false }) and
 * syncRememberMePreference do on the destructive path of a login call.
 */
function simulateLoginSideEffects() {
  AccountManager.removeAccount(USER_ID);
  window.localStorage.removeItem("dashpoint_device_trusted");
  window.localStorage.setItem(PREF_KEY, "false");
  // syncRememberMePreference sees "false" and demotes tokens to sessionStorage
  migrateSession(false);
}

describe("verify-session-guard", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
  });

  test("restores remembered-device state clobbered by the login side effect", () => {
    seedSavedAccount();
    window.localStorage.setItem(PREF_KEY, "true");
    window.localStorage.setItem("dashpoint_device_trusted", "true");
    window.localStorage.setItem("access_token", "token-a");
    window.localStorage.setItem(
      "user",
      JSON.stringify({ id: USER_ID, name: "Self User" }),
    );

    const state = captureDeviceSessionState(USER_ID);
    simulateLoginSideEffects();

    // sanity: the damage is real
    expect(AccountManager.getAccount(USER_ID)).toBeNull();
    expect(window.localStorage.getItem(PREF_KEY)).toBe("false");
    expect(window.sessionStorage.getItem("access_token")).toBe("token-a");

    restoreDeviceSessionState(state, USER_ID);

    expect(AccountManager.getAccount(USER_ID)?.id).toBe(USER_ID);
    expect(window.localStorage.getItem(PREF_KEY)).toBe("true");
    expect(window.localStorage.getItem("dashpoint_device_trusted")).toBe("true");
    // tokens migrated back to persistent storage
    expect(window.localStorage.getItem("access_token")).toBe("token-a");
    expect(window.sessionStorage.getItem("access_token")).toBeNull();
  });

  test("keeps an explicitly ephemeral session ephemeral after restore", () => {
    window.localStorage.setItem(PREF_KEY, "false");
    window.sessionStorage.setItem("access_token", "token-b");

    const state = captureDeviceSessionState(USER_ID);
    simulateLoginSideEffects();

    restoreDeviceSessionState(state, USER_ID);

    expect(window.localStorage.getItem(PREF_KEY)).toBe("false");
    expect(AccountManager.getAccount(USER_ID)).toBeNull();
    expect(window.sessionStorage.getItem("access_token")).toBe("token-b");
    expect(window.localStorage.getItem("access_token")).toBeNull();
  });

  test("removes the preference key again when it was never set", () => {
    seedSavedAccount();
    window.localStorage.setItem("access_token", "token-c");
    window.localStorage.setItem("user", JSON.stringify({ id: USER_ID }));

    const state = captureDeviceSessionState(USER_ID);
    simulateLoginSideEffects();

    restoreDeviceSessionState(state, USER_ID);

    expect(window.localStorage.getItem(PREF_KEY)).toBeNull();
    expect(AccountManager.getAccount(USER_ID)?.id).toBe(USER_ID);
  });
});
