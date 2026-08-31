import { describe, expect, test } from "vitest";

import { resolveDesktopBuildEnvironment } from "../../scripts/desktop-build-config.mjs";

describe("resolveDesktopBuildEnvironment", () => {
  test("defaults desktop builds to the VPS API", () => {
    expect(resolveDesktopBuildEnvironment({})).toMatchObject({
      NEXT_PUBLIC_API_URL: "https://dashpoint.my.id/api/v1",
      NEXT_PUBLIC_DESKTOP_BUILD: "true",
    });
    expect(resolveDesktopBuildEnvironment({})).not.toHaveProperty(
      "NEXT_PUBLIC_CLIENT_SLUG",
    );
  });

  test("preserves an explicit local API override", () => {
    expect(
      resolveDesktopBuildEnvironment({
        NEXT_PUBLIC_API_URL: "http://localhost:8080/api/v1",
      }),
    ).toMatchObject({
      NEXT_PUBLIC_API_URL: "http://localhost:8080/api/v1",
      NEXT_PUBLIC_DESKTOP_BUILD: "true",
    });
    expect(
      resolveDesktopBuildEnvironment({
        NEXT_PUBLIC_API_URL: "http://localhost:8080/api/v1",
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
