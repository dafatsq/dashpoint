import type { SavedAccount } from "@/lib/account-manager";

export type LoginTab = "saved" | "email";

export const LOGIN_INFO_MESSAGES: Record<string, string> = {
  account_inactive:
    "Your account has been deactivated. Please contact an administrator.",
  account_deactivated:
    "Your account has been deactivated by an administrator.",
  account_deleted: "Your account has been deleted.",
  force_logout: "You have been logged out by an administrator.",
  permissions_changed: "Your permissions have been updated. Please log in again.",
  role_changed: "Your role has been changed. Please log in again.",
  role_changed_relogin: "Your role has been changed. Please log in again.",
};

export const DEMO_LOGIN_CREDENTIALS = [
  { role: "Owner", email: "owner@dashpoint.local", pass: "owner123" },
  { role: "Manager", email: "manager@dashpoint.local", pass: "manager123" },
  { role: "Cashier", email: "cashier@dashpoint.local", pass: "cashier123" },
] as const;

export function getLoginInfoMessage(message: string | null | undefined): string {
  if (!message) {
    return "";
  }

  return LOGIN_INFO_MESSAGES[message] ?? "";
}

export function isStoredPreferenceEnabled(value: string | null): boolean {
  return value === "true";
}

export function getShowDemoAccessPreference(
  envEnabled: boolean,
  storedValue: string | null,
): boolean {
  return envEnabled || isStoredPreferenceEnabled(storedValue);
}

export function getDefaultLoginTab(hasSavedAccounts: boolean): LoginTab {
  return hasSavedAccounts ? "saved" : "email";
}

export function getHasSavedAccounts(accounts: SavedAccount[]): boolean {
  return accounts.length > 0;
}

export function getEffectiveSaveAccountDecision(
  isDeviceTrusted: boolean,
  saveLogin: boolean,
): boolean {
  return isDeviceTrusted || saveLogin;
}

export function getSaveLoginControlState(
  isDeviceTrusted: boolean,
  saveLogin: boolean,
): { checked: boolean; disabled: boolean } {
  if (isDeviceTrusted) {
    return {
      checked: true,
      disabled: true,
    };
  }

  return {
    checked: saveLogin,
    disabled: false,
  };
}
