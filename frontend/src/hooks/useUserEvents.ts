'use client';

import { useEffect, useRef, useCallback, useState } from 'react';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

// Helper function to refresh the access token
async function tryRefreshToken(): Promise<string | null> {
  const refreshToken = localStorage.getItem('refresh_token');
  if (!refreshToken) return null;

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      return null;
    }

    const data = await response.json();
    localStorage.setItem('access_token', data.access_token);
    localStorage.setItem('refresh_token', data.refresh_token);
    return data.access_token;
  } catch {
    return null;
  }
}

export type UserEventType =
  | 'connected'
  | 'user_updated'
  | 'user_deactivated'
  | 'user_activated'
  | 'user_deleted'
  | 'permissions_changed'
  | 'role_changed'
  | 'force_logout';

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
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;
  const baseReconnectDelay = 1000;
  const enabledRef = useRef(enabled);

  // Store callbacks in refs to avoid dependency issues
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

  // Keep enabled ref in sync
  useEffect(() => {
    enabledRef.current = enabled;
  }, [enabled]);

  // Update refs when callbacks change
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
  }, [onUserUpdated, onUserDeactivated, onUserActivated, onUserDeleted, onPermissionsChanged, onRoleChanged, onForceLogout, onAnyEvent, onConnected, onDisconnected, onError]);

  const getAccessToken = useCallback(() => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('access_token');
  }, []);

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

  // Use a ref to store the connect function so we can call it recursively
  const connectRef = useRef<() => void>(() => {});

  const connect = useCallback(() => {
    const token = getAccessToken();
    console.log('[SSE] Connect called, token:', token ? 'present' : 'missing', 'enabled:', enabledRef.current);
    if (!token || !enabledRef.current) {
      console.log('[SSE] Skipping connection - no token or not enabled');
      return;
    }

    // Don't reconnect if already connected
    if (eventSourceRef.current && eventSourceRef.current.readyState === EventSource.OPEN) {
      console.log('[SSE] Already connected, skipping');
      return;
    }

    // Close any existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    const url = `${API_BASE_URL}/events/subscribe?token=${encodeURIComponent(token)}`;
    console.log('[SSE] Connecting to:', API_BASE_URL + '/events/subscribe');
    
    // Pre-flight check: verify token is valid by making a quick HEAD request
    // This helps diagnose issues before EventSource silently fails
    fetch(`${API_BASE_URL}/me`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }).then(async response => {
      let validToken = token;
      
      if (response.status === 401) {
        console.log('[SSE] Token expired, attempting refresh...');
        const newToken = await tryRefreshToken();
        if (newToken) {
          console.log('[SSE] Token refreshed successfully');
          validToken = newToken;
        } else {
          console.error('[SSE] Token refresh failed');
          callbacksRef.current.onError?.(new Event('Token refresh failed'));
          return;
        }
      } else if (!response.ok) {
        console.error('[SSE] Token validation failed, status:', response.status);
        callbacksRef.current.onError?.(new Event('Token validation failed'));
        return;
      }
      
      console.log('[SSE] Token validated, establishing SSE connection...');
      
      const sseUrl = `${API_BASE_URL}/events/subscribe?token=${encodeURIComponent(validToken)}`;
      const eventSource = new EventSource(sseUrl);
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        console.log('[SSE] Connection opened successfully');
        setIsConnected(true);
        reconnectAttemptsRef.current = 0;
        callbacksRef.current.onConnected?.();
      };

      eventSource.onerror = () => {
        // Note: EventSource errors don't provide details due to browser security
        // ERR_INCOMPLETE_CHUNKED_ENCODING is normal when navigating away
        const readyState = eventSource.readyState;
        const isClosed = readyState === EventSource.CLOSED;
        
        if (isClosed) {
          // This is expected when navigating - don't log as warning
          console.log('[SSE] Connection closed');
        } else {
          // Only log actual errors, not navigation-related disconnects
          console.log('[SSE] Connection interrupted, will attempt reconnect');
        }
        
        setIsConnected(false);
        callbacksRef.current.onDisconnected?.();

        // Close the connection on error
        eventSource.close();
        eventSourceRef.current = null;

        // Attempt to reconnect with exponential backoff
        if (reconnectAttemptsRef.current < maxReconnectAttempts && enabledRef.current) {
          const delay = baseReconnectDelay * Math.pow(2, reconnectAttemptsRef.current);
          reconnectAttemptsRef.current += 1;
          
          reconnectTimeoutRef.current = setTimeout(() => {
            // Re-check if still enabled before reconnecting
            if (enabledRef.current && getAccessToken()) {
              connectRef.current();
            }
          }, delay);
        } else {
          console.log('[SSE] Max reconnect attempts reached or not enabled, giving up');
        }
      };

      // Listen for different event types
      const eventTypes: UserEventType[] = [
        'connected',
        'user_updated',
        'user_deactivated',
        'user_activated',
        'user_deleted',
        'permissions_changed',
        'role_changed',
        'force_logout',
      ];

      eventTypes.forEach((eventType) => {
        eventSource.addEventListener(eventType, (e: MessageEvent) => {
          try {
            const event: UserEvent = JSON.parse(e.data);
            console.log('[SSE] Received event:', event.type, 'for user:', event.user_id);
            console.log('[SSE] Event data:', JSON.stringify(event));
            
            // Call specific handler
            switch (event.type) {
              case 'connected':
                // Already handled by onopen
                break;
              case 'user_updated':
                console.log('[SSE] Calling onUserUpdated handler');
                callbacksRef.current.onUserUpdated?.(event);
                break;
              case 'user_deactivated':
                console.log('[SSE] Calling onUserDeactivated handler');
                callbacksRef.current.onUserDeactivated?.(event);
                break;
              case 'user_activated':
                console.log('[SSE] Calling onUserActivated handler');
                callbacksRef.current.onUserActivated?.(event);
                break;
              case 'user_deleted':
                console.log('[SSE] Calling onUserDeleted handler');
                callbacksRef.current.onUserDeleted?.(event);
                break;
              case 'permissions_changed':
                console.log('[SSE] Calling onPermissionsChanged handler');
                callbacksRef.current.onPermissionsChanged?.(event);
                break;
              case 'role_changed':
                console.log('[SSE] Calling onRoleChanged handler');
                callbacksRef.current.onRoleChanged?.(event);
                break;
              case 'force_logout':
                console.log('[SSE] Calling onForceLogout handler');
                callbacksRef.current.onForceLogout?.(event);
                break;
            }

            // Call generic handler for all events
            console.log('[SSE] Calling onAnyEvent handler');
            callbacksRef.current.onAnyEvent?.(event);
          } catch (err) {
            console.error('[SSE] Failed to parse event:', err);
          }
        });
      });
    }).catch(err => {
      console.error('[SSE] Pre-flight token check failed:', err);
      callbacksRef.current.onError?.(new Event('Pre-flight check failed'));
    });
  }, [getAccessToken]);

  // Keep the connect ref in sync
  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  // Connect on mount, disconnect on unmount
  useEffect(() => {
    if (enabled) {
      // Use ref to avoid dependency on connect function
      connectRef.current();
    }
    // Cleanup function handles disconnect
    return () => {
      // Only disconnect on unmount or when enabled changes to false
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
    };
  }, [enabled]); // Remove connect from dependencies - use ref instead

  // Reconnect when visibility changes (tab becomes visible again)
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible' && enabledRef.current) {
        // Small delay to ensure token is still valid
        setTimeout(() => {
          if (!eventSourceRef.current && getAccessToken()) {
            reconnectAttemptsRef.current = 0;
            connectRef.current();
          }
        }, 100);
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [getAccessToken]);

  return {
    isConnected,
    connect,
    disconnect,
  };
}

export default useUserEvents;
