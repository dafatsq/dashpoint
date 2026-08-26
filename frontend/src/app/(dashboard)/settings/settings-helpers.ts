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

// Dependency: Automatic Sign-In requires Quick Access. Toggling auto on
// force-enables Quick Access; disabling Quick Access force-disables auto.
export function normalizeSettingsPreferences(
  preferences: SettingsPreferences,
): SettingsPreferences {
  if (!preferences.quickAccess) {
    return { rememberMe: false, quickAccess: false };
  }
  return preferences;
}

export function updateRememberMePreference(
  current: SettingsPreferences,
  rememberMe: boolean,
): SettingsPreferences {
  if (rememberMe) {
    return { rememberMe: true, quickAccess: true };
  }
  return { ...current, rememberMe: false };
}

export function updateQuickAccessPreference(
  current: SettingsPreferences,
  quickAccess: boolean,
): SettingsPreferences {
  if (!quickAccess) {
    return { rememberMe: false, quickAccess: false };
  }
  return { ...current, quickAccess: true };
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
