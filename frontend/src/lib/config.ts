const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";

export interface ApiEnvironment {
  apiUrl?: string;
}

export function resolveApiBaseUrl(environment: ApiEnvironment = {}): string {
  const explicitApiUrl = environment.apiUrl?.trim();
  if (explicitApiUrl) {
    return explicitApiUrl;
  }

  return DEFAULT_API_BASE_URL;
}

export const API_BASE_URL = resolveApiBaseUrl({
  apiUrl: process.env.NEXT_PUBLIC_API_URL,
});

export const BACKEND_BASE_URL = API_BASE_URL.replace(/\/api\/v1\/?$/, "");

export function buildBackendUrl(path: string): string {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }

  return `${BACKEND_BASE_URL}${path}`;
}
