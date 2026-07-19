const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? DEFAULT_API_BASE_URL;

export const BACKEND_BASE_URL = API_BASE_URL.replace(/\/api\/v1\/?$/, "");

export function buildBackendUrl(path: string): string {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }

  return `${BACKEND_BASE_URL}${path}`;
}
