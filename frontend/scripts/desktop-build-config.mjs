const DEFAULT_DESKTOP_API_URL = "https://dashpoint.my.id/api/v1";

export function resolveDesktopBuildEnvironment(environment = {}) {
  const explicitApiUrl = environment.NEXT_PUBLIC_API_URL?.trim();

  return {
    ...environment,
    NEXT_PUBLIC_API_URL: explicitApiUrl || DEFAULT_DESKTOP_API_URL,
    NEXT_PUBLIC_DESKTOP_BUILD: "true",
  };
}
