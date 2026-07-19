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

import { OfflineBanner } from "./offline-banner";

describe("OfflineBanner", () => {
  let container: HTMLDivElement;
  let root: Root;
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    container.remove();
    vi.unstubAllGlobals();
  });

  test("stays hidden when browser internet is offline but backend and database health are still reachable", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({
        status: "ok",
        services: {
          database: "connected",
        },
      }),
    });

    Object.defineProperty(window.navigator, "onLine", {
      configurable: true,
      value: true,
    });

    await act(async () => {
      root.render(createElement(OfflineBanner));
      await new Promise((resolve) => window.setTimeout(resolve, 0));
    });
    expect(container.textContent).toBe("");

    Object.defineProperty(window.navigator, "onLine", {
      configurable: true,
      value: false,
    });

    await act(async () => {
      window.dispatchEvent(new Event("offline"));
      await Promise.resolve();
    });
    expect(container.textContent).toBe("");
  });

  test("renders when backend health check fails and hides after backend connectivity returns", async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: "ok",
        services: {
          database: "connected",
        },
      }),
    });

    await act(async () => {
      root.render(createElement(OfflineBanner));
      await new Promise((resolve) => window.setTimeout(resolve, 0));
    });
    expect(container.textContent).toBe("");

    fetchMock.mockRejectedValueOnce(new Error("backend unavailable"));

    await act(async () => {
      window.dispatchEvent(new Event("offline"));
      await new Promise((resolve) => window.setTimeout(resolve, 0));
    });
    expect(container.textContent).toContain("Connection to backend or database lost");

    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: "ok",
        services: {
          database: "connected",
        },
      }),
    });

    Object.defineProperty(window.navigator, "onLine", {
      configurable: true,
      value: true,
    });

    await act(async () => {
      window.dispatchEvent(new Event("online"));
      await new Promise((resolve) => window.setTimeout(resolve, 0));
    });
    expect(container.textContent).toBe("");
  });
});
