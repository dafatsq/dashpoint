// @vitest-environment node
import { describe, expect, test } from "vitest";

import { NextRequest } from "next/server";

import { proxy } from "./proxy";

function request(path: string, withRefreshCookie = false): NextRequest {
  const headers: Record<string, string> = {};
  if (withRefreshCookie) {
    headers.cookie = "refresh_token=placeholder";
  }
  return new NextRequest(`https://app.test${path}`, { headers });
}

describe("proxy route gate", () => {
  test("redirects logged-out requests on protected routes to /login", () => {
    const res = proxy(request("/pos"));
    expect(res.status).toBe(307);
    expect(res.headers.get("location")).toBe("https://app.test/login");
  });

  test("passes requests carrying a refresh cookie", () => {
    const res = proxy(request("/products", true));
    expect(res.status).toBe(200);
  });

  test("always allows /login", () => {
    const res = proxy(request("/login"));
    expect(res.status).toBe(200);
  });

  test("redirects nested dashboard routes too", () => {
    const res = proxy(request("/users/some-id/edit"));
    expect(res.headers.get("location")).toBe("https://app.test/login");
  });
});
