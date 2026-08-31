# DashPoint desktop build

DashPoint's desktop client is a Wails v2 application that embeds the shared Next.js frontend from this repository. It uses the selected client's HTTPS API; it does not start a local backend, create a local database, or connect directly to PostgreSQL.

## One-time Windows setup

Install these tools on the Windows build machine:

- Git
- Go 1.25 or newer
- Node.js 22 or newer, including npm
- Microsoft WebView2 Evergreen Runtime

The build script installs the pinned Wails CLI v2.13.0 when the required version is not available.

## Build a client executable privately

Each client's executable is built from that client's deployment branch (`clients/<client-slug>`, cut from `main`), so any client-specific frontend code is included. Run from the repository root. The default target is the DashPoint demo API:

```powershell
git checkout clients/<client-slug>
git pull
.\desktop\scripts\build-windows.ps1
```

To build the executable for a client's VPS, pass that client's public HTTPS API URL:

```powershell
.\desktop\scripts\build-windows.ps1 `
  -ApiBaseUrl "https://client.example.com/api/v1"
```

The output is always:

```text
build\bin\DashPoint.exe
```

Record the branch and commit the executable was built from alongside the build date — the desktop build is not tracked anywhere automatic. The `build` directory and executable are ignored by Git. Transfer the executable to the client through the private distribution channel chosen for that client; do not commit or publish it in a GitHub release.

For the full client provisioning flow—including the VPS environment file, DNS, Caddy, CI/CD target, and desktop API URL—see [CLIENTS/DEPLOY_NEW_CLIENT.md](../CLIENTS/DEPLOY_NEW_CLIENT.md).

## Local API and demo overrides

Local API access requires an explicit development flag:

```powershell
.\desktop\scripts\build-windows.ps1 `
  -ApiBaseUrl "http://localhost:8080/api/v1" `
  -AllowLocalApi
```

Demo access is not part of the core product. It exists only on the `dashpoint-demo` deployment branch; build demo executables from that branch. Production API URLs must use HTTPS. The executable contains only the public API URL and shared frontend assets. It never contains database credentials, JWT secrets, SSH keys, VPS secrets, or production `.env` files.

## Mac development

The existing frontend can still run normally:

```bash
cd frontend
npm ci
npm run dev
```

To run the Wails development shell from the repository root:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
wails dev
```

Wails uses the Next.js development server at `http://localhost:3000`. Set `NEXT_PUBLIC_API_URL` when the local shell should use a different backend.

## Production API and authentication

The production backend must allow these exact origins in `CORS_ORIGINS`:

```text
wails://wails
http://wails.localhost
```

The website origin must remain included as well. The backend adapts the refresh cookie for Wails origins while retaining strict cookies for normal website requests. HTTPS is required for secure refresh-cookie behavior.

GitHub Actions validates the desktop frontend assets and root Wails module, but it does not build, upload, or release a Windows executable.
