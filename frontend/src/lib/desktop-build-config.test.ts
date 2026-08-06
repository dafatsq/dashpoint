import { describe, expect, test } from "vitest";

import {
  resolveDesktopBuildEnvironment,
  resolveDesktopClientSlug,
} from "../../scripts/desktop-build-config.mjs";

describe("resolveDesktopBuildEnvironment", () => {
  test("defaults desktop builds to the VPS API and disables demo access", () => {
    expect(resolveDesktopBuildEnvironment({})).toMatchObject({
      NEXT_PUBLIC_API_URL: "https://dashpoint.my.id/api/v1",
      NEXT_PUBLIC_CLIENT_SLUG: "dashpoint-demo",
      NEXT_PUBLIC_DESKTOP_BUILD: "true",
      NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS: "false",
    });
  });

  test("preserves explicit local API and demo-access overrides", () => {
    expect(
      resolveDesktopBuildEnvironment({
        NEXT_PUBLIC_API_URL: "http://localhost:8080/api/v1",
        NEXT_PUBLIC_CLIENT_SLUG: "acme-store",
        NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS: "true",
      }),
    ).toMatchObject({
      NEXT_PUBLIC_API_URL: "http://localhost:8080/api/v1",
      NEXT_PUBLIC_CLIENT_SLUG: "acme-store",
      NEXT_PUBLIC_DESKTOP_BUILD: "true",
      NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS: "true",
    });
  });

  test("rejects unsafe client slugs", () => {
    expect(() => resolveDesktopClientSlug("Acme Store")).toThrow(
      "Invalid desktop client slug",
    );
  });

  test("accepts lowercase hyphenated client slugs", () => {
    expect(resolveDesktopClientSlug("acme-store-01")).toBe("acme-store-01");
  });
});
