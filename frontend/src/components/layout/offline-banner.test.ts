// @vitest-environment jsdom

import { act, createElement } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, test } from "vitest";

import { OfflineBanner } from "./offline-banner";

describe("OfflineBanner", () => {
  let container: HTMLDivElement;
  let root: Root;

  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    container.remove();
  });

  test("renders when the browser goes offline and hides when it comes back online", () => {
    Object.defineProperty(window.navigator, "onLine", {
      configurable: true,
      value: true,
    });

    act(() => {
      root.render(createElement(OfflineBanner));
    });
    expect(container.textContent).toBe("");

    Object.defineProperty(window.navigator, "onLine", {
      configurable: true,
      value: false,
    });
    act(() => {
      window.dispatchEvent(new Event("offline"));
    });
    expect(container.textContent).toContain("No internet connection");

    Object.defineProperty(window.navigator, "onLine", {
      configurable: true,
      value: true,
    });
    act(() => {
      window.dispatchEvent(new Event("online"));
    });
    expect(container.textContent).toBe("");
  });
});
