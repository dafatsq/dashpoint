import type { User } from "@/types";

import { AccountManager } from "@/lib/account-manager";
import { API_BASE_URL } from "@/lib/config";
import {
  clearSession,
  getRememberMeKey,
  getSessionItem,
  migrateSession,
  removeSessionItem,
  setSessionItem,
} from "@/lib/session";

import {
  AuthPayload,
  hydrateStoredUser,
  normalizeUser,
  type ApiUserPayload,
} from "./auth-user";

export function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return getSessionItem("access_token");
}

export function getRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  return getSessionItem("refresh_token");
}

export function setAuthTokens(accessToken: string, refreshToken: string): void {
  if (typeof window === "undefined") return;

  setSessionItem("access_token", accessToken);
  setSessionItem("refresh_token", refreshToken);
}

export function persistUserSession(user: User): void {
  setSessionItem("user", JSON.stringify(user));
}

export function syncRememberMePreference(userId: string): void {
  if (typeof window === "undefined") return;

  const preference = window.localStorage.getItem(getRememberMeKey(userId));
  if (preference === "false") {
    migrateSession(false);
  }
}

export function syncSavedAccount(user: User): void {
  if (!user.has_pin) return;

  AccountManager.saveAccount({
    id: user.id,
    name: user.name,
    email: user.email,
    role_name: user.role_name,
    has_pin: user.has_pin,
  });
}

export function persistAuthUser(
  apiUser: ApiUserPayload | User,
  options: { saveAccount?: boolean } = {},
): User {
  const user = normalizeUser(apiUser);

  persistUserSession(user);
  syncRememberMePreference(user.id);

  if (options.saveAccount !== false) {
    syncSavedAccount(user);
  }

  return user;
}

export function persistAuthPayload(
  payload: AuthPayload,
  options: { saveAccount?: boolean } = {},
): User | null {
  setAuthTokens(payload.access_token, payload.refresh_token);

  if (!payload.user) {
    return null;
  }

  return persistAuthUser(payload.user, options);
}

export function clearAuthSession(): void {
  clearSession();
}

export function loadStoredUser(): User | null {
  const storedUser = getSessionItem("user");
  if (!storedUser) return null;

  const hydratedUser = normalizeStoredUser(storedUser);
  if (!hydratedUser) {
    removeSessionItem("user");
    return null;
  }

  persistUserSession(hydratedUser);
  return hydratedUser;
}

function normalizeStoredUser(storedUser: string): User | null {
  return hydrateStoredUser(storedUser);
}

export async function refreshSessionTokens(): Promise<boolean> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      clearAuthSession();
      return false;
    }

    const data = (await response.json()) as AuthPayload;
    persistAuthPayload(data);
    return true;
  } catch {
    clearAuthSession();
    return false;
  }
}
