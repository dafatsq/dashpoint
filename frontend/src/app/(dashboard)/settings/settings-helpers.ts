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

export function buildSettingsPreferences(savedPreference: string | null, hasSavedAccount: boolean): SettingsPreferences {
  const rememberMe = savedPreference !== "false";

  return {
    rememberMe,
    quickAccess: hasSavedAccount,
  };
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
