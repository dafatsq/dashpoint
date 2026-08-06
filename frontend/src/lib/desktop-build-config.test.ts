import { describe, expect, test } from "vitest";

import { resolveDesktopBuildEnvironment } from "../../scripts/desktop-build-config.mjs";

describe("resolveDesktopBuildEnvironment", () => {
  test("defaults desktop builds to the VPS API and disables demo access", () => {
    expect(resolveDesktopBuildEnvironment({})).toMatchObject({
      NEXT_PUBLIC_API_URL: "https://dashpoint.my.id/api/v1",
      NEXT_PUBLIC_DESKTOP_BUILD: "true",
      NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS: "false",
    });
    expect(resolveDesktopBuildEnvironment({})).not.toHaveProperty(
      "NEXT_PUBLIC_CLIENT_SLUG",
    );
  });

  test("preserves explicit local API and demo-access overrides", () => {
    expect(
      resolveDesktopBuildEnvironment({
        NEXT_PUBLIC_API_URL: "http://localhost:8080/api/v1",
        NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS: "true",
      }),
    ).toMatchObject({
      NEXT_PUBLIC_API_URL: "http://localhost:8080/api/v1",
      NEXT_PUBLIC_DESKTOP_BUILD: "true",
      NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS: "true",
    });
    expect(
      resolveDesktopBuildEnvironment({
        NEXT_PUBLIC_API_URL: "http://localhost:8080/api/v1",
        NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS: "true",
      }),
    ).not.toHaveProperty("NEXT_PUBLIC_CLIENT_SLUG");
  });

  test("preserves unrelated environment values", () => {
    expect(
      resolveDesktopBuildEnvironment({ CUSTOM_BUILD_FLAG: "enabled" }),
    ).toMatchObject({
      CUSTOM_BUILD_FLAG: "enabled",
    });
  });
});
