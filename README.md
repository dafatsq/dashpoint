# DashPoint POS

DashPoint is a production-grade, multi-tenant-ready **Point of Sale (POS) platform** built as a reusable SaaS template. It pairs a hardened **Go (Fiber) backend** with a modern **Next.js 16 / React 19 frontend**, backed by **PostgreSQL**, and ships with a complete multi-client deployment pipeline (Docker Compose + Caddy + GitHub Actions).

The application targets small retail stores (per the frontend metadata: "Point of Sale System for Indonesian Stores") and is designed so that **one VPS can host many isolated client instances**, each with its own domain, database, credentials, and deployment branch.

---

## Table of Contents

- [Key Features](#key-features)
- [Architecture Overview](#architecture-overview)
- [Tech Stack](#tech-stack)
- [Repository Layout](#repository-layout)
- [Prerequisites](#prerequisites)
- [Getting Started (Local Development)](#getting-started-local-development)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Set Up Environment Variables](#2-set-up-environment-variables)
  - [3. Start the Database and Backend API (Docker)](#3-start-the-database-and-backend-api-docker)
  - [4. Alternative: Run the Backend Natively](#4-alternative-run-the-backend-natively)
  - [5. Start the Frontend (Next.js)](#5-start-the-frontend-nextjs)
  - [6. First Run: Owner Setup](#6-first-run-owner-setup)
  - [7. Access the Application](#7-access-the-application)
- [Demo Data Seeding](#demo-data-seeding)
- [Running Tests](#running-tests)
- [Backend API Overview](#backend-api-overview)
  - [Conventions](#conventions)
  - [Public Endpoints](#public-endpoints)
  - [Protected Endpoints](#protected-endpoints)
  - [Roles & Permissions](#roles--permissions)
- [Authentication & Security Model](#authentication--security-model)
- [Database & Migrations](#database--migrations)
- [Realtime Events (SSE)](#realtime-events-sse)
- [Configuration Reference](#configuration-reference)
- [Deployment](#deployment)
  - [Single-Client Deployment](#single-client-deployment)
  - [Multi-Client VPS (SaaS Template)](#multi-client-vps-saas-template)
  - [CI/CD Pipeline](#cicd-pipeline)
  - [Backups & Data Safety](#backups--data-safety)
- [Operational Scripts](#operational-scripts)
- [Troubleshooting](#troubleshooting)
- [Additional Documentation](#additional-documentation)

---

## Key Features

**Point of Sale**
- Fast POS screen with product grid, barcode/SKU lookup, cart management, item & cart-level discounts, and per-item tax.
- Checkout with multiple payment methods: cash (with change calculation), card, QRIS, transfer, voucher, credit, other.
- Cart validation endpoint for pre-checkout stock and price verification.
- Sale voiding with mandatory reason, audit-tracked.

**Shift & Cash Drawer Management**
- Open/close register shifts with opening/expected/closing cash and cash-difference tracking.
- Pay-in / pay-out cash drawer operations with reasons, tied to the active shift.
- Single-open-shift enforcement at the database level (no concurrent open shifts).

**Inventory**
- Stock quantities with per-product low-stock thresholds and dashboard low-stock alerts.
- Stock adjustments (purchase, damage, loss, adjustment, count) with before/after history, permission-gated per adjustment type.
- Optimistic concurrency control (`expected_updated_at`) to prevent lost updates from concurrent edits.

**Catalog**
- Products with SKU, barcode, description, cost/price, tax rate, image upload, and categories.
- Soft deletes plus permanent deletes for products, categories, users, and expense categories.

**Expenses**
- Expense tracking with categories, vendors, reference numbers, and monthly summaries.
- Special `inventory_purchase` category that can restock inventory directly from an expense entry.

**Reports & Analytics**
- Daily sales, sales range, top sellers, inventory valuation, cash report, per-employee sales, per-category sales.
- CSV exports (sales, inventory, top sellers, comprehensive) gated behind `manage_reports_page`.

**Dashboard & Audit**
- Live dashboard with stats, low-stock list, and a "Recent Changes" activity feed.
- Comprehensive audit logging of every mutation (who, what, old/new values, IP, user agent, request ID).
- Admin audit-log browser with filters, entity history, and per-user activity views.

**Users, Roles & Security**
- Three seeded roles: **owner** (all permissions), **manager**, **cashier** (curated permission sets).
- Password + PIN login paths; bcrypt (cost 12) hashing for both.
- Short-lived in-memory access tokens + httpOnly refresh-token cookie with rotation, token-family reuse detection, and grace-window handling for sibling tabs.
- Account lockout (5 failures → 15 min), per-IP rate limits, strict JSON body parsing, magic-byte image validation, trusted-proxy allowlisting, and security headers.

**Realtime**
- Server-Sent Events push notifications so role changes, deactivations, deletions, and forced logouts take effect immediately across sessions.

---

## Architecture Overview

```
┌────────────────────────────────────────────────────────────────────┐
│                          Internet (HTTPS)                          │
└──────────────────────────────┬─────────────────────────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   Caddy (80/443)    │  one global reverse proxy
                    │  auto-TLS per host  │  per VPS
                    └──────────┬──────────┘
              ┌────────────────┴────────────────┐
              │                                 │
   ┌──────────▼──────────┐          ┌──────────▼──────────┐
   │  Frontend container │  proxy   │  Backend container  │
   │  Next.js 16 :3000   │─────────►│  Go Fiber  :8080    │
   └─────────────────────┘          └──────────┬──────────┘
                                               │ (TLS in prod)
                                    ┌──────────▼──────────┐
                                    │ PostgreSQL 15 :5432 │
                                    │ per-client, private │
                                    └─────────────────────┘
```

- **Frontend → Backend**: the browser talks to the Go API directly (`NEXT_PUBLIC_API_URL`, e.g. `/api/v1` behind Caddy). Caddy routes `/api/v1/*` and `/uploads/*` to the backend, everything else to the frontend.
- **Backend layers**: `cmd/server` (wiring & routes) → `internal/handlers` (HTTP) → `internal/repository` (SQL via pgx) → PostgreSQL. Business logic lives in handlers/repositories, never in route registration.
- **Isolation (multi-client)**: each client gets its own Compose project name, env file, database, uploads directory, hostname, and a long-lived `clients/<slug>` Git deployment branch.

---

## Tech Stack

| Layer     | Technology |
|-----------|------------|
| Backend   | Go 1.25, [Fiber v2](https://gofiber.io) (fasthttp), pgx v5 (PostgreSQL), golang-migrate, golang-jwt v5, bcrypt (golang.org/x/crypto), shopspring/decimal, zerolog, godotenv |
| Frontend  | Next.js 16 (App Router, Turbopack), React 19, TypeScript 5, Tailwind CSS 4, Radix UI primitives, lucide-react, react-day-picker, react-easy-crop, date-fns |
| Testing   | Go `testing` (+ `govulncheck`), Vitest 3 (jsdom) — 45 backend test files, 34 frontend test files |
| Database  | PostgreSQL 15 (pinned image digest), 61 numbered schema migrations |
| Infra     | Docker Compose (dev, prod, caddy), Caddy 2 reverse proxy, GitHub Actions CI/CD |

---

## Repository Layout

```
.
├── backend/                     # Go API service
│   ├── cmd/server/              # entrypoint, app wiring, route registration
│   ├── internal/
│   │   ├── audit/               # audit-log service (context-aware logging)
│   │   ├── auth/                # JWT manager, password/PIN hashing
│   │   ├── authz/               # role → permission capability maps
│   │   ├── config/              # env config loading + validation
│   │   ├── database/            # pgx pool + migration runner
│   │   ├── handlers/            # HTTP handlers (auth, sales, products, …)
│   │   ├── middleware/          # auth, RBAC, rate limit, lockout, CORS, …
│   │   ├── models/              # domain models (JSON shapes)
│   │   └── repository/          # SQL data access layer
│   ├── migrations/              # 61 numbered up/down SQL migrations
│   ├── scripts/                 # seed_demo.go, reset_pin.go (go run helpers)
│   └── Dockerfile               # multi-stage build, non-root runtime user
├── frontend/                    # Next.js web app
│   └── src/
│       ├── app/                 # App Router: login, setup, dashboard pages
│       ├── components/          # ui/ primitives, layout/, shared/
│       ├── contexts/            # AuthProvider, ErrorProvider
│       ├── hooks/               # useUserEvents (SSE client)
│       ├── lib/                 # api/ client, auth session, permissions, config
│       └── types/               # shared TypeScript domain types
├── CLIENTS/                     # per-client env template + deployment guide
│   └── DEPLOY_NEW_CLIENT.md     # new-client installation procedure
├── ci/fixtures/                 # CI-only compose env fixture
├── docs/                        # internal planning notes
├── scripts/
│   ├── deploy-vps.sh            # orchestrate a VPS client deploy
│   └── render-caddyfile.sh      # generate Caddy routes from client envs
├── .github/workflows/
│   ├── ci.yml                   # tests, build, audit, secret scan
│   └── deploy.yml               # validate + deploy to VPS
├── docker-compose.yml           # local dev (db + backend)
├── docker-compose.prod.yml      # reusable production client stack
├── docker-compose.caddy.yml     # shared global Caddy container
├── Caddyfile                    # single-client template (generated in multi-client mode)
├── .env.example                 # root template for local dev / compose
├── VPS_DEPLOYMENT.md            # VPS architecture & operations guide
└── AGENTS.md                    # coding conventions for AI assistants
```

> Note: `desktop/` exists but is empty (placeholder). The desktop/Windows executable mentioned in the deployment docs is built from a client branch, not from this tree.

---

## Prerequisites

- **[Git](https://git-scm.com/)** — for cloning the repository.
- **[Docker & Docker Compose](https://www.docker.com/products/docker-desktop/)** — recommended way to run PostgreSQL (and optionally the backend API).
- **[Node.js](https://nodejs.org/)** 22+ & **npm** — for the Next.js frontend (CI uses Node 22; the frontend Docker image is Node 22).
- **[Go](https://go.dev/)** 1.25+ — only required if you want to run or test the backend natively outside Docker.
- **PostgreSQL client tools** (optional) — for `pg_dump` backups if running the DB natively.

---

## Getting Started (Local Development)

### 1. Clone the Repository

```bash
git clone <repository-url> dashpoint
cd dashpoint
```

### 2. Set Up Environment Variables

The root `.env.example` is the template for Docker Compose and local bare-metal runs:

```bash
cp .env.example .env
```

Default values are preconfigured for local development and work seamlessly with the provided Docker Compose setup. If you also want to run the backend natively, create `backend/.env` (the config loader reads it via godotenv); for the frontend, create `frontend/.env.local`:

```env
# frontend/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

> ⚠️ **Before any real or shared deployment:** replace the placeholder `JWT_SECRET` with a unique generated secret — `openssl rand -base64 48`. Never deploy with the example value. The backend enforces this: in production it refuses to start with a short secret, wildcard CORS, non-TLS `DATABASE_URL`, or `JWT_EXPIRY_MINUTES > 15`.

### 3. Start the Database and Backend API (Docker)

The easiest way to run the backend and its PostgreSQL database is Docker Compose. `docker-compose.yml` spins up:

- **`db`** — PostgreSQL 15 (pinned digest), with healthcheck and a named volume for data.
- **`backend`** — the Go API, built from `backend/Dockerfile`, running migrations automatically on startup, with a healthcheck against `/api/v1/health`.

```bash
docker compose up -d
```

> The `-d` flag runs containers in detached mode. View logs with `docker compose logs -f`.

Both services bind to loopback only (`127.0.0.1` and `[::1]`) — the database is never exposed to the network. Ports default to `BACKEND_PORT=8080` and `DB_PORT=5432` from `.env`.

The backend API is now available at **http://localhost:8080** (base path `/api/v1`).

### 4. Alternative: Run the Backend Natively

Start only the database container, then run Go directly:

```bash
docker compose up -d db
cd backend
go mod download
go run cmd/server/main.go
```

Configuration is read from the environment (and `backend/.env` if present). Required: `DATABASE_URL`, `JWT_SECRET`. The migration runner looks for `./migrations` and falls back to `./backend/migrations`.

### 5. Start the Frontend (Next.js)

In a new terminal:

```bash
cd frontend
npm install
npm run dev
```

The Next.js dev server starts on **http://localhost:3000**.

Other frontend scripts:

| Command         | Description                              |
|-----------------|------------------------------------------|
| `npm run dev`   | Start the development server             |
| `npm run build` | Production build                         |
| `npm run start` | Serve the production build               |
| `npm run lint`  | ESLint                                   |
| `npm test`      | Run Vitest suite once                    |
| `npm run test:watch` | Vitest in watch mode                |

### 6. First Run: Owner Setup

On first launch, the login screen shows the **initial owner setup** form (it checks `GET /api/v1/setup/status`, which reports `setup_required: true` only while the database has zero active users). Create the owner account with name, email, password (min 8 chars) and PIN (4–6 digits). Creation is atomic — once any active user exists, the endpoint refuses (`SETUP_ALREADY_COMPLETED`).

You can then sign in with **email + password** or **email-independent PIN login** from the login screen. "Save login" enables the persistent refresh cookie (automatic sign-in) and remembers the account in the account switcher.

### 7. Access the Application

Open your browser at **[http://localhost:3000](http://localhost:3000)**. The sidebar (Dashboard, POS, Shifts, Products, Inventory, Sales, Reports, Categories, Expenses, Users, Audit Logs, Settings, Recent Changes) is filtered by your permissions.

---

## Demo Data Seeding

A standalone Go program populates a rich demo dataset (Indonesian-rupiah pricing, 7 categories, 42 products, 13 shifts, 36 sales with payments, expenses, stock adjustments, and audit logs). It only runs when an owner already exists, refuses production without `ALLOW_PRODUCTION_SEED=true`, and cleans up its own `[seed]`-marked data on re-run.

```bash
cd backend
# full dataset (default):
SEED_MODE=full DATABASE_URL=postgres://dashpoint:dashpoint_dev@localhost:5432/dashpoint_dev?sslmode=disable \
  go run scripts/seed_demo.go
# minimal (users + products + inventory + shifts only):
SEED_MODE=core DATABASE_URL=... go run scripts/seed_demo.go
```

Seeded demo users (login with password or PIN — credentials are defined in `scripts/seed_demo.go` as bcrypt hashes for `manager@dashpoint.local`, `cashier@dashpoint.local`, etc.).

An owner must already exist — complete the owner setup in the UI (or deploy fresh) before seeding.

---

## Running Tests

```bash
# Backend: unit tests (no DB required)
cd backend
go test ./...

# Backend: vulnerability scan (as run in CI)
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...

# Backend: integration race test (requires a live database)
TEST_DATABASE_URL=postgres://dashpoint:dashpoint_dev@127.0.0.1:5432/dashpoint_dev?sslmode=disable \
  go test -tags=integration ./internal/repository/ -run TestCheckoutStockRace -v

# Frontend
cd frontend
npm test            # vitest run
npm run lint
npm run build
```

CI (`.github/workflows/ci.yml`) runs: frontend dependency audit (`npm audit --omit=dev --audit-level=high`), frontend tests, production frontend build, backend tests, govulncheck, and a secret-scanning pass that fails the build on committed private keys, cloud tokens, real IP-derived hostnames, or hardcoded `JWT_SECRET`/`POSTGRES_PASSWORD` values.

---

## Backend API Overview

Base URL: `http://localhost:8080/api/v1` (dev) or `https://<client-domain>/api/v1` (prod).

### Conventions

- All error responses share one shape: `{ "code": "<ERROR_CODE>", "message": "<human message>", "request_id": "<id>" }`.
- Authenticated requests use `Authorization: Bearer <access_token>`.
- Mutations accept strict JSON (unknown fields rejected on sensitive endpoints) and support optimistic concurrency via `expected_updated_at` (RFC3339) — a mismatch returns a "record changed" error.
- Deleting endpoints come in pairs: `DELETE /:id` (soft/archive) and `DELETE /:id/permanent` (hard delete).
- Successful responses return handler-specific JSON; list endpoints return paginated envelopes (`total`, `page`, `per_page`, `total_pages` or `limit`/`offset`).

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | API identity (`name`, `version`, `status`) |
| GET | `/api/v1/health` | Health with DB status; `503` when degraded |
| GET | `/api/v1/ping` | Lightweight liveness (`pong`) |
| GET | `/api/v1/setup/status` | `{ setup_required }` — true while no active user exists |
| POST | `/api/v1/setup/owner` | Create the first owner account (rate-limited; refuses once set up) |
| POST | `/api/v1/auth/login` | Email+password login → access token + refresh cookie (rate-limited + lockout) |
| POST | `/api/v1/auth/pin-login` | PIN login (rate-limited + lockout) |
| POST | `/api/v1/auth/refresh` | Rotate refresh cookie → new pair (cookie-gated rate limit) |
| POST | `/api/v1/auth/logout` | Revoke refresh token, clear cookie |
| GET | `/api/v1/events/subscribe` | SSE stream for the authenticated user (Bearer token) |

### Protected Endpoints

All routes below require a valid access token (middleware also re-checks account activity and token version). Permission gates are noted per group; **owner** bypasses permission checks.

**Session / Users / Roles**
| Method | Path | Permission |
|--------|------|------------|
| GET | `/me` | — |
| GET | `/users/basic` | — (id+name pairs for selectors) |
| GET / POST | `/users`, GET/PATCH/DELETE `/users/:id` | `access_users_page` / `manage_users_page` |
| PATCH | `/users/:id/password`, `/users/:id/pin` | self, or `manage_users_page` |
| DELETE | `/users/:id/permanent` | `manage_users_page` |
| GET | `/roles`, `/roles/:id` | `access_users_page` |
| PATCH | `/roles/:id/permissions` | **owner role only** |

**Catalog (Products / Categories / Inventory)**
| Method | Path | Permission |
|--------|------|------------|
| GET | `/products`, `/products/:id` | any of `access_products_page`, `access_inventory_page`, `access_pos_page` |
| GET | `/products/lookup?code=` | `access_pos_page` (barcode/SKU scan) |
| GET | `/products/:id/inventory` | `access_inventory_page` or `manage_inventory_page` |
| PATCH | `/products/:id/inventory` | `manage_inventory_page` (low-stock threshold) |
| POST / PATCH / DELETE | `/products…` | `manage_products_page` |
| GET | `/categories`, `/categories/:id` | any of `access_categories_page`, `access_products_page`, `manage_products_page`, `access_pos_page` |
| POST / PATCH / DELETE | `/categories…` | `manage_categories_page` |
| GET | `/inventory/low-stock` | `access_inventory_page` |
| POST | `/inventory/adjust` | `manage_inventory_page` (permission chosen by `adjustment_type`) |

**Operations (Shifts / Sales)**
| Method | Path | Permission |
|--------|------|------------|
| GET | `/shifts/current` | any of `access_pos_page`, `access_shifts_page`, `manage_shifts_page` |
| POST | `/shifts/start`, `/shifts/close`, `/shifts/pay-in`, `/shifts/pay-out` | `manage_shifts_page` |
| GET | `/shifts`, `/shifts/:id`, `/shifts/:id/operations` | `access_shifts_page` |
| POST | `/sales`, `/sales/validate` | `manage_pos_page` |
| GET | `/sales`, `/sales/:id`, `/sales/invoice/:invoiceNo`, `/sales/summary/daily` | `access_sales_page` |
| POST | `/sales/:id/void` | `manage_sales_page` |

**Reports** — all GET `/reports/*` require `access_reports_page` (`daily`, `sales`, `top-sellers`, `inventory`, `cash`, `by-employee`, `by-category`); all GET `/reports/export/*` require `manage_reports_page` (`sales`, `inventory`, `top-sellers`, `comprehensive`).

**Expenses**
| Method | Path | Permission |
|--------|------|------------|
| GET | `/expenses`, `/expenses/:id`, `/expenses/summary`, `/expenses/monthly` | `access_expenses_page` |
| POST / PATCH / DELETE | `/expenses…` | `manage_expenses_page` |
| GET | `/expenses/categories`, `/expenses/categories/:id` | `access_expenses_page` or `access_categories_page` |
| POST / PATCH / DELETE | `/expenses/categories…` | `manage_categories_page` |

**Audit / Dashboard / Uploads**
| Method | Path | Permission |
|--------|------|------------|
| GET | `/dashboard/changes` | `access_changes_page` |
| GET | `/logs`, `/logs/:id`, `/logs/actions`, `/logs/summary`, `/logs/entity/:type/:id`, `/logs/user/:id` | `access_audit_page` |
| GET | `/audit`, `/audit/:id` | `access_audit_page` |
| POST | `/upload/image` | `manage_products_page` or `manage_expenses_page` |
| DELETE | `/upload/image/:filename` | `manage_products_page` or `manage_expenses_page` |

Uploaded images are served publicly at `/uploads/<filename>` (GET-only, 1-hour cache, CORS-applied) — files are validated by declared Content-Type **and** magic-byte sniffing, capped at 5 MB, stored under a UUID name.

### Roles & Permissions

Permissions are flat snake_case strings (`access_*_page` / `manage_*_page`) covering: `pos`, `products`, `inventory`, `sales`, `reports`, `expenses`, `categories`, `users`, `shifts`, `changes`, `audit` domains. Roles are defined in `backend/internal/authz/role_capabilities.go`:

| Role | Permissions |
|------|-------------|
| `owner` | everything (plus implicit bypass in UI + sole access to role-permission editing) |
| `manager` | all 20 permissions (full operational + administrative access) |
| `cashier` | `access_pos_page`, `manage_pos_page`, `access_sales_page`, `access_shifts_page`, `manage_shifts_page` |

Permission checks run **server-side on every request** (`RequirePermission`, `RequireAnyPermission`, `RequirePermissionOrSelfParam`, `RequireRole`) and are mirrored client-side for navigation/route gating.

---

## Authentication & Security Model

The app uses a **two-token design with memory-only access tokens**:

1. **Access token (JWT HS256, 15 min max)** — returned in the login/refresh response body, kept in JavaScript memory only (never persisted), sent as `Bearer`. Claims include user, role, token type, and a `tv` (token version). Access-token lifetime is capped at 15 minutes in production by config validation.
2. **Refresh token (JWT, 7 days default)** — issued as an **httpOnly cookie** (`refresh_token`); a server-side record stores only its SHA-256 hash, grouped in a *token family*.

Key mechanisms (`backend/internal/handlers/auth_workflow.go`, `backend/internal/repository/refresh_token.go`):

- **Rotation**: every refresh mints a new pair and revokes the presented token. A 30-second *grace window* lets a sibling tab that just rotated still exchange (prevents cross-tab logout storms, since every page load refreshes).
- **Reuse detection**: replaying a rotated token beyond the grace window revokes the entire token family (assumed token theft).
- **Instant invalidation**: credential changes (password/PIN/role/permissions) bump `users.token_version`; the auth middleware rejects tokens whose version no longer matches — no waiting for expiry.
- **Force events**: deactivation/deletion/role change broadcasts an SSE event; clients log out or refresh immediately.

Additional hardening:

- **Rate limiting** — login/PIN: 10 req / 15 min / IP; refresh: 60 req / 15 min / IP (Fiber limiter).
- **Account lockout** — 5 consecutive failures on one email/user id locks that account for 15 minutes (in-memory, per-process), with `Retry-After`.
- **Trusted proxies** — `X-Forwarded-For/Proto` honored only from `TRUSTED_PROXIES` CIDRs; otherwise client IP is the socket address (spoof-proof for rate limits and audit).
- **Security headers** — `nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, restrictive `Permissions-Policy`, and HSTS only on genuinely-HTTPS requests.
- **Strict JSON parsing** with body-size caps on sensitive endpoints; generic 6 MB Fiber body limit (uploads).
- **Production config guardrails** — the server refuses to start unless: `JWT_SECRET` ≥ 32 chars, `JWT_EXPIRY_MINUTES ≤ 15`, `DATABASE_URL` uses `sslmode=require|verify-ca|verify-full`, and explicit wildcard-free `CORS_ORIGINS` is set.
- **No secrets in client payloads** — password/PIN hashes and token version are `json:"-"`.

---

## Database & Migrations

- PostgreSQL 15 (dev via Docker; production runs with TLS using per-client self-signed certs under `DATA_DIR/postgres-tls`).
- **61 sequential migrations** in `backend/migrations/` (`up`/`down` pairs) using golang-migrate, applied automatically at backend startup (idempotent).
- Core tables: `roles`, `users`, `refresh_tokens` (with `family_id` + `token_version` on users), `categories`, `products`, `inventory_items`, `stock_adjustments`, `shifts`, `sales`, `sale_items`, `payments`, `cash_drawer_operations`, `expenses`, `expense_categories`, `audit_logs`, `page_role_permissions`.
- Money and quantities use `NUMERIC` and `shopspring/decimal` end-to-end (no float money).
- Soft-delete pattern (`deleted_at`/`is_active`) across products, categories, users, and expense categories, with guarded permanent deletes.
- Migration history shows the schema was deliberately simplified over time (e.g. dropped customer fields from sales, simplified payments to a single record per sale, role-only RBAC after dropping the granular permission tables).

---

## Realtime Events (SSE)

- Endpoint: `GET /api/v1/events/subscribe` (Bearer-token authenticated, `text/event-stream`).
- Event types: `connected`, `user_updated`, `user_deactivated`, `user_activated`, `user_deleted`, `permissions_changed`, `role_changed`, `force_logout`.
- Client hook: `useUserEvents` parses the stream, auto-reconnects with exponential backoff (max 5 attempts), and retries on tab focus; `AuthProvider` reacts (refresh user, hard reload on role change, forced logout).
- Server caps: 3 streams per user, 500 total, 30-minute TTL with automatic reconnect, 15-second keepalive comments.

---

## Configuration Reference

### Backend (`backend/.env` / compose environment)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `ENVIRONMENT` | `development` | `development` or `production` (enables strict validation) |
| `DATABASE_URL` | — (required) | `postgres://user:pass@host:5432/db?sslmode=…` |
| `JWT_SECRET` | — (required) | HS256 signing key; ≥ 32 chars in production |
| `JWT_EXPIRY_MINUTES` | `15` | Access-token lifetime (≤ 15 in production) |
| `REFRESH_EXPIRY_HOURS` | `168` | Refresh-token lifetime (≤ 168) |
| `CORS_ORIGINS` | `http://localhost:3000` | Comma-separated exact origins; wildcard forbidden in production |
| `TRUSTED_PROXIES` | *(empty)* | Comma-separated IPs/CIDRs allowed to set `X-Forwarded-*` |
| `DEBUG` | `false` | `true` sets zerolog to debug level |

### Frontend (`frontend/.env.local` / build arg)

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080/api/v1` | API base URL used by the browser client |

### Compose / deployment (root `.env` or `CLIENTS/.env.<client>`)

| Variable | Purpose |
|----------|---------|
| `PROJECT_NAME` | Compose project + container prefix (must be unique per client) |
| `BACKEND_PORT`, `DB_PORT` | Host port bindings (dev) |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | Database credentials (unique per client) |
| `CADDY_SITE_ADDRESS` | Public hostname served by Caddy |
| `CADDY_API_ONLY` | API-only variant flag consumed by the Caddy renderer |
| `DATA_DIR` | Unique absolute host path for DB data, TLS files, uploads |
| `NEXT_PUBLIC_API_URL` | Set to `/api/v1` in production (same-origin via Caddy) |
| `PROXY_NETWORK` | External Docker network shared with Caddy (default `dashpoint_proxy`) |
| `TRUSTED_PROXIES` | Docker bridge subnet behind Caddy (e.g. `172.18.0.0/16`) |

---

## Deployment

### Single-Client Deployment

`docker-compose.prod.yml` defines a reusable production stack: `frontend`, `backend`, `db` on an internal network, with the backend/front on an external `proxy` network for Caddy. Frontend and backend images can be prebuilt (CI) or built locally. Caddy terminates TLS per hostname using `CADDY_SITE_ADDRESS` (see `Caddyfile`).

Validate, then start:

```bash
docker compose --env-file CLIENTS/.env.acme -p acme -f docker-compose.prod.yml config --quiet
docker compose --env-file CLIENTS/.env.acme -p acme -f docker-compose.prod.yml up -d
```

### Multi-Client VPS (SaaS Template)

DashPoint ships as a **multi-tenant SaaS template**: one VPS, one global Caddy, many isolated client stacks. Each client:

- is deployed from its own long-lived `clients/<client-slug>` branch (cut from `main`; bug fixes are cherry-picked, features merged deliberately after compatibility review),
- has a private `CLIENTS/.env.<slug>` file (mode 600, never committed),
- gets a unique `PROJECT_NAME`, `DATA_DIR`, database name/credentials, `JWT_SECRET`, and hostname,
- runs its own PostgreSQL, backend, and frontend containers — nothing is shared except the Caddy proxy network.

The complete step-by-step provisioning guide (VPS prep, branch strategy, env files, DNS, PostgreSQL TLS, first deploy, verification, GitHub Actions targeting, go-live checklist, and the mandatory per-client security requirements) lives in **[CLIENTS/DEPLOY_NEW_CLIENT.md](CLIENTS/DEPLOY_NEW_CLIENT.md)**.

The shared VPS architecture, operational rules, and data-safety policies are described in **[VPS_DEPLOYMENT.md](VPS_DEPLOYMENT.md)**.

Helper scripts:

```bash
# Regenerate Caddy routes from all client env files (validates names/hosts)
./scripts/render-caddyfile.sh CLIENTS Caddyfile

# One-command deploy of selected clients (validates compose config, up -d,
# renders+reloads Caddy, health-checks each public domain)
APP_DIR=/opt/dashpoint CLIENT_ENV_DIR=/opt/dashpoint/CLIENTS \
CLIENT_ENV_FILTER=.env.acme CADDY_FILE=/opt/dashpoint/Caddyfile \
./scripts/deploy-vps.sh
```

### CI/CD Pipeline

- **`ci.yml` ("CI")** — on pushes to `main` and all PRs: frontend audit/tests/build, backend `go test ./...`, govulncheck, secret scan. Also callable (`workflow_call`) by the deploy workflow.
- **`deploy.yml` ("Test And Deploy")** — triggered after a successful CI run on a `clients/<slug>` push (deploys **only that client** via its env file) or manually via `workflow_dispatch` with an optional `client_env` filter (empty = deploy all active clients). The pipeline:
  1. Runs CI (on dispatch) and validates the production Compose file and the rendered Caddyfile.
  2. Streams the exact source commit to the VPS as a Git archive over SSH (`tar`), recording `.deploy-commit`.
  3. **Builds both Docker images on the runner** and ships them with `docker save | ssh docker load`, tagged with the source commit (avoids OOM kills on small VPSes).
  4. Runs `scripts/deploy-vps.sh` remotely (detached, with polling): `up -d` from loaded images, migration on startup, Caddy validate+reload, and an HTTPS health check per deployed domain.
  5. Selects VPS credentials from GitHub Environment secrets (`deploy-<slug>` per branch, repository-level fallback).

  Required secrets: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`, `VPS_KNOWN_HOSTS`, `VPS_APP_DIR`, `VPS_CLIENT_ENV_DIR`, `VPS_CADDY_FILE`.

Pushes to `main` never deploy; deploys are branch-scoped and env-file-scoped so co-located clients can never touch each other.

### Backups & Data Safety

```bash
# Database dump per client
docker exec acme-db-prod sh -c \
  'pg_dump -Fc --no-owner --no-acl -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  > /opt/dashpoint/backups/acme-$(date +%F).dump
chmod 600 /opt/dashpoint/backups/acme-*.dump

# Uploads archive
tar -czf /opt/dashpoint/backups/acme-uploads-$(date +%F).tar.gz \
  /opt/dashpoint/clients/acme/data/uploads
```

- Normal deployment **never** replaces client databases; imports/resets are explicit, separate operations.
- Never run `docker compose down --volumes` or delete `DATA_DIR` unless destroying client data intentionally. Stop-while-preserving-data is plain `docker compose down`.
- Keep multiple restore points and test restores before relying on them.

---

## Operational Scripts

| Script | Purpose |
|--------|---------|
| `backend/scripts/seed_demo.go` | Populate/clean a rich demo dataset (`SEED_MODE=core|full`, refuses production without `ALLOW_PRODUCTION_SEED=true`) |
| `backend/scripts/reset_pin.go` | Reset a user's PIN directly: `go run scripts/reset_pin.go -email <email> -pin <4-6 digits>` |
| `scripts/render-caddyfile.sh` | Generate the shared Caddyfile from all `CLIENTS/.env.*` files with validation |
| `scripts/deploy-vps.sh` | Full client deploy: compose validate → up → Caddy render/validate/reload → per-domain health check |

---

## Troubleshooting

- **Database connection issues** — Ensure Docker is running; the backend waits for the `db` healthcheck (`pg_isready`). Verify the `db` container is healthy before the backend starts: `docker compose ps`.
- **Port conflicts** — Ports `8080` (backend), `5432` (Postgres), and `3000` (frontend) must be free, or override `BACKEND_PORT`/`DB_PORT` in `.env` / the frontend start port.
- **`401` loops in the browser** — The transport auto-refreshes on any 401 and redirects to `/login` if refresh fails; clear the `refresh_token` cookie for the site if a stale cookie lingers.
- **Locked out / rate limited** — Login allows 10 attempts per IP per 15 min; 5 consecutive account failures lock the account for 15 minutes. Both reset with time (and a restart clears lockouts).
- **Refresh rejected (`INVALID_TOKEN`)** — Refresh tokens rotate on every use; replaying an old cookie beyond the 30-second grace window revokes the family by design. Sign in again.
- **Uploads rejected** — Images only (JPEG/PNG/GIF/WebP), ≤ 5 MB, validated by magic bytes; the Fiber body limit is 6 MB total.
- **SSE not connecting** — Check the bearer token is valid and that proxies don't buffer `text/event-stream` (backend sends `X-Accel-Buffering: no`).
- **VPS builds OOM-killed** — The VPS needs swap (see VPS_DEPLOYMENT.md) — but preferably rely on CI, which builds images on the runner and ships them prebuilt.
- **Health check fails after deploy** — `docker logs --tail=100 <project>-backend-prod`; the deploy script prints logs automatically on failure.
- **Old Docker images accumulate** — CI tags images per commit; occasionally run `docker image prune -f` on the VPS.

---

## Additional Documentation

| Document | Contents |
|----------|----------|
| [CLIENTS/DEPLOY_NEW_CLIENT.md](CLIENTS/DEPLOY_NEW_CLIENT.md) | Full new-client installation: VPS, branch, env, DNS, TLS, first deploy, verification, CI targeting, go-live checklist, mandatory security requirements |
| [VPS_DEPLOYMENT.md](VPS_DEPLOYMENT.md) | Multi-client VPS architecture, distribution/branching model, operations, backups, CI/CD secrets |
| [CLIENTS/.env.example](CLIENTS/.env.example) | Safe template for per-client production env files |
| [.env.example](.env.example) | Template for local development and root compose variables |
| [AGENTS.md](AGENTS.md) | Architectural and behavioral conventions for AI coding assistants working in this repo |

---

Built with Go, Next.js, PostgreSQL, Docker, and Caddy. Designed to be forked per client.
