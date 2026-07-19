'use client';

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";

import { useUserEvents, type UserEvent } from "@/hooks/useUserEvents";
import api from "@/lib/api";
import {
  hasAllPermissions,
  hasAnyPermission,
  hasPermission,
  PERMISSIONS,
} from "@/lib/auth-permissions";
import {
  clearAuthSession,
  loadStoredUser,
  persistAuthPayload,
  persistAuthUser,
} from "@/lib/auth-session";
import type { AuthPayload } from "@/lib/auth-user";
import type { User } from "@/types";

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isRealtimeConnected: boolean;
  login: (
    email: string,
    password: string,
    saveAccount?: boolean,
  ) => Promise<{ success: boolean; error?: string }>;
  pinLogin: (
    userId: string,
    pin: string,
  ) => Promise<{ success: boolean; error?: string }>;
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

function resolveLogoutMessage(eventType: UserEvent["type"]): string {
  switch (eventType) {
    case "user_deactivated":
      return "account_deactivated";
    case "user_deleted":
      return "account_deleted";
    default:
      return "force_logout";
  }
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const isRefreshingRef = useRef(false);
  const isProcessingEventRef = useRef(false);
  const hasBootstrappedRefreshRef = useRef(false);

  const applyAuthPayload = useCallback(
    (payload: AuthPayload, options?: { saveAccount?: boolean }) => {
      const nextUser = persistAuthPayload(payload, options);
      setUser(nextUser);
      return nextUser;
    },
    [],
  );

  const refreshUser = useCallback(
    async (options?: { checkRoleChange?: boolean; previousRole?: string }) => {
      if (isRefreshingRef.current) return;
      isRefreshingRef.current = true;

      try {
        const result = await api.getMe();
        if (result.error) return;
        if (!result.data) return;

        const nextUser = persistAuthUser(result.data);
        setUser(nextUser);

        if (
          options?.checkRoleChange &&
          options.previousRole &&
          options.previousRole !== nextUser.role_name &&
          typeof window !== "undefined" &&
          window.location.pathname !== "/"
        ) {
          window.location.href = "/";
        }
      } catch (error) {
        console.error("Failed to refresh user data:", error);
      } finally {
        isRefreshingRef.current = false;
      }
    },
    [],
  );

  const requestUserRefresh = useCallback(() => {
    void refreshUser();
  }, [refreshUser]);

  const forceLogout = useCallback((message?: string) => {
    clearAuthSession();
    setUser(null);

    if (typeof window !== "undefined") {
      const query = message ? `?message=${message}` : "";
      window.location.href = `/login${query}`;
    }
  }, []);

  const handleUserEvent = useCallback(
    async (event: UserEvent) => {
      if (isProcessingEventRef.current) return;
      if (!user || event.user_id !== user.id) return;

      isProcessingEventRef.current = true;

      try {
        switch (event.type) {
          case "role_changed": {
            const tokenRefreshed = await api.refreshTokens();
            if (tokenRefreshed) {
              if (typeof window !== "undefined") {
                window.location.href = "/?role_updated=true";
              }
            } else {
              forceLogout("role_changed_relogin");
            }
            break;
          }
          case "user_updated":
          case "permissions_changed":
          case "user_activated":
            await refreshUser();
            break;
          case "user_deactivated":
          case "user_deleted":
          case "force_logout":
            forceLogout(resolveLogoutMessage(event.type));
            break;
          default:
            break;
        }
      } catch (error) {
        console.error("[Auth] Error processing background event:", error);
      } finally {
        isProcessingEventRef.current = false;
      }
    },
    [forceLogout, refreshUser, user],
  );

  const { isConnected: isRealtimeConnected } = useUserEvents({
    onAnyEvent: handleUserEvent,
    enabled: !!user,
  });

  useEffect(() => {
    const storedUser = loadStoredUser();
    setUser(storedUser);
    setIsLoading(false);

    if (storedUser && !hasBootstrappedRefreshRef.current) {
      hasBootstrappedRefreshRef.current = true;
      requestUserRefresh();
    }
  }, [requestUserRefresh]);

  useEffect(() => {
    if (!user) return;

    const handleWindowFocus = () => {
      requestUserRefresh();
    };

    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        requestUserRefresh();
      }
    };

    window.addEventListener("focus", handleWindowFocus);
    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      window.removeEventListener("focus", handleWindowFocus);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [requestUserRefresh, user]);

  const login = useCallback(
    async (email: string, password: string, saveAccount = true) => {
      const result = await api.login(email, password);
      if (result.error || !result.data) {
        return { success: false, error: result.error ?? "Login failed" };
      }

      applyAuthPayload(result.data, { saveAccount });
      return { success: true };
    },
    [applyAuthPayload],
  );

  const pinLogin = useCallback(
    async (userId: string, pin: string) => {
      const result = await api.pinLogin(userId, pin);
      if (result.error || !result.data) {
        return { success: false, error: result.error ?? "PIN login failed" };
      }

      applyAuthPayload(result.data);
      return { success: true };
    },
    [applyAuthPayload],
  );

  const logout = useCallback(
    async (removeFromSaved = false) => {
      const currentUserId = user?.id;

      await api.logout();
      clearAuthSession();
      setUser(null);

      if (removeFromSaved && currentUserId) {
        const { AccountManager } = await import("@/lib/account-manager");
        AccountManager.removeAccount(currentUserId);
      }
    },
    [user],
  );

  const value = useMemo<AuthContextType>(
    () => ({
      user,
      isLoading,
      isAuthenticated: !!user,
      isRealtimeConnected,
      login,
      pinLogin,
      logout,
      refreshUser,
      hasPermission: (permission) => hasPermission(user, permission),
      hasAnyPermission: (permissions) => hasAnyPermission(user, permissions),
      hasAllPermissions: (permissions) => hasAllPermissions(user, permissions),
    }),
    [
      isLoading,
      isRealtimeConnected,
      login,
      logout,
      pinLogin,
      refreshUser,
      user,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

export { PERMISSIONS };
