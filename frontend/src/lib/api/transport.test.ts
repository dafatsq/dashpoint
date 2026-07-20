import { beforeEach, describe, expect, test, vi } from "vitest";

import { ApiTransport } from "./transport";

const {
  clearAuthSessionMock,
  getAccessTokenMock,
  persistAuthPayloadMock,
  refreshSessionTokensMock,
} = vi.hoisted(() => ({
  clearAuthSessionMock: vi.fn(),
  getAccessTokenMock: vi.fn(),
  persistAuthPayloadMock: vi.fn(),
  refreshSessionTokensMock: vi.fn(),
}));

vi.mock("@/lib/auth-session", () => ({
  clearAuthSession: clearAuthSessionMock,
  getAccessToken: getAccessTokenMock,
  persistAuthPayload: persistAuthPayloadMock,
  refreshSessionTokens: refreshSessionTokensMock,
}));

describe("ApiTransport", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    getAccessTokenMock.mockReset();
    refreshSessionTokensMock.mockReset();
    clearAuthSessionMock.mockReset();
    persistAuthPayloadMock.mockReset();
  });

  test("retries protected requests without sending cookies", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ code: "INVALID_TOKEN" }), {
          status: 401,
          headers: { "Content-Type": "application/json" },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ ok: true }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      );

    getAccessTokenMock
      .mockReturnValueOnce("expired-access")
      .mockReturnValueOnce("fresh-access");
    refreshSessionTokensMock.mockResolvedValue(true);

    const transport = new ApiTransport("http://localhost:8080/api/v1");
    const result = await transport.request<{ ok: boolean }>("/me");

    expect(result.data).toEqual({ ok: true });
    expect(refreshSessionTokensMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://localhost:8080/api/v1/me",
      expect.objectContaining({
        credentials: "same-origin",
        headers: expect.objectContaining({
          Authorization: "Bearer expired-access",
        }),
      }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://localhost:8080/api/v1/me",
      expect.objectContaining({
        credentials: "same-origin",
        headers: expect.objectContaining({
          Authorization: "Bearer fresh-access",
        }),
      }),
    );
  });

  test("refreshes a cookie-backed session when the access token is missing", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ code: "MISSING_TOKEN" }), {
          status: 401,
          headers: { "Content-Type": "application/json" },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ ok: true }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      );

    getAccessTokenMock
      .mockReturnValueOnce(null)
      .mockReturnValueOnce("fresh-access");
    refreshSessionTokensMock.mockResolvedValue(true);

    const transport = new ApiTransport("http://localhost:8080/api/v1");
    const result = await transport.request<{ ok: boolean }>("/me");

    expect(result.data).toEqual({ ok: true });
    expect(refreshSessionTokensMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://localhost:8080/api/v1/me",
      expect.objectContaining({
        headers: expect.not.objectContaining({
          Authorization: expect.any(String),
        }),
      }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://localhost:8080/api/v1/me",
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: "Bearer fresh-access",
        }),
      }),
    );
  });

  test("supports cookie-backed auth requests without Authorization", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ access_token: "access" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    getAccessTokenMock.mockReturnValue("stale-access");

    const transport = new ApiTransport("http://localhost:8080/api/v1");
    await transport.request("/auth/pin-login", {
      method: "POST",
      body: { user_id: "user-1", pin: "1234" },
      credentials: "include",
      skipAuth: true,
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/auth/pin-login",
      expect.objectContaining({
        credentials: "include",
        headers: expect.not.objectContaining({
          Authorization: expect.any(String),
        }),
      }),
    );
  });

  test("retries blob downloads with cookie credentials", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ code: "INVALID_TOKEN" }), {
          status: 401,
          headers: { "Content-Type": "application/json" },
        }),
      )
      .mockResolvedValueOnce(
        new Response("csv,data\n", {
          status: 200,
          headers: { "Content-Type": "text/csv" },
        }),
      );

    getAccessTokenMock
      .mockReturnValueOnce("expired-access")
      .mockReturnValueOnce("fresh-access");
    refreshSessionTokensMock.mockResolvedValue(true);

    const transport = new ApiTransport("http://localhost:8080/api/v1");
    const result = await transport.requestBlob("/reports/export/sales");

    expect(result.data).toBeDefined();
    expect(refreshSessionTokensMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://localhost:8080/api/v1/reports/export/sales",
      expect.objectContaining({
        credentials: "include",
        headers: expect.objectContaining({
          Authorization: "Bearer expired-access",
        }),
      }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://localhost:8080/api/v1/reports/export/sales",
      expect.objectContaining({
        credentials: "include",
        headers: expect.objectContaining({
          Authorization: "Bearer fresh-access",
        }),
      }),
    );
  });

  test("does not log full error response payloads", async () => {
    const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          code: "FORBIDDEN",
          message: "Denied",
          request_id: "req-1",
          secret: "do-not-log",
        }),
        {
          status: 403,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );
    getAccessTokenMock.mockReturnValue(null);

    const transport = new ApiTransport("http://localhost:8080/api/v1");
    await transport.request("/restricted");

    expect(warnSpy).toHaveBeenCalledWith(
      "API Client Error:",
      expect.objectContaining({
        status: 403,
        message: "Denied",
        code: "FORBIDDEN",
        request_id: "req-1",
      }),
    );
    expect(warnSpy.mock.calls[0]?.[1]).not.toHaveProperty("data");
  });
});
