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

export function setAuthTokens(accessToken: string): void {
  if (typeof window === "undefined") return;

  setSessionItem("access_token", accessToken);
  removeSessionItem("refresh_token");
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

  persistUserSession(user);
  syncRememberMePreference(user.id);

  if (shouldSaveAccount) {
    syncSavedAccount(user);
  } else {
    AccountManager.removeAccount(user.id);
  }

  return user;
}

export function persistAuthPayload(
  payload: AuthPayload,
  options: { saveAccount?: boolean } = {},
): User | null {
  setAuthTokens(payload.access_token);

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
    persistAuthPayload(data, { saveAccount });
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
