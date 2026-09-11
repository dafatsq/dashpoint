import { describe, expect, test } from "vitest";

import { resolveApiBaseUrl } from "./config";

describe("resolveApiBaseUrl", () => {
  test("uses the explicit API URL when provided", () => {
    expect(
      resolveApiBaseUrl({ apiUrl: "https://api.example.test/api/v1" }),
    ).toBe("https://api.example.test/api/v1");
  });

  test("keeps localhost as the normal web default", () => {
    expect(resolveApiBaseUrl({})).toBe("http://localhost:8080/api/v1");
  });
});
