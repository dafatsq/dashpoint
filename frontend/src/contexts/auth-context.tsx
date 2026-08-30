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
  readRememberScope,
  refreshSessionUser,
  writeRememberScope,
} from "@/lib/auth-session";
import { IS_DESKTOP_BUILD } from "@/lib/config";
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
    let cancelled = false;

    if (IS_DESKTOP_BUILD) {
      // Desktop (Wails) keeps the legacy storage-backed session.
      const storedUser = loadStoredUser();
      setUser(storedUser);
      setIsLoading(false);
      if (storedUser && !hasBootstrappedRefreshRef.current) {
        hasBootstrappedRefreshRef.current = true;
        requestUserRefresh();
      }
      return;
    }

    // Web: the access token only exists in memory, so every page load starts
    // empty and the httpOnly refresh cookie decides whether a session
    // survived the reload.
    void (async () => {
      // Hard cap the boot spinner — a hung request must never trap the user
      // on a loading screen.
      const failsafe = window.setTimeout(() => {
        if (cancelled) return;
        console.warn("Auth bootstrap timed out; releasing loading state.");
        setUser(null);
        setIsLoading(false);
      }, 12000);
      try {
        const refreshedUser = await refreshSessionUser();
        if (cancelled) return;
        setUser(refreshedUser);
        setIsLoading(false);
        if (refreshedUser && !hasBootstrappedRefreshRef.current) {
          hasBootstrappedRefreshRef.current = true;
          requestUserRefresh();
        }
      } finally {
        window.clearTimeout(failsafe);
      }
    })();

    return () => {
      cancelled = true;
    };
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
    async (email: string, password: string, saveAccount?: boolean) => {
      // Automatic sign-in requires BOTH the device scope (settings toggle)
      // and Save-login on this login. Unchecking "Save login" means this
      // device is not saved, so it also disables auto sign-in: the saved
      // account is removed, the device scope drops, and the refresh cookie
      // is session-scoped.
      // With saveAccount undefined (identity checks such as the settings
      // password verification) the current scope is re-sent unchanged: the
      // backend then mints a cookie with exactly the scope the user had.
      const remembered =
        saveAccount === undefined
          ? readRememberScope()
          : readRememberScope() && Boolean(saveAccount);
      const result = await api.login(email, password, remembered);
      if (result.error || !result.data) {
        return { success: false, error: result.error ?? "Login failed" };
      }

      if (saveAccount === false) {
        // Unchecked "Save login" = unsave this device: drop the saved
        // account and disable automatic sign-in so settings stay consistent
        // with the actual cookie scope.
        const { AccountManager } = await import("@/lib/account-manager");
        AccountManager.removeAccount(result.data.user?.id ?? "");
        writeRememberScope(false);
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
