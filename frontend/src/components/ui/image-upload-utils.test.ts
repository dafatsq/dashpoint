import { afterEach, describe, expect, test, vi } from "vitest";

import {
  MAX_IMAGE_UPLOAD_SIZE,
  uploadImageBlob,
  validateImageFile,
} from "./image-upload-utils";

describe("image upload utils", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("rejects non-image files", () => {
    const file = new File(["x"], "notes.txt", { type: "text/plain" });
    expect(validateImageFile(file)).toBe("Please select an image file");
  });

  test("rejects oversized image files", () => {
    const file = new File(["x".repeat(10)], "large.png", { type: "image/png" });
    Object.defineProperty(file, "size", { value: MAX_IMAGE_UPLOAD_SIZE + 1 });
    expect(validateImageFile(file)).toBe("Image must be less than 5MB");
  });

  test("builds the upload request and returns the backend url", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ url: "/uploads/example.jpg" }),
    });

    const blob = new Blob(["image"], { type: "image/jpeg" });
    const url = await uploadImageBlob({
      blob,
      fetchImpl: fetchMock,
      apiBaseUrl: "http://localhost:8080/api/v1",
      accessToken: "token-123",
    });

    expect(url).toBe("/uploads/example.jpg");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [endpoint, request] = fetchMock.mock.calls[0];
    expect(endpoint).toBe("http://localhost:8080/api/v1/upload/image");
    expect(request.headers).toEqual({
      Authorization: "Bearer token-123",
    });
  });
});
