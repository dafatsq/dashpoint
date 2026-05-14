import { describe, expect, test } from "vitest";

import {
  getDefaultLoginTab,
  getEffectiveSaveAccountDecision,
  getLoginInfoMessage,
  getSaveLoginControlState,
  getShowDemoAccessPreference,
  isStoredPreferenceEnabled,
} from "./login-helpers";

describe("login helpers", () => {
  test("maps supported logout messages", () => {
    expect(getLoginInfoMessage("force_logout")).toBe(
      "You have been logged out by an administrator.",
    );
    expect(getLoginInfoMessage("role_changed_relogin")).toBe(
      "Your role has been changed. Please log in again.",
    );
    expect(getLoginInfoMessage("unknown")).toBe("");
    expect(getLoginInfoMessage(null)).toBe("");
  });

  test("parses trusted-device and demo-access preferences", () => {
    expect(isStoredPreferenceEnabled("true")).toBe(true);
    expect(isStoredPreferenceEnabled("false")).toBe(false);
    expect(isStoredPreferenceEnabled(null)).toBe(false);
    expect(getShowDemoAccessPreference(false, "true")).toBe(true);
    expect(getShowDemoAccessPreference(true, null)).toBe(true);
    expect(getShowDemoAccessPreference(false, null)).toBe(false);
  });

  test("derives the default tab from saved-account presence", () => {
    expect(getDefaultLoginTab(true)).toBe("saved");
    expect(getDefaultLoginTab(false)).toBe("email");
  });

  test("computes the effective save-account decision from trust and checkbox state", () => {
    expect(getEffectiveSaveAccountDecision(false, false)).toBe(false);
    expect(getEffectiveSaveAccountDecision(false, true)).toBe(true);
    expect(getEffectiveSaveAccountDecision(true, false)).toBe(true);
    expect(getEffectiveSaveAccountDecision(true, true)).toBe(true);
  });

  test("derives the save-login control state from trusted-device and checkbox values", () => {
    expect(getSaveLoginControlState(false, false)).toEqual({
      checked: false,
      disabled: false,
    });
    expect(getSaveLoginControlState(false, true)).toEqual({
      checked: true,
      disabled: false,
    });
    expect(getSaveLoginControlState(true, false)).toEqual({
      checked: true,
      disabled: true,
    });
    expect(getSaveLoginControlState(true, true)).toEqual({
      checked: true,
      disabled: true,
    });
  });
});
