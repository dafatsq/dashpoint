const DEFAULT_DESKTOP_API_URL = "https://dashpoint.my.id/api/v1";
const DEFAULT_DESKTOP_CLIENT_SLUG = "dashpoint-demo";
const DESKTOP_CLIENT_SLUG_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export function resolveDesktopClientSlug(value) {
  const clientSlug = value?.trim() || DEFAULT_DESKTOP_CLIENT_SLUG;

  if (!DESKTOP_CLIENT_SLUG_PATTERN.test(clientSlug)) {
    throw new Error(
      `Invalid desktop client slug: ${clientSlug}. Use lowercase letters, numbers, and single hyphens.`,
    );
  }

  return clientSlug;
}

export function resolveDesktopBuildEnvironment(environment = {}) {
  const explicitApiUrl = environment.NEXT_PUBLIC_API_URL?.trim();
  const clientSlug = resolveDesktopClientSlug(
    environment.NEXT_PUBLIC_CLIENT_SLUG,
  );
  const demoAccessEnabled =
    environment.NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS === "true";

  return {
    ...environment,
    NEXT_PUBLIC_API_URL: explicitApiUrl || DEFAULT_DESKTOP_API_URL,
    NEXT_PUBLIC_CLIENT_SLUG: clientSlug,
    NEXT_PUBLIC_DESKTOP_BUILD: "true",
    NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS: demoAccessEnabled ? "true" : "false",
  };
}
