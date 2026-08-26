import type { User } from "@/types";

import { AccountManager } from "@/lib/account-manager";
import { API_BASE_URL, IS_DESKTOP_BUILD } from "@/lib/config";
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

// Web builds keep the raw JWT in module memory only: nothing writable by a
// compromised script survives the page, and the httpOnly refresh cookie is
// what actually persists a session across reloads. Desktop (Wails) builds
// keep the legacy storage behavior until their cookie handling is verified.
let memoryAccessToken: string | null = null;
let lastRefreshedUser: User | null = null;

export function getAccessToken(): string | null {
  if (IS_DESKTOP_BUILD) return getSessionItem("access_token");
  return memoryAccessToken ?? getSessionItem("access_token");
}

export function setAuthTokens(accessToken: string): void {
  if (IS_DESKTOP_BUILD) {
    setSessionItem("access_token", accessToken);
    removeSessionItem("refresh_token");
    return;
  }
  memoryAccessToken = accessToken;
  // Scrub copies written by earlier versions of the app.
  removeSessionItem("access_token");
  removeSessionItem("refresh_token");
}

export function persistUserSession(user: User): void {
  if (!IS_DESKTOP_BUILD) return;
  setSessionItem("user", JSON.stringify(user));
}

export function syncRememberMePreference(userId: string): void {
  if (!IS_DESKTOP_BUILD) return;

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

function shouldPreserveSavedAccount(userId: string): boolean {
  return AccountManager.getAccount(userId) !== null;
}

function resolveSaveAccountPreference(
  user: User,
  requestedSaveAccount?: boolean,
): boolean {
  if (requestedSaveAccount !== undefined) {
    return requestedSaveAccount;
  }

  return shouldPreserveSavedAccount(user.id);
}

export function persistAuthUser(
  apiUser: ApiUserPayload | User,
  options: { saveAccount?: boolean } = {},
): User {
  const user = normalizeUser(apiUser);
  const shouldSaveAccount = resolveSaveAccountPreference(
    user,
    options.saveAccount,
  );

  if (shouldSaveAccount) {
    syncSavedAccount(user);
    if (typeof window !== "undefined") {
      window.localStorage.setItem("dashpoint_device_trusted", "true");
      const prefKey = getRememberMeKey(user.id);
      if (window.localStorage.getItem(prefKey) !== "false") {
        window.localStorage.setItem(prefKey, "true");
      }
    }
    persistUserSession(user);
    syncRememberMePreference(user.id);
  } else {
    AccountManager.removeAccount(user.id);
    if (typeof window !== "undefined") {
      window.localStorage.removeItem("dashpoint_device_trusted");
      if (options.saveAccount === false) {
        const prefKey = getRememberMeKey(user.id);
        window.localStorage.setItem(prefKey, "false");
      }
    }
    persistUserSession(user);
    syncRememberMePreference(user.id);
  }

  return user;
}

export function persistAuthPayload(
  payload: AuthPayload,
  options: { saveAccount?: boolean } = {},
): User | null {
  if (!payload.user) {
    setAuthTokens(payload.access_token);
    return null;
  }

  const user = persistAuthUser(payload.user, options);
  setAuthTokens(payload.access_token);
  return user;
}

export function clearAuthSession(): void {
  clearSession();
  memoryAccessToken = null;
  lastRefreshedUser = null;
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

let refreshPromise: Promise<boolean> | null = null;

async function refreshSessionTokensInternal(): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
    });

    if (!response.ok) {
      clearAuthSession();
      return false;
    }

    const data = (await response.json()) as AuthPayload;
    const saveAccount =
      data.user ? shouldPreserveSavedAccount(normalizeUser(data.user).id) : false;
    lastRefreshedUser = persistAuthPayload(data, { saveAccount });
    return true;
  } catch {
    clearAuthSession();
    return false;
  }
}

export function refreshSessionTokens(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = refreshSessionTokensInternal().finally(() => {
      refreshPromise = null;
    });
  }

  return refreshPromise;
}

/**
 * Silent-refreshes the session and resolves with the refreshed user (null on
 * failure). Web builds depend on this after every page load, since the access
 * token only lives in memory.
 */
export async function refreshSessionUser(): Promise<User | null> {
  const ok = await refreshSessionTokens();
  return ok ? lastRefreshedUser : null;
}

export function clearMemoryAccessToken(): void {
  memoryAccessToken = null;
}
