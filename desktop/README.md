# DashPoint desktop build

DashPoint's desktop client is a Wails v2 application that embeds the existing Next.js frontend. It uses the same HTTPS backend and PostgreSQL database as the VPS deployment; it does not start a local backend, create a local database, or require Node.js after the executable is built.

## One-time Windows setup

Install these tools on the Windows laptop:

- Git
- Go 1.25 or newer
- Node.js 22 or newer, including npm
- Microsoft WebView2 Evergreen Runtime

The build script installs the pinned Wails CLI v2.13.0 when the required version is not available.

## Build the Windows executable

Run from the repository root:

```powershell
git pull
.\desktop\scripts\build-windows.ps1
```

The default endpoint is:

```text
https://dashpoint.my.id/api/v1
```

The result is:

```text
build\bin\DashPoint.exe
```

## API endpoint and demo overrides

Use an explicit HTTPS endpoint for staging or another client:

```powershell
.\desktop\scripts\build-windows.ps1 -ApiBaseUrl "https://client.example.com/api/v1"
```

Local API access requires an explicit development flag:

```powershell
.\desktop\scripts\build-windows.ps1 `
  -ApiBaseUrl "http://localhost:8080/api/v1" `
  -AllowLocalApi
```

Demo access is disabled by default. Enable it only for a deliberate demo build:

```powershell
.\desktop\scripts\build-windows.ps1 -EnableQuickDemoAccess:$true
```

The executable never connects directly to PostgreSQL. It only calls the configured HTTPS API.

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

Generated frontend output, Wails binaries, Go caches, Node dependencies, and local environment files are ignored by Git. Only source files and dependency manifests should be pushed.
