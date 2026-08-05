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

  const abortControllerRef = useRef<AbortController | null>(null);
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

    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
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

  const handleStreamData = useCallback(
    (buffer: string) => {
      const segments = buffer.split("\n\n");
      const remainder = segments.pop() ?? "";

      for (const segment of segments) {
        const lines = segment.split("\n");
        let eventType = "";
        const dataLines: string[] = [];

        for (const line of lines) {
          if (!line || line.startsWith(":")) continue;
          if (line.startsWith("event:")) {
            eventType = line.slice("event:".length).trim();
            continue;
          }
          if (line.startsWith("data:")) {
            dataLines.push(line.slice("data:".length).trimStart());
          }
        }

        if (!eventType || dataLines.length === 0) {
          continue;
        }

        if (!USER_EVENT_TYPES.includes(eventType as UserEventType)) {
          continue;
        }

        try {
          const event = JSON.parse(dataLines.join("\n")) as UserEvent;
          dispatchEvent(event);
        } catch (error) {
          console.error("[SSE] Failed to parse event:", error);
        }
      }

      return remainder;
    },
    [dispatchEvent],
  );

  const connect = useCallback(() => {
    const token = getAccessToken();
    if (!token || !enabledRef.current) return;

    if (abortControllerRef.current) {
      return;
    }

    const controller = new AbortController();
    abortControllerRef.current = controller;

    const openStream = async () => {
      let validToken = token;

      const connectResponse = await fetch(`${API_BASE_URL}/events/subscribe`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${validToken}`,
          Accept: "text/event-stream",
          "Cache-Control": "no-cache",
        },
        credentials: "include",
        signal: controller.signal,
      });

      if (connectResponse.status === 401) {
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

        const retryResponse = await fetch(`${API_BASE_URL}/events/subscribe`, {
          method: "GET",
          headers: {
            Authorization: `Bearer ${validToken}`,
            Accept: "text/event-stream",
            "Cache-Control": "no-cache",
          },
          credentials: "include",
          signal: controller.signal,
        });

        return retryResponse;
      }

      return connectResponse;
    };

    openStream()
      .then(async (response) => {
        if (!response) {
          abortControllerRef.current = null;
          return;
        }
        if (!response.ok || !response.body) {
          abortControllerRef.current = null;
          callbacksRef.current.onError?.(new Event("Token validation failed"));
          return;
        }

        setIsConnected(true);
        reconnectAttemptsRef.current = 0;
        callbacksRef.current.onConnected?.();

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) {
            break;
          }

          buffer += decoder.decode(value, { stream: true });
          buffer = handleStreamData(buffer);
        }

        if (!controller.signal.aborted) {
          setIsConnected(false);
          callbacksRef.current.onDisconnected?.();
          abortControllerRef.current = null;
          scheduleReconnect();
        }
      })
      .catch((error) => {
        abortControllerRef.current = null;
        if (controller.signal.aborted) {
          return;
        }
        setIsConnected(false);
        callbacksRef.current.onDisconnected?.();
        callbacksRef.current.onError?.(
          error instanceof Event ? error : new Event("Stream connection failed"),
        );
        scheduleReconnect();
      });
  }, [handleStreamData, scheduleReconnect]);

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
        if (!abortControllerRef.current && getAccessToken()) {
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
