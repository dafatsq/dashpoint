// @vitest-environment jsdom

import { act, createElement } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";

import { AuthProvider, useAuth, PERMISSIONS } from "./auth-context";

import type { User } from "@/types";
import type { UseUserEventsOptions, UserEvent } from "@/hooks/useUserEvents";

const {
  getMeMock,
  loadStoredUserMock,
  persistAuthUserMock,
  latestUserEventsOptions,
} = vi.hoisted(() => ({
  getMeMock: vi.fn(),
  loadStoredUserMock: vi.fn(),
  persistAuthUserMock: vi.fn((user: User) => user),
  latestUserEventsOptions: {} as { current?: UseUserEventsOptions },
}));

vi.mock("@/lib/api", () => ({
  default: {
    getMe: getMeMock,
    refreshTokens: vi.fn(),
    logout: vi.fn(),
    login: vi.fn(),
    pinLogin: vi.fn(),
  },
}));

vi.mock("@/lib/auth-session", () => ({
  clearAuthSession: vi.fn(),
  loadStoredUser: loadStoredUserMock,
  persistAuthPayload: vi.fn(),
  persistAuthUser: persistAuthUserMock,
}));

vi.mock("@/hooks/useUserEvents", () => ({
  useUserEvents: (options: UseUserEventsOptions = {}) => {
    latestUserEventsOptions.current = options;
    return { isConnected: true, connect: vi.fn(), disconnect: vi.fn() };
  },
}));

function createUser(overrides: Partial<User> = {}): User {
  return {
    id: "user-1",
    email: "manager@dashpoint.local",
    name: "Test Manager",
    role_id: "role-manager",
    role_name: "manager",
    is_active: true,
    has_pin: true,
    permissions: [],
    created_at: "2026-05-16T00:00:00Z",
    updated_at: "2026-05-16T00:00:00Z",
    ...overrides,
  };
}

function AuthProbe() {
  const { user, isLoading, hasPermission } = useAuth();

  return createElement(
    "div",
    {
      "data-loading": String(isLoading),
      "data-user-id": user?.id ?? "",
      "data-can-view-pos": String(hasPermission(PERMISSIONS.POS_VIEW)),
      "data-can-start-shift": String(
        hasPermission(PERMISSIONS.POS_SHIFT_START),
      ),
      "data-can-end-shift": String(hasPermission(PERMISSIONS.POS_SHIFT_END)),
    },
    user?.permissions?.join(",") ?? "",
  );
}

describe("AuthProvider", () => {
  let container: HTMLDivElement;
  let root: Root;

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);

    getMeMock.mockReset();
    loadStoredUserMock.mockReset();
    persistAuthUserMock.mockClear();
    latestUserEventsOptions.current = undefined;
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    container.remove();
  });

  async function renderProvider() {
    await act(async () => {
      root.render(
        createElement(
          AuthProvider,
          null,
          createElement(AuthProbe),
        ),
      );
    });
  }

  async function flushUpdates() {
    await act(async () => {
      await Promise.resolve();
    });
  }

  test("refreshes the stored user from /me on bootstrap so a page refresh picks up permission changes", async () => {
    loadStoredUserMock.mockReturnValue(
      createUser({
        permissions: [PERMISSIONS.POS_VIEW, PERMISSIONS.POS_SHIFT_START],
      }),
    );
    getMeMock.mockResolvedValue({
      data: createUser({
        permissions: [PERMISSIONS.POS_VIEW],
      }),
    });

    await renderProvider();
    await flushUpdates();

    const probe = container.querySelector("div[data-user-id]");
    expect(getMeMock).toHaveBeenCalledOnce();
    expect(probe?.getAttribute("data-can-view-pos")).toBe("true");
    expect(probe?.getAttribute("data-can-start-shift")).toBe("false");
    expect(probe?.getAttribute("data-can-end-shift")).toBe("false");
    expect(probe?.textContent).toBe(PERMISSIONS.POS_VIEW);
  });

  test("refreshes permissions immediately when a permissions_changed user event arrives", async () => {
    loadStoredUserMock.mockReturnValue(
      createUser({
        permissions: [PERMISSIONS.POS_VIEW, PERMISSIONS.POS_SHIFT_START],
      }),
    );
    getMeMock
      .mockResolvedValueOnce({
        data: createUser({
          permissions: [PERMISSIONS.POS_VIEW, PERMISSIONS.POS_SHIFT_START],
        }),
      })
      .mockResolvedValueOnce({
        data: createUser({
          permissions: [PERMISSIONS.POS_VIEW],
        }),
      });

    await renderProvider();
    await flushUpdates();

    const event: UserEvent = {
      type: "permissions_changed",
      user_id: "user-1",
      timestamp: new Date().toISOString(),
    };

    await act(async () => {
      await latestUserEventsOptions.current?.onAnyEvent?.(event);
    });
    await flushUpdates();

    const probe = container.querySelector("div[data-user-id]");
    expect(getMeMock).toHaveBeenCalledTimes(2);
    expect(probe?.getAttribute("data-can-view-pos")).toBe("true");
    expect(probe?.getAttribute("data-can-start-shift")).toBe("false");
    expect(probe?.textContent).toBe(PERMISSIONS.POS_VIEW);
  });

  test("revalidates the current user on window focus as a fallback for cross-session permission changes", async () => {
    loadStoredUserMock.mockReturnValue(
      createUser({
        permissions: [PERMISSIONS.POS_VIEW, PERMISSIONS.POS_SHIFT_START],
      }),
    );
    getMeMock
      .mockResolvedValueOnce({
        data: createUser({
          permissions: [PERMISSIONS.POS_VIEW, PERMISSIONS.POS_SHIFT_START],
        }),
      })
      .mockResolvedValueOnce({
        data: createUser({
          permissions: [PERMISSIONS.POS_VIEW],
        }),
      });

    await renderProvider();
    await flushUpdates();

    await act(async () => {
      window.dispatchEvent(new Event("focus"));
    });
    await flushUpdates();

    const probe = container.querySelector("div[data-user-id]");
    expect(getMeMock).toHaveBeenCalledTimes(2);
    expect(probe?.getAttribute("data-can-view-pos")).toBe("true");
    expect(probe?.getAttribute("data-can-start-shift")).toBe("false");
    expect(probe?.textContent).toBe(PERMISSIONS.POS_VIEW);
  });
});
