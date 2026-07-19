import type { SavedAccount } from "@/lib/account-manager";

export function filterSwitchableAccounts(
  accounts: SavedAccount[],
  excludeUserId?: string,
): SavedAccount[] {
  if (!excludeUserId) return accounts;
  return accounts.filter((account) => account.id !== excludeUserId);
}

export function shouldRemoveSavedAccountAfterPINFailure(error?: string): boolean {
  return (
    error === "Invalid credentials" ||
    error === "Account is inactive" ||
    error === "Your account has been deactivated"
  );
}
