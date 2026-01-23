/**
 * Account Manager - Handles multiple employee accounts on a single device
 * Stores account information (without sensitive tokens) for quick switching
 */

export interface SavedAccount {
  id: string;
  name: string;
  email?: string;
  role_name: 'owner' | 'manager' | 'cashier';
  has_pin: boolean;
  saved_at: string;
}

const STORAGE_KEY = 'dashpoint_saved_accounts';
const MAX_ACCOUNTS = 10;

export class AccountManager {
  /**
   * Get all saved accounts from localStorage
   */
  static getSavedAccounts(): SavedAccount[] {
    if (typeof window === 'undefined') return [];
    
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (!stored) return [];
      
      const accounts = JSON.parse(stored) as SavedAccount[];
      return accounts.sort((a, b) => new Date(b.saved_at).getTime() - new Date(a.saved_at).getTime());
    } catch {
      return [];
    }
  }

  /**
   * Save or update an account (called after successful login)
   */
  static saveAccount(account: Omit<SavedAccount, 'saved_at'>): void {
    if (typeof window === 'undefined') return;
    
    const accounts = this.getSavedAccounts();
    
    // Remove existing account with same ID
    const filtered = accounts.filter(a => a.id !== account.id);
    
    // Add new/updated account at the beginning
    const newAccount: SavedAccount = {
      ...account,
      saved_at: new Date().toISOString(),
    };
    
    const updated = [newAccount, ...filtered].slice(0, MAX_ACCOUNTS);
    
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
    } catch (error) {
      console.error('Failed to save account:', error);
    }
  }

  /**
   * Remove an account from saved accounts
   */
  static removeAccount(accountId: string): void {
    if (typeof window === 'undefined') return;
    
    const accounts = this.getSavedAccounts();
    const filtered = accounts.filter(a => a.id !== accountId);
    
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(filtered));
    } catch (error) {
      console.error('Failed to remove account:', error);
    }
  }

  /**
   * Get a specific saved account by ID
   */
  static getAccount(accountId: string): SavedAccount | null {
    const accounts = this.getSavedAccounts();
    return accounts.find(a => a.id === accountId) || null;
  }

  /**
   * Clear all saved accounts
   */
  static clearAll(): void {
    if (typeof window === 'undefined') return;
    
    try {
      localStorage.removeItem(STORAGE_KEY);
    } catch (error) {
      console.error('Failed to clear accounts:', error);
    }
  }

  /**
   * Check if an account is currently logged in
   */
  static isAccountActive(accountId: string): boolean {
    if (typeof window === 'undefined') return false;
    
    try {
      const currentUser = localStorage.getItem('user');
      if (!currentUser) return false;
      
      const user = JSON.parse(currentUser);
      return user.id === accountId;
    } catch {
      return false;
    }
  }
}
