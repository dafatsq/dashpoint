'use client';

import { useCallback, useEffect, useRef, useState } from "react";

import { getAccessToken, refreshSessionTokens } from "@/lib/auth-session";
import { API_BASE_URL } from "@/lib/config";

export type UserEventType =
  | "connected"
  | "user_updated"
  | "user_deactivated"
  | "user_activated"
  | "user_deleted"
  | "permissions_changed"
  | "role_changed"
  | "force_logout";

export interface UserEvent {
  type: UserEventType;
  user_id: string;
  changed_by?: string;
  timestamp: string;
  details?: Record<string, unknown>;
}

export interface UseUserEventsOptions {
  onUserUpdated?: (event: UserEvent) => void;
  onUserDeactivated?: (event: UserEvent) => void;
  onUserActivated?: (event: UserEvent) => void;
  onUserDeleted?: (event: UserEvent) => void;
  onPermissionsChanged?: (event: UserEvent) => void;
  onRoleChanged?: (event: UserEvent) => void;
  onForceLogout?: (event: UserEvent) => void;
  onAnyEvent?: (event: UserEvent) => void;
  onConnected?: () => void;
  onDisconnected?: () => void;
  onError?: (error: Event) => void;
  enabled?: boolean;
}

const USER_EVENT_TYPES: UserEventType[] = [
  "connected",
  "user_updated",
  "user_deactivated",
  "user_activated",
  "user_deleted",
  "permissions_changed",
  "role_changed",
  "force_logout",
];

export function useUserEvents(options: UseUserEventsOptions = {}) {
  const {
    onUserUpdated,
    onUserDeactivated,
    onUserActivated,
    onUserDeleted,
    onPermissionsChanged,
    onRoleChanged,
    onForceLogout,
    onAnyEvent,
    onConnected,
    onDisconnected,
    onError,
    enabled = true,
  } = options;

  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const connectRef = useRef<() => void>(() => {});
  const reconnectAttemptsRef = useRef(0);
  const enabledRef = useRef(enabled);
  const [isConnected, setIsConnected] = useState(false);

  const maxReconnectAttempts = 5;
  const baseReconnectDelay = 1000;

  const callbacksRef = useRef({
    onUserUpdated,
    onUserDeactivated,
    onUserActivated,
    onUserDeleted,
    onPermissionsChanged,
    onRoleChanged,
    onForceLogout,
    onAnyEvent,
    onConnected,
    onDisconnected,
    onError,
  });

  useEffect(() => {
    enabledRef.current = enabled;
  }, [enabled]);

  useEffect(() => {
    callbacksRef.current = {
      onUserUpdated,
      onUserDeactivated,
      onUserActivated,
      onUserDeleted,
      onPermissionsChanged,
      onRoleChanged,
      onForceLogout,
      onAnyEvent,
      onConnected,
      onDisconnected,
      onError,
    };
  }, [
    onAnyEvent,
    onConnected,
    onDisconnected,
    onError,
    onForceLogout,
    onPermissionsChanged,
    onRoleChanged,
    onUserActivated,
    onUserDeactivated,
    onUserDeleted,
    onUserUpdated,
  ]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    setIsConnected(false);
  }, []);

  const scheduleReconnect = useCallback(() => {
    if (
      reconnectAttemptsRef.current >= maxReconnectAttempts ||
      !enabledRef.current
    ) {
      return;
    }

    const delay = baseReconnectDelay * 2 ** reconnectAttemptsRef.current;
    reconnectAttemptsRef.current += 1;

    reconnectTimeoutRef.current = setTimeout(() => {
      if (enabledRef.current && getAccessToken()) {
        connectRef.current();
      }
    }, delay);
  }, []);

  const dispatchEvent = useCallback((event: UserEvent) => {
    switch (event.type) {
      case "user_updated":
        callbacksRef.current.onUserUpdated?.(event);
        break;
      case "user_deactivated":
        callbacksRef.current.onUserDeactivated?.(event);
        break;
      case "user_activated":
        callbacksRef.current.onUserActivated?.(event);
        break;
      case "user_deleted":
        callbacksRef.current.onUserDeleted?.(event);
        break;
      case "permissions_changed":
        callbacksRef.current.onPermissionsChanged?.(event);
        break;
      case "role_changed":
        callbacksRef.current.onRoleChanged?.(event);
        break;
      case "force_logout":
        callbacksRef.current.onForceLogout?.(event);
        break;
      default:
        break;
    }

    callbacksRef.current.onAnyEvent?.(event);
  }, []);

  const connect = useCallback(() => {
    const token = getAccessToken();
    if (!token || !enabledRef.current) return;

    if (
      eventSourceRef.current &&
      eventSourceRef.current.readyState === EventSource.OPEN
    ) {
      return;
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    fetch(`${API_BASE_URL}/me`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then(async (response) => {
        let validToken = token;

        if (response.status === 401) {
          const refreshed = await refreshSessionTokens();
          if (!refreshed) {
            callbacksRef.current.onError?.(new Event("Token refresh failed"));
            return;
          }

          const nextToken = getAccessToken();
          if (!nextToken) {
            callbacksRef.current.onError?.(new Event("Token refresh failed"));
            return;
          }
          validToken = nextToken;
        } else if (!response.ok) {
          callbacksRef.current.onError?.(new Event("Token validation failed"));
          return;
        }

        const eventSource = new EventSource(
          `${API_BASE_URL}/events/subscribe?token=${encodeURIComponent(validToken)}`,
        );
        eventSourceRef.current = eventSource;

        eventSource.onopen = () => {
          setIsConnected(true);
          reconnectAttemptsRef.current = 0;
          callbacksRef.current.onConnected?.();
        };

        eventSource.onerror = () => {
          setIsConnected(false);
          callbacksRef.current.onDisconnected?.();
          eventSource.close();
          eventSourceRef.current = null;
          scheduleReconnect();
        };

        USER_EVENT_TYPES.forEach((eventType) => {
          eventSource.addEventListener(eventType, (messageEvent: MessageEvent) => {
            try {
              const event = JSON.parse(messageEvent.data) as UserEvent;
              dispatchEvent(event);
            } catch (error) {
              console.error("[SSE] Failed to parse event:", error);
            }
          });
        });
      })
      .catch(() => {
        callbacksRef.current.onError?.(new Event("Pre-flight check failed"));
      });
  }, [dispatchEvent, scheduleReconnect]);

  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  useEffect(() => {
    if (enabled) {
      connectRef.current();
    }

    return () => {
      disconnect();
    };
  }, [disconnect, enabled]);

  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState !== "visible" || !enabledRef.current) {
        return;
      }

      setTimeout(() => {
        if (!eventSourceRef.current && getAccessToken()) {
          reconnectAttemptsRef.current = 0;
          connectRef.current();
        }
      }, 100);
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, []);

  return {
    isConnected,
    connect,
    disconnect,
  };
}

export default useUserEvents;
