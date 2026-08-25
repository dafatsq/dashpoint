// @vitest-environment jsdom

import { act, createElement } from "react";
import { createRoot, type Root } from "react-dom/client";
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  test,
  vi,
} from "vitest";

import { LoginScreen } from "./login-screen";

const routerPush = vi.fn();
const loginMock = vi.fn();
const searchParamsState = {
  message: null as string | null,
};

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: routerPush,
  }),
  useSearchParams: () => ({
    get: (key: string) => (key === "message" ? searchParamsState.message : null),
  }),
}));

vi.mock("@/contexts/auth-context", () => ({
  useAuth: () => ({
    login: loginMock,
    isLoading: false,
  }),
}));

vi.mock("@/components/account-switcher", () => ({
  AccountSwitcher: ({
    onAccountsChange,
  }: {
    onAccountsChange?: () => void;
  }) =>
    createElement(
      "button",
      {
        type: "button",
        onClick: () => {
          window.localStorage.setItem("dashpoint_saved_accounts", "[]");
          onAccountsChange?.();
        },
      },
      "Remove Saved Account",
    ),
}));

describe("LoginScreen", () => {
  let container: HTMLDivElement;
  let root: Root;
  let originalDemoAccessEnv: string | undefined;

  beforeEach(() => {
    // Radix components expect these browser APIs in jsdom tests.
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    globalThis.ResizeObserver = class ResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
    window.localStorage.clear();
    routerPush.mockReset();
    loginMock.mockReset();
    searchParamsState.message = null;
    originalDemoAccessEnv = process.env.NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS;
    delete process.env.NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS;
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    container.remove();
    if (originalDemoAccessEnv === undefined) {
      delete process.env.NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS;
    } else {
      process.env.NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS = originalDemoAccessEnv;
    }
  });

  async function renderScreen() {
    await act(async () => {
      root.render(createElement(LoginScreen));
    });
  }

  test("renders the logout message from the query string", async () => {
    searchParamsState.message = "force_logout";

    await renderScreen();

    expect(container.textContent).toContain(
      "You have been logged out by an administrator.",
    );
  });

  test("does not render login settings controls", async () => {
    await renderScreen();

    expect(container.textContent).not.toContain("Auto Login (Quick Access)");
    expect(container.textContent).not.toContain("Quick Demo Login");
  });

  test("falls back to email login immediately when the last saved account is removed", async () => {
    window.localStorage.setItem(
      "dashpoint_saved_accounts",
      JSON.stringify([
        {
          id: "user-1",
          name: "Owner",
          email: "owner@dashpoint.local",
          role_name: "owner",
          has_pin: true,
          saved_at: new Date().toISOString(),
        },
      ]),
    );

    await renderScreen();

    const removeButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent?.includes("Remove Saved Account"),
    );
    expect(removeButton).toBeTruthy();

    await act(async () => {
      removeButton?.dispatchEvent(
        new MouseEvent("click", { bubbles: true, cancelable: true }),
      );
    });

    const emailInput = container.querySelector(
      'input[type="email"]',
    ) as HTMLInputElement | null;
    const quickAccessTrigger = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent?.includes("Quick Access"),
    );

    expect(emailInput).not.toBeNull();
    expect(quickAccessTrigger?.getAttribute("data-disabled")).toBe("");
  });

  test("uses the trusted-device preference in the effective save-account decision", async () => {
    window.localStorage.setItem("dashpoint_device_trusted", "true");
    process.env.NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS = "true";
    loginMock.mockResolvedValue({ success: true });

    await renderScreen();
    // Demo credentials load via a build-flag-gated dynamic import; wait for
    // the button they render instead of relying on import timing.
    const demoButton = await vi.waitFor(() => {
      const button = Array.from(container.querySelectorAll("button")).find(
        (button) => button.textContent?.includes("owner@dashpoint.local"),
      );
      if (!button) {
        throw new Error("demo access button not rendered yet");
      }
      return button;
    });
    const form = container.querySelector("form");

    await act(async () => {
      demoButton?.dispatchEvent(
        new MouseEvent("click", { bubbles: true, cancelable: true }),
      );
    });

    await act(async () => {
      form?.dispatchEvent(
        new Event("submit", { bubbles: true, cancelable: true }),
      );
    });

    expect(loginMock).toHaveBeenCalledWith(
      "owner@dashpoint.local",
      "owner123",
      true,
    );
    expect(routerPush).toHaveBeenCalledWith("/");
  });

  test("demo autofill updates credentials without submitting", async () => {
    process.env.NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS = "true";

    await renderScreen();
    // Demo credentials load via a build-flag-gated dynamic import; wait for
    // the button they render instead of relying on import timing.
    const demoButton = await vi.waitFor(() => {
      const button = Array.from(container.querySelectorAll("button")).find(
        (button) => button.textContent?.includes("owner@dashpoint.local"),
      );
      if (!button) {
        throw new Error("demo access button not rendered yet");
      }
      return button;
    });

    await act(async () => {
      demoButton?.dispatchEvent(
        new MouseEvent("click", { bubbles: true, cancelable: true }),
      );
    });

    const emailInput = container.querySelector(
      'input[type="email"]',
    ) as HTMLInputElement;
    const passwordInput = container.querySelector(
      'input[type="password"]',
    ) as HTMLInputElement;

    expect(emailInput.value).toBe("owner@dashpoint.local");
    expect(passwordInput.value).toBe("owner123");
    expect(loginMock).not.toHaveBeenCalled();
  });

  test("switches back to Quick Access when saved accounts appear on focus refresh", async () => {
    await renderScreen();

    expect(container.querySelector('input[type="email"]')).not.toBeNull();

    window.localStorage.setItem(
      "dashpoint_saved_accounts",
      JSON.stringify([
        {
          id: "user-1",
          name: "Owner",
          email: "owner@dashpoint.local",
          role_name: "owner",
          has_pin: true,
          saved_at: new Date().toISOString(),
        },
      ]),
    );

    await act(async () => {
      window.dispatchEvent(new Event("focus"));
    });

    const removeButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent?.includes("Remove Saved Account"),
    );

    expect(removeButton).toBeTruthy();
    expect(container.querySelector('input[type="email"]')).toBeNull();
  });
});
