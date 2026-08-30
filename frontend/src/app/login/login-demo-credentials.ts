// Demo credentials are injected at BUILD time through
// NEXT_PUBLIC_DEMO_CREDENTIALS_JSON (a JSON array of {role, email, pass}).
// No credential literals exist in the repository source, so only builds that
// opt into demo mode — and set the variable at build time, e.g. the demo
// VPS client env — carry the values in their bundle. Every other build,
// including the default production compose build, contains none.
export interface DemoCredential {
  role: string;
  email: string;
  pass: string;
}

function parseDemoCredentials(raw: string | undefined): readonly DemoCredential[] {
  if (!raw || process.env.NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS !== "true") {
    return [];
  }
  try {
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return [];
    }
    return parsed.filter(
      (item): item is DemoCredential =>
        typeof item === "object" &&
        item !== null &&
        typeof (item as DemoCredential).role === "string" &&
        typeof (item as DemoCredential).email === "string" &&
        typeof (item as DemoCredential).pass === "string",
    );
  } catch {
    return [];
  }
}

export const DEMO_LOGIN_CREDENTIALS: readonly DemoCredential[] = parseDemoCredentials(
  process.env.NEXT_PUBLIC_DEMO_CREDENTIALS_JSON,
);
