import type { UpdateUserRequest, User } from "@/types";

export interface SettingsPreferences {
  rememberMe: boolean;
  quickAccess: boolean;
}

export interface SettingsProfileForm {
  name: string;
  email: string;
  password: string;
  pin: string;
}

// Quick Access (saved account + PIN fast login) and Automatic Sign-In
// (persistent refresh cookie) are independent toggles.
export function normalizeSettingsPreferences(
  preferences: SettingsPreferences,
): SettingsPreferences {
  return preferences;
}

export function updateRememberMePreference(
  current: SettingsPreferences,
  rememberMe: boolean,
): SettingsPreferences {
  return { ...current, rememberMe };
}

export function updateQuickAccessPreference(
  current: SettingsPreferences,
  quickAccess: boolean,
): SettingsPreferences {
  return { ...current, quickAccess };
}

export function buildSettingsPreferences(rememberScopeEnabled: boolean, hasSavedAccount: boolean): SettingsPreferences {
  return normalizeSettingsPreferences({
    rememberMe: rememberScopeEnabled,
    quickAccess: hasSavedAccount,
  });
}

export function hasSettingsPreferenceChanges(
  current: SettingsPreferences,
  initial: SettingsPreferences,
): boolean {
  return current.rememberMe !== initial.rememberMe || current.quickAccess !== initial.quickAccess;
}

export function profileHasChanges(user: User, form: SettingsProfileForm): boolean {
  return (
    form.name !== user.name ||
    form.email !== (user.email || "") ||
    form.password !== "" ||
    form.pin !== ""
  );
}

export function buildProfileUpdatePayload(user: User, form: SettingsProfileForm): UpdateUserRequest {
  const payload: UpdateUserRequest = {};

  if (form.name !== user.name) payload.name = form.name;
  if (form.email !== (user.email || "")) payload.email = form.email;
  if (form.password) payload.password = form.password;
  if (form.pin) payload.pin = form.pin;

  return payload;
}

/**
 * Automatic sign-in requires the account to be saved on this device: if it
 * is not, the refresh cookie is session-scoped and auto login is disabled.
 */
export function effectiveRememberMe(
  rememberMe: boolean,
  hasSavedAccount: boolean,
): boolean {
  return rememberMe && hasSavedAccount;
}
