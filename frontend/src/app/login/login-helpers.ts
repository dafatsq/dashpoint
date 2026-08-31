import type { SavedAccount } from "@/lib/account-manager";

export type LoginTab = "saved" | "email";

export const LOGIN_INFO_MESSAGES: Record<string, string> = {
  account_inactive:
    "Your account has been deactivated. Please contact an administrator.",
  account_deactivated:
    "Your account has been deactivated by an administrator.",
  account_deleted: "Your account has been deleted.",
  force_logout: "You have been logged out by an administrator.",
  owner_created:
    "Owner account created successfully. Please sign in with your new credentials.",
  permissions_changed: "Your permissions have been updated. Please log in again.",
  role_changed: "Your role has been changed. Please log in again.",
  role_changed_relogin: "Your role has been changed. Please log in again.",
};

export function getLoginInfoMessage(message: string | null | undefined): string {
  if (!message) {
    return "";
  }

  return LOGIN_INFO_MESSAGES[message] ?? "";
}

export function isStoredPreferenceEnabled(value: string | null): boolean {
  return value === "true";
}

export function getDefaultLoginTab(hasSavedAccounts: boolean): LoginTab {
  return hasSavedAccounts ? "saved" : "email";
}

export function getHasSavedAccounts(accounts: SavedAccount[]): boolean {
  return accounts.length > 0;
}

export function getEffectiveSaveAccountDecision(
  _isDeviceTrusted: boolean,
  saveLogin: boolean,
): boolean {
  return saveLogin;
}

export function getSaveLoginControlState(
  _isDeviceTrusted: boolean,
  saveLogin: boolean,
): { checked: boolean; disabled: boolean } {
  return {
    checked: saveLogin,
    disabled: false,
  };
}
