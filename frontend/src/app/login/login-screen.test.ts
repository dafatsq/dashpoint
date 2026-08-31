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
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    container.remove();
  });

  async function renderScreen() {
    await act(async () => {
      root.render(createElement(LoginScreen));
    });
  }

  function setInputValue(input: HTMLInputElement, value: string) {
    const setter = Object.getOwnPropertyDescriptor(
      window.HTMLInputElement.prototype,
      "value",
    )?.set;
    setter?.call(input, value);
    input.dispatchEvent(new Event("input", { bubbles: true }));
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
    loginMock.mockResolvedValue({ success: true });

    await renderScreen();

    const emailInput = container.querySelector(
      'input[type="email"]',
    ) as HTMLInputElement;
    const passwordInput = container.querySelector(
      'input[type="password"]',
    ) as HTMLInputElement;
    const form = container.querySelector("form");

    await act(async () => {
      setInputValue(emailInput, "owner@dashpoint.local");
      setInputValue(passwordInput, "owner123");
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
