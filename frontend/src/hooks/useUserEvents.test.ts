// @vitest-environment jsdom

import { act, createElement } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";

import { useUserEvents } from "./useUserEvents";

const { getAccessTokenMock } = vi.hoisted(() => ({
  getAccessTokenMock: vi.fn(),
}));

vi.mock("@/lib/auth-session", () => ({
  getAccessToken: getAccessTokenMock,
  refreshSessionTokens: vi.fn(),
}));

function EventStreamProbe() {
  useUserEvents({ enabled: true });
  return null;
}

describe("useUserEvents", () => {
  let container: HTMLDivElement;
  let root: Root;

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
    getAccessTokenMock.mockReturnValue("access-token");
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    container.remove();
    vi.restoreAllMocks();
    getAccessTokenMock.mockReset();
  });

  test("includes cookies when opening the SSE stream", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(new Response(null, { status: 200 }));

    await act(async () => {
      root.render(createElement(EventStreamProbe));
      await Promise.resolve();
    });

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/events/subscribe"),
      expect.objectContaining({
        credentials: "include",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
          Accept: "text/event-stream",
        }),
      }),
    );
  });
});
