'use client';

import { createContext, useContext, useEffect, useState, useCallback, ReactNode, useRef } from 'react';
import { User } from '@/types';
import api from '@/lib/api';
import { AccountManager } from '@/lib/account-manager';
import { useUserEvents, UserEvent } from '@/hooks/useUserEvents';
import { getSessionItem, setSessionItem, removeSessionItem, clearSession } from '@/lib/session';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isRealtimeConnected: boolean;
  login: (email: string, password: string, saveAccount?: boolean) => Promise<{ success: boolean; error?: string }>;
  pinLogin: (userId: string, pin: string) => Promise<{ success: boolean; error?: string }>;
  logout: (removeFromSaved?: boolean) => Promise<void>;
  refreshUser: () => Promise<void>;
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
  const isRefreshingRef = useRef(false);
  const isProcessingEventRef = useRef(false);

  // Refresh user data from server and optionally redirect on role change
  const refreshUser = useCallback(async (options?: { checkRoleChange?: boolean; previousRole?: string }) => {
    if (isRefreshingRef.current) return;
    isRefreshingRef.current = true;

    try {
      const token = getSessionItem('access_token');
      if (!token) {
        isRefreshingRef.current = false;
        return;
      }

      const result = await api.getMe();
      if (result.data) {
        const userData: User = {
          id: result.data.id,
          email: result.data.email || undefined,
          name: result.data.name,
          role_id: result.data.role_id || '',
          role_name: result.data.role_name as User['role_name'],
          is_active: result.data.is_active,
          has_pin: result.data.has_pin || false,
          permissions: result.data.permissions || [],
          created_at: result.data.created_at || '',
          updated_at: result.data.updated_at || '',
        };
        setSessionItem('user', JSON.stringify(userData));
        setUser(userData);

        // Update saved account info if has PIN
        if (userData.has_pin) {
          AccountManager.saveAccount({
            id: userData.id,
            name: userData.name,
            email: userData.email,
            role_name: userData.role_name,
            has_pin: userData.has_pin,
          });
        }

        // Check if role changed and redirect to dashboard if needed
        if (options?.checkRoleChange && options.previousRole && options.previousRole !== userData.role_name) {
          console.log('[Auth] Role changed from', options.previousRole, 'to', userData.role_name);
          // Redirect to dashboard to ensure user is on an accessible page
          if (typeof window !== 'undefined' && window.location.pathname !== '/') {
            console.log('[Auth] Redirecting to dashboard due to role change');
            window.location.href = '/';
          }
        }
      }
    } catch (error) {
      console.error('Failed to refresh user data:', error);
    } finally {
      isRefreshingRef.current = false;
    }
  }, []);

  // Handle user events from SSE
  const handleUserEvent = useCallback(async (event: UserEvent) => {
    console.log('[Auth] Received SSE event:', event.type, 'for user:', event.user_id);

    // Prevent processing multiple events at once (e.g., from multiple tabs)
    if (isProcessingEventRef.current) {
      console.log('[Auth] Already processing an event, skipping');
      return;
    }

    // Get current user from sessionStorage/localStorage since the closure might have stale data
    const storedUser = getSessionItem('user');
    if (!storedUser) {
      console.log('[Auth] No user in localStorage, ignoring event');
      return;
    }

    const currentUser = JSON.parse(storedUser);
    if (event.user_id !== currentUser.id) {
      console.log('[Auth] Ignoring event - not for current user');
      return;
    }

    isProcessingEventRef.current = true;
    console.log('[Auth] Processing event:', event.type);
    
    try {
      switch (event.type) {
        case 'role_changed':
          // Role changed - refresh tokens to get new JWT with updated role
          console.log('[Auth] Role changed, refreshing tokens to get new permissions');
          const tokenRefreshed = await api.refreshTokens();
          if (tokenRefreshed) {
            console.log('[Auth] Tokens refreshed successfully, forcing full page reload');
            if (typeof window !== 'undefined') {
              window.location.href = '/?role_updated=true';
            }
          } else {
            console.error('[Auth] Failed to refresh tokens after role change, forcing re-login');
            api.clearTokens();
            removeSessionItem('user');
            setUser(null);
            if (typeof window !== 'undefined') {
              window.location.href = '/login?message=role_changed_relogin';
            }
          }
          break;

        case 'user_updated':
        case 'permissions_changed':
        case 'user_activated':
          // Refresh user data to get the latest changes
          console.log('[Auth] Refreshing user data due to:', event.type);
          await refreshUser();
          break;

        case 'user_deactivated':
        case 'user_deleted':
        case 'force_logout':
          // Force logout the user
          api.clearTokens();
          removeSessionItem('user');
          setUser(null);
          if (typeof window !== 'undefined') {
            const message = event.type === 'user_deactivated'
              ? 'account_deactivated'
              : event.type === 'user_deleted'
                ? 'account_deleted'
                : 'force_logout';
            window.location.href = `/login?message=${message}`;
          }
          break;
      }
    } catch (error) {
      console.error('[Auth] Error processing background event:', error);
    } finally {
      isProcessingEventRef.current = false;
      console.log('[Auth] Event processing finished:', event.type);
    }
  }, [refreshUser]);

  // Subscribe to real-time user events
  const { isConnected: isRealtimeConnected } = useUserEvents({
    onAnyEvent: handleUserEvent,
    onConnected: () => {
      console.log('[Auth] SSE connected for real-time updates');
    },
    onDisconnected: () => {
      console.log('[Auth] SSE disconnected');
    },
    // Note: SSE "errors" are often just normal disconnects/reconnects
    // The actual error handling happens inside useUserEvents
    enabled: !!user,
  });

  // Check for existing session on mount
  useEffect(() => {
    const checkAuth = async () => {
      const token = getSessionItem('access_token');
      const storedUser = getSessionItem('user');

      if (token && storedUser) {
        try {
          const parsedUser = JSON.parse(storedUser);
          // Handle migration from old user format (role) to new format (role_name)
          if (parsedUser.role && !parsedUser.role_name) {
            // Old format - if role is a string, use it as role_name
            if (typeof parsedUser.role === 'string') {
              parsedUser.role_name = parsedUser.role;
            }
            // Save migrated user back to session storage
            setSessionItem('user', JSON.stringify(parsedUser));
          }
          setUser(parsedUser);
        } catch {
          removeSessionItem('user');
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
      setSessionItem('user', JSON.stringify(userData));
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
      setSessionItem('user', JSON.stringify(userData));
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

    await api.clearTokens();
    clearSession();
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
    isRealtimeConnected,
    login,
    pinLogin,
    logout,
    refreshUser,
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
  USERS_CREATE: 'can_create_user',
  USERS_EDIT: 'can_edit_user',
  USERS_DELETE: 'can_delete_user',
  USERS_PERMISSIONS: 'can_manage_permissions',

  // Product management
  PRODUCTS_VIEW: 'can_view_products',
  PRODUCTS_CREATE: 'can_create_product',
  PRODUCTS_EDIT: 'can_edit_product',
  PRODUCTS_DELETE: 'can_delete_product',

  // Inventory management
  INVENTORY_VIEW: 'can_view_inventory',
  INVENTORY_EDIT: 'can_edit_inventory',

  // Sales
  SALES_CREATE: 'can_create_sale',
  SALES_VIEW: 'can_view_sales',
  SALES_VOID: 'can_void_sale',

  // Reports
  REPORTS_VIEW: 'can_view_reports',
  REPORTS_EXPORT: 'can_export_data',

  // Expenses
  EXPENSES_VIEW: 'can_view_expenses',
  EXPENSES_MANAGE: 'can_manage_expenses',

  // Audit logs
  AUDIT_VIEW: 'can_view_audit_logs',

  // Settings - Removed: Everyone can access Settings page now
  // SETTINGS_MANAGE: 'can_manage_settings',

  // POS
  POS_VIEW: 'can_view_pos',
  POS_SHIFT_START: 'can_start_shift',
  POS_SHIFT_END: 'can_end_shift',

  // Categories
  CATEGORIES_VIEW: "can_view_categories",
  CATEGORIES_MANAGE: 'can_manage_categories',
} as const;
