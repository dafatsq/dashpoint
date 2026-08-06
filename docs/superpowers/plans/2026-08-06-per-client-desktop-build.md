# Per-client DashPoint desktop builds Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the Wails desktop build reusable for multiple isolated DashPoint client VPSs by applying an explicit client API URL and safe client slug at build time.

**Architecture:** Keep `/frontend` as the canonical UI and continue embedding it into Wails. The build script validates a client slug and API URL, passes public build configuration into Next.js, and produces a client-named executable. GitHub Actions keeps automatic demo/template artifacts while manual client builds and releases accept explicit client inputs.

**Tech Stack:** Wails v2.13.0, Go 1.25+, PowerShell, Next.js 16, Node.js 22+, Vitest, GitHub Actions.

## Global Constraints

- The executable will never contain database credentials, JWT secrets, SSH keys, VPS secrets, or production `.env` files.
- Production API URLs must use HTTPS; localhost URLs require `-AllowLocalApi`.
- Client slugs use lowercase letters, numbers, and single hyphens only: `^[a-z0-9]+(?:-[a-z0-9]+)*$`.
- The shared frontend remains the only UI source for website and desktop builds.
- The existing `v0.1.0` demo release remains unchanged.
- No new npm or Go dependencies are required.

---

### Task 1: Add tested desktop client configuration

**Files:**
- Modify: `frontend/scripts/desktop-build-config.mjs`
- Test: `frontend/src/lib/desktop-build-config.test.ts`

**Interfaces:**
- `resolveDesktopClientSlug(value?: string): string` returns `dashpoint-demo` for an empty value and validates explicit slugs.
- `resolveDesktopBuildEnvironment(environment?: Record<string, string>): Record<string, string>` sets `NEXT_PUBLIC_CLIENT_SLUG` in addition to the existing desktop environment values.

- [ ] **Step 1: Write the failing tests**

Add tests for the default slug, explicit slug, and invalid slug:

```ts
test("defaults the desktop client slug to dashpoint-demo", () => {
  expect(resolveDesktopBuildEnvironment({}).NEXT_PUBLIC_CLIENT_SLUG).toBe(
    "dashpoint-demo",
  );
});

test("preserves a valid explicit client slug", () => {
  expect(
    resolveDesktopBuildEnvironment({ NEXT_PUBLIC_CLIENT_SLUG: "acme-store" })
      .NEXT_PUBLIC_CLIENT_SLUG,
  ).toBe("acme-store");
});

test("rejects unsafe client slugs", () => {
  expect(() =>
    resolveDesktopBuildEnvironment({ NEXT_PUBLIC_CLIENT_SLUG: "Acme Store" }),
  ).toThrow("client slug");
});
```

- [ ] **Step 2: Run the focused test and verify it fails**

Run:

```bash
cd frontend
npx vitest run src/lib/desktop-build-config.test.ts
```

Expected: the new slug assertions fail because the current helper neither validates nor returns `NEXT_PUBLIC_CLIENT_SLUG`.

- [ ] **Step 3: Implement the minimal configuration helper**

Add the default slug, validation regex, and `resolveDesktopClientSlug`. Use the resolved slug in `resolveDesktopBuildEnvironment` while preserving the current API URL, desktop marker, and demo-access behavior.

- [ ] **Step 4: Run the focused test and verify it passes**

Run the same Vitest command. Expected: all desktop build configuration tests pass.

- [ ] **Step 5: Commit the tested configuration change**

```bash
git add frontend/scripts/desktop-build-config.mjs frontend/src/lib/desktop-build-config.test.ts
git commit -m "feat: add desktop client build metadata"
```

### Task 2: Make the Windows build client-specific and safe

**Files:**
- Modify: `desktop/scripts/build-windows.ps1`
- Modify: `desktop/README.md`

**Interfaces:**
- Add `-ClientSlug "client-example"` with default `dashpoint-demo`.
- Preserve `-ApiBaseUrl`, `-EnableQuickDemoAccess`, and `-AllowLocalApi`.
- Produce both the normal Wails output `build\bin\DashPoint.exe` and a client-named copy `build\bin\DashPoint-<client-slug>.exe`.

- [ ] **Step 1: Add the client slug validation test fixture**

Use the focused Node configuration test from Task 1 as the executable build contract, and add a PowerShell smoke-check command to the task verification instructions:

```powershell
.\desktop\scripts\build-windows.ps1 -ClientSlug "acme-store" -ApiBaseUrl "https://acme.example.com/api/v1"
```

- [ ] **Step 2: Implement the minimal PowerShell changes**

Add the `ClientSlug` parameter and validate it before tool installation or Wails compilation. Set `NEXT_PUBLIC_CLIENT_SLUG` before invoking Wails. If `ClientSlug` is not `dashpoint-demo`, reject an empty API URL instead of silently targeting the demo VPS. Keep the demo fallback for backwards-compatible demo builds. Copy and scan the executable to the client-named output path after Wails succeeds.

- [ ] **Step 3: Update desktop build documentation**

Document the client build command, output filename, safe URL rules, and the distinction between the demo default and explicit client builds. Explain that the slug and API URL are public configuration only.

- [ ] **Step 4: Verify the Windows build contract**

On Windows, run the client-specific command above and confirm both files exist:

```powershell
Test-Path .\build\bin\DashPoint.exe
Test-Path .\build\bin\DashPoint-acme-store.exe
```

Also verify that an invalid slug and a non-demo build without `-ApiBaseUrl` fail before producing a client artifact.

- [ ] **Step 5: Commit the build-script and documentation changes**

```bash
git add desktop/scripts/build-windows.ps1 desktop/README.md
git commit -m "feat: support client-specific desktop builds"
```

### Task 3: Add deliberate per-client GitHub Actions inputs and artifacts

**Files:**
- Modify: `.github/workflows/desktop.yml`

**Interfaces:**
- Add workflow inputs `client_slug` and `api_base_url`.
- Use `dashpoint-demo` as the automatic-build default.
- Upload `DashPoint-<client-slug>.exe` under an artifact named `DashPoint-<client-slug>-windows`.
- Release assets use the same client-specific filename.

- [ ] **Step 1: Add workflow input and naming changes**

Add a `client_slug` string input with default `dashpoint-demo`. Keep `api_base_url` optional only for the demo default; non-demo builds must provide it. Pass both values to `build-windows.ps1`, upload the client-named executable, and download/upload the same artifact in the release job.

- [ ] **Step 2: Add release validation**

Reject a manual non-demo release when `api_base_url` is empty. Document the tag convention in the workflow input description: `v1.0.0` for the demo and `<client-slug>-v1.0.0` for client-specific releases.

- [ ] **Step 3: Validate workflow syntax and expressions locally**

Run:

```bash
ruby -e 'require "yaml"; ARGV.each { |path| YAML.load_file(path); puts "parsed #{path}" }' .github/workflows/ci.yml .github/workflows/desktop.yml .github/workflows/deploy.yml
git diff --check
```

- [ ] **Step 4: Commit the workflow changes**

```bash
git add .github/workflows/desktop.yml
git commit -m "ci: add per-client desktop releases"
```

### Task 4: Run the complete verification suite

**Files:**
- Verify: `frontend/src/lib/desktop-build-config.test.ts`
- Verify: `.github/workflows/desktop.yml`
- Verify: `.github/workflows/deploy.yml`

- [ ] **Step 1: Run frontend tests and lint**

```bash
cd frontend
npm test -- --run
npm run lint -- --quiet
```

- [ ] **Step 2: Run frontend builds and Go tests**

```bash
npm run build
npm run build:desktop
cd ..
go test -tags webkit2_41 ./...
cd backend
go test ./...
```

- [ ] **Step 3: Run deployment and secret checks**

```bash
cd ..
docker compose --env-file ci/fixtures/compose.env -f docker-compose.prod.yml config --quiet
if rg -n --hidden -g '!.git/**' -g '!frontend/node_modules/**' -g '!frontend/.next/**' -g '!frontend/dist/**' -g '!build/**' -g '!ci/fixtures/compose.env' '(-----BEGIN (RSA|OPENSSH|EC|DSA) PRIVATE KEY-----|ghp_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|AKIA[0-9A-Z]{16})' .; then exit 1; fi
```

- [ ] **Step 4: Verify GitHub behavior**

Confirm that a shared frontend commit produces the demo artifact and deploys the VPS, while a manual client dispatch produces a client-named artifact with the selected API URL and does not change the VPS unless the commit also contains deployable source changes.

- [ ] **Step 5: Commit and push the final verified state**

```bash
git status --short --branch
git push origin main
```
