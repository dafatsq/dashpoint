import { describe, expect, test } from "vitest";

import type { User } from "@/types";

import {
  buildProfileUpdatePayload,
  buildSettingsPreferences,
  hasSettingsPreferenceChanges,
  normalizeSettingsPreferences,
  profileHasChanges,
  updateQuickAccessPreference,
  updateRememberMePreference,
} from "./settings-helpers";

const user: User = {
  id: "u1",
  name: "Alice",
  email: "alice@example.com",
  role_id: "r1",
  role_name: "manager",
  is_active: true,
  has_pin: true,
  created_at: "2026-05-14T00:00:00Z",
  updated_at: "2026-05-14T00:00:00Z",
};

describe("settings helpers", () => {
  test("builds initial settings preferences", () => {
    expect(buildSettingsPreferences(null, true)).toEqual({
      rememberMe: true,
      quickAccess: true,
    });

    expect(buildSettingsPreferences("false", false)).toEqual({
      rememberMe: false,
      quickAccess: false,
    });

    expect(buildSettingsPreferences("true", false)).toEqual({
      rememberMe: false,
      quickAccess: false,
    });
  });

  test("normalizes dependent remember me and quick access preferences", () => {
    expect(
      normalizeSettingsPreferences({
        rememberMe: true,
        quickAccess: false,
      }),
    ).toEqual({
      rememberMe: false,
      quickAccess: false,
    });

    expect(
      normalizeSettingsPreferences({
        rememberMe: false,
        quickAccess: true,
      }),
    ).toEqual({
      rememberMe: false,
      quickAccess: true,
    });

    expect(
      normalizeSettingsPreferences({
        rememberMe: true,
        quickAccess: true,
      }),
    ).toEqual({
      rememberMe: true,
      quickAccess: true,
    });
  });

  test("enabling remember me automatically enables quick access", () => {
    expect(
      updateRememberMePreference(
        { rememberMe: false, quickAccess: false },
        true,
      ),
    ).toEqual({
      rememberMe: true,
      quickAccess: true,
    });

    expect(
      updateRememberMePreference(
        { rememberMe: true, quickAccess: true },
        false,
      ),
    ).toEqual({
      rememberMe: false,
      quickAccess: true,
    });
  });

  test("disabling quick access automatically disables remember me", () => {
    expect(
      updateQuickAccessPreference(
        { rememberMe: true, quickAccess: true },
        false,
      ),
    ).toEqual({
      rememberMe: false,
      quickAccess: false,
    });

    expect(
      updateQuickAccessPreference(
        { rememberMe: false, quickAccess: false },
        true,
      ),
    ).toEqual({
      rememberMe: false,
      quickAccess: true,
    });
  });

  test("detects preference changes", () => {
    expect(
      hasSettingsPreferenceChanges(
        { rememberMe: true, quickAccess: true },
        { rememberMe: true, quickAccess: true },
      ),
    ).toBe(false);

    expect(
      hasSettingsPreferenceChanges(
        { rememberMe: true, quickAccess: true },
        { rememberMe: false, quickAccess: true },
      ),
    ).toBe(true);
  });

  test("detects profile form changes", () => {
    expect(
      profileHasChanges(user, {
        name: "Alice",
        email: "alice@example.com",
        password: "",
        pin: "",
      }),
    ).toBe(false);

    expect(
      profileHasChanges(user, {
        name: "Alice Jones",
        email: "alice@example.com",
        password: "",
        pin: "",
      }),
    ).toBe(true);
  });

  test("builds typed profile update payload with only changed values", () => {
    expect(
      buildProfileUpdatePayload(user, {
        name: "Alice Jones",
        email: "alice@example.com",
        password: "",
        pin: "",
      }),
    ).toEqual({
      name: "Alice Jones",
    });

    expect(
      buildProfileUpdatePayload(user, {
        name: "Alice",
        email: "alice@newmail.com",
        password: "new-password",
        pin: "123456",
      }),
    ).toEqual({
      email: "alice@newmail.com",
      password: "new-password",
      pin: "123456",
    });
  });
});
