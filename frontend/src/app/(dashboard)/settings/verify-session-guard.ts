import { AccountManager } from "@/lib/account-manager";
import { getRememberMeKey, migrateSession } from "@/lib/session";

export interface DeviceSessionState {
  savedPreference: string | null;
  savedAccount: ReturnType<typeof AccountManager.getAccount>;
  deviceWasTrusted: boolean;
}

/**
 * Captures the remembered-device state that a `login(..., saveAccount=false)`
 * call destroys as a side effect (saved quick-access account, remember-me
 * preference, trusted-device flag).
 */
export function captureDeviceSessionState(userId: string): DeviceSessionState {
  return {
    savedPreference: window.localStorage.getItem(getRememberMeKey(userId)),
    savedAccount: AccountManager.getAccount(userId),
    deviceWasTrusted:
      window.localStorage.getItem("dashpoint_device_trusted") === "true",
  };
}

/**
 * Restores the captured state and re-aligns token storage with the original
 * remember-me preference (the login call migrates live tokens into
 * sessionStorage when it forces the preference off).
 */
export function restoreDeviceSessionState(state: DeviceSessionState, userId: string): void {
  const preferenceKey = getRememberMeKey(userId);
  if (state.savedPreference === null) {
    window.localStorage.removeItem(preferenceKey);
  } else {
    window.localStorage.setItem(preferenceKey, state.savedPreference);
  }

  if (state.savedAccount) {
    AccountManager.saveAccount(state.savedAccount);
  } else {
    AccountManager.removeAccount(userId);
  }

  if (state.deviceWasTrusted) {
    window.localStorage.setItem("dashpoint_device_trusted", "true");
  } else {
    window.localStorage.removeItem("dashpoint_device_trusted");
  }

  migrateSession(state.savedPreference !== "false");
}
