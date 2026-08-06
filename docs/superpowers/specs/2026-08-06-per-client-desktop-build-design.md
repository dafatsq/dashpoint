# Per-client DashPoint desktop builds

## Goal

Treat the repository as the canonical DashPoint POS SaaS template while producing desktop artifacts that connect to the intended client's VPS. The shared frontend remains the only UI source; client-specific public configuration is applied when the Wails executable is built.

The existing `v0.1.0` demo release is not changed by this work.

## Scope

Each client desktop build will contain:

- The canonical frontend bundled into the Wails executable.
- The client's public API base URL, such as `https://client.example.com/api/v1`.
- A client slug used to name CI artifacts and release assets unambiguously.
- Demo access disabled unless deliberately enabled for a demo build.

The executable will never contain database credentials, JWT secrets, SSH keys, VPS secrets, or production `.env` files. Client branding and automatic in-app update delivery remain separate follow-up work.

## Build contract

The Windows build script will accept:

```powershell
-ApiBaseUrl "https://client.example.com/api/v1"
-ClientSlug "client-example"
-EnableQuickDemoAccess:$false
```

The API URL is validated as HTTPS for production builds. Localhost URLs remain available only with `-AllowLocalApi`. The client slug is restricted to a safe filename form and is used only for artifact naming and build metadata.

The build continues to set `NEXT_PUBLIC_API_URL`, `NEXT_PUBLIC_DESKTOP_BUILD`, and `NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS` before Wails embeds the frontend. The generated executable is scanned for private configuration before it is accepted.

## GitHub Actions distribution

The desktop workflow will retain three paths:

1. Successful `CI` runs can create the normal demo/template artifact for shared frontend or desktop changes.
2. Manual dispatch can build a selected client by accepting `client_slug` and `api_base_url` inputs.
3. A tag or explicit manual release publishes the selected artifact with a client-specific filename.

Per-client releases are created deliberately; a frontend commit does not silently publish a production executable for every client. A client release therefore records which public API target was selected without exposing private credentials.

## Runtime and deployment boundaries

The executable communicates only with the configured API. It does not connect directly to PostgreSQL and does not alter the client's VPS, Docker, Caddy, or database configuration. Website, backend, migration, and deployment changes continue through the existing VPS workflow.

## Verification

Verification will cover:

- Valid client API URLs are embedded in the desktop frontend build environment.
- Local API URLs are rejected unless `-AllowLocalApi` is supplied.
- Unsafe client slugs are rejected before building.
- Production/demo access defaults remain safe.
- Client artifact names are distinct and deterministic.
- Generated assets and executables contain no forbidden private configuration.
- Existing frontend, backend, Wails, Compose, Caddy, and secret checks remain passing.
- A manual client build can be traced to its selected API URL and client slug.
