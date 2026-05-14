import { describe, expect, test } from "vitest";

import type { User } from "@/types";

import {
  buildProfileUpdatePayload,
  buildSettingsPreferences,
  hasSettingsPreferenceChanges,
  profileHasChanges,
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
