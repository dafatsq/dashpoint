const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";
const DEFAULT_DESKTOP_API_BASE_URL = "https://dashpoint.my.id/api/v1";

export interface ApiEnvironment {
  apiUrl?: string;
  desktopBuild?: boolean;
}

export function resolveApiBaseUrl(environment: ApiEnvironment = {}): string {
  const explicitApiUrl = environment.apiUrl?.trim();
  if (explicitApiUrl) {
    return explicitApiUrl;
  }

  return environment.desktopBuild
    ? DEFAULT_DESKTOP_API_BASE_URL
    : DEFAULT_API_BASE_URL;
}

export const IS_DESKTOP_BUILD = process.env.NEXT_PUBLIC_DESKTOP_BUILD === "true";

export const API_BASE_URL = resolveApiBaseUrl({
  apiUrl: process.env.NEXT_PUBLIC_API_URL,
  desktopBuild: IS_DESKTOP_BUILD,
});

export const BACKEND_BASE_URL = API_BASE_URL.replace(/\/api\/v1\/?$/, "");

export function buildBackendUrl(path: string): string {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }

  return `${BACKEND_BASE_URL}${path}`;
}
