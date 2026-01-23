'use client';

import { createContext, useContext, useEffect, useState, useCallback, ReactNode } from 'react';
import { User } from '@/types';
import api from '@/lib/api';
import { AccountManager } from '@/lib/account-manager';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string, saveAccount?: boolean) => Promise<{ success: boolean; error?: string }>;
  pinLogin: (userId: string, pin: string) => Promise<{ success: boolean; error?: string }>;
  logout: (removeFromSaved?: boolean) => Promise<void>;
  hasPermission: (permission: string) => boolean;
  hasAnyPermission: (permissions: string[]) => boolean;
  hasAllPermissions: (permissions: string[]) => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Check for existing session on mount
  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem('access_token');
      const storedUser = localStorage.getItem('user');
      
      if (token && storedUser) {
        try {
          const parsedUser = JSON.parse(storedUser);
          // Handle migration from old user format (role) to new format (role_name)
          if (parsedUser.role && !parsedUser.role_name) {
            // Old format - if role is a string, use it as role_name
            if (typeof parsedUser.role === 'string') {
              parsedUser.role_name = parsedUser.role;
            }
            // Save migrated user back to localStorage
            localStorage.setItem('user', JSON.stringify(parsedUser));
          }
          setUser(parsedUser);
        } catch {
          localStorage.removeItem('user');
        }
      }
      setIsLoading(false);
    };

    checkAuth();
  }, []);

  const login = useCallback(async (email: string, password: string, saveAccount = true) => {
    const result = await api.login(email, password);
    
    if (result.error) {
      return { success: false, error: result.error };
    }

    if (result.data) {
      api.setTokens(result.data.access_token, result.data.refresh_token);
      // Build User object from API response
      const userData: User = {
        id: result.data.user.id,
        email: result.data.user.email || undefined,
        name: result.data.user.name,
        role_id: result.data.user.role_id || '',
        role_name: result.data.user.role_name as User['role_name'],
        is_active: result.data.user.is_active,
        has_pin: result.data.user.has_pin || false,
        permissions: result.data.user.permissions || [],
        created_at: result.data.user.created_at || '',
        updated_at: result.data.user.updated_at || '',
      };
      localStorage.setItem('user', JSON.stringify(userData));
      setUser(userData);
      
      // Save account for quick switching (if enabled and user has PIN)
      if (saveAccount && userData.has_pin) {
        AccountManager.saveAccount({
          id: userData.id,
          name: userData.name,
          email: userData.email,
          role_name: userData.role_name,
          has_pin: userData.has_pin,
        });
      }
      
      return { success: true };
    }

    return { success: false, error: 'Login failed' };
  }, []);

  const pinLogin = useCallback(async (userId: string, pin: string) => {
    const result = await api.pinLogin(userId, pin);
    
    if (result.error) {
      return { success: false, error: result.error };
    }

    if (result.data) {
      api.setTokens(result.data.access_token, result.data.refresh_token);
      // Build User object from API response
      const userData: User = {
        id: result.data.user.id,
        email: result.data.user.email || undefined,
        name: result.data.user.name,
        role_id: result.data.user.role_id || '',
        role_name: result.data.user.role_name as User['role_name'],
        is_active: result.data.user.is_active,
        has_pin: result.data.user.has_pin || false,
        permissions: result.data.user.permissions || [],
        created_at: result.data.user.created_at || '',
        updated_at: result.data.user.updated_at || '',
      };
      localStorage.setItem('user', JSON.stringify(userData));
      setUser(userData);
      
      // Update saved account info
      if (userData.has_pin) {
        AccountManager.saveAccount({
          id: userData.id,
          name: userData.name,
          email: userData.email,
          role_name: userData.role_name,
          has_pin: userData.has_pin,
        });
      }
      
      return { success: true };
    }

    return { success: false, error: 'PIN login failed' };
  }, []);

  const logout = useCallback(async (removeFromSaved = false) => {
    const currentUserId = user?.id;
    
    await api.logout();
    localStorage.removeItem('user');
    setUser(null);
    
    // Optionally remove from saved accounts
    if (removeFromSaved && currentUserId) {
      AccountManager.removeAccount(currentUserId);
    }
  }, [user]);

  const hasPermission = useCallback((permission: string) => {
    if (!user) return false;
    // Owner has all permissions
    if (user.role_name === 'owner') return true;
    return user.permissions?.includes(permission) ?? false;
  }, [user]);

  const hasAnyPermission = useCallback((permissions: string[]) => {
    if (!user) return false;
    if (user.role_name === 'owner') return true;
    return permissions.some(p => user.permissions?.includes(p));
  }, [user]);

  const hasAllPermissions = useCallback((permissions: string[]) => {
    if (!user) return false;
    if (user.role_name === 'owner') return true;
    return permissions.every(p => user.permissions?.includes(p));
  }, [user]);

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated: !!user,
    login,
    pinLogin,
    logout,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// Permission constants for easy reference
// These must match the 'key' column in the database 'permissions' table
export const PERMISSIONS = {
  // User management
  USERS_VIEW: 'can_view_users',
  USERS_MANAGE: 'can_manage_users',
  USERS_PERMISSIONS: 'can_manage_permissions',
  
  // Product management
  PRODUCTS_VIEW: 'can_view_products',
  PRODUCTS_CREATE: 'can_create_product',
  PRODUCTS_EDIT: 'can_edit_product',
  PRODUCTS_DELETE: 'can_delete_product',
  PRODUCTS_EDIT_PRICE: 'can_edit_price',
  
  // Inventory management
  INVENTORY_VIEW: 'can_view_inventory',
  INVENTORY_ADJUST: 'can_adjust_stock',
  INVENTORY_RECEIVE: 'can_receive_stock',
  
  // Sales
  SALES_CREATE: 'can_create_sale',
  SALES_VOID: 'can_void_sale',
  SALES_DISCOUNT: 'can_apply_discount',
  SALES_REFUND: 'can_refund',
  
  // Reports
  REPORTS_VIEW: 'can_view_reports',
  REPORTS_EXPORT: 'can_export_data',
  
  // Audit logs
  AUDIT_VIEW: 'can_view_audit_logs',
  
  // Settings
  SETTINGS_MANAGE: 'can_manage_settings',
  SETTINGS_BACKUP: 'can_backup_data',
} as const;
