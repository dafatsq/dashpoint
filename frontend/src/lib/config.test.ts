import { describe, expect, test } from "vitest";

import { resolveApiBaseUrl } from "./config";

describe("resolveApiBaseUrl", () => {
  test("uses the explicit API URL for a desktop build", () => {
    expect(
      resolveApiBaseUrl({
        apiUrl: "https://api.example.test/api/v1",
        desktopBuild: true,
      }),
    ).toBe("https://api.example.test/api/v1");
  });

  test("uses the production endpoint for a desktop build without an override", () => {
    expect(resolveApiBaseUrl({ desktopBuild: true })).toBe(
      "https://dashpoint.my.id/api/v1",
    );
  });

  test("keeps localhost as the normal web default", () => {
    expect(resolveApiBaseUrl({ desktopBuild: false })).toBe(
      "http://localhost:8080/api/v1",
    );
  });
});
