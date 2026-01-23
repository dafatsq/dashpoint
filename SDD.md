# POS System — Software Design Document

**Title:** POS for Local Indonesian Store (Owner, Manager, Cashier)

**Version:** 1.0

**Date:** 2026-01-17

**Prepared by:** ChatGPT

---

# 1. Executive Summary

This document describes the software design for a Point-of-Sale (POS) system targeted at average local Indonesian stores. The system supports three roles: **Owner**, **Manager**, and **Cashier**. It focuses on quick, offline-capable sales, inventory, basic accounting reports, and role-based permissions with per-employee feature access control.

Three deployment/test phases are covered:

- **Phase 1 — Local simulation & testing** (developer laptop / MacBook):
  - Frontend: Next.js (local)
  - Backend: Go + Fiber (local MacBook)
  - DB: PostgreSQL (local MacBook)

- **Phase 2 — Remote simulation & testing** (old laptop as VPS with non-static IP):
  - Frontend: Next.js (Vercel)
  - Backend: Go + Fiber (old laptop running Ubuntu, SSH)
  - Cloudflare Tunnel (to expose backend securely over non-static IP)
  - DB: PostgreSQL (old laptop on Ubuntu)

- **Production:**
  - Frontend: Next.js (Vercel)
  - Backend: Go + Fiber (real VPS)
  - DB: PostgreSQL (real VPS)

Primary focus for Phase 1 is features and UI; Phase 2 validates connectivity, tunnels, and near-production flows; Production uses hardened servers.

---

# 2. Goals & Non-Goals

**Goals**
- Fast, reliable checkout flow for cash/card payments (supports manual card entry and integrated/card-on-file later).
- Simple inventory management (stock counts, low-stock alerts).
- Role-based access control (Owner > Manager > Cashier) with per-employee per-feature toggles.
- Sales history, daily reports, basic analytics and export (CSV).
- Easy deployment: local dev -> tunnel -> production VPS.

**Non-Goals (v1)**
- Integrated payment provider (v1 will support offline/cash, v2 add payment gateway integration).
- Complex accounting, payroll, multi-location sync (reserved for future releases).

---

# 3. Personas & Roles

- **Owner** (full control): configure store, financial reports, add/remove managers, manage permissions, export data, backup/restore, system settings.
- **Manager** (operational control): manage inventory, promotions, view reports (less detailed), create/edit products, adjust stock, manage cashiers' access to features.
- **Cashier** (operational): process sales, refunds (if allowed), view product search, apply discounts (if enabled), view own shift summary.

**Per-employee feature access**: The Owner/Manager can toggle access for any employee for any specific feature — e.g., permit Cashier A to do returns but not apply discounts, or permit Manager to edit prices but not delete products.

---

# 4. Functional Requirements

## 4.1 Core Features
- **Login / Authentication**
  - Email/password for Owner; PIN or username/password for employees.
  - JWT-based sessions with refresh tokens.

- **Role & Permission Management**
  - Roles: owner, manager, cashier.
  - Permissions: granular Boolean flags (e.g., `can_refund`, `can_edit_price`, `can_adjust_stock`, `can_view_reports`, `can_manage_users`).
  - Per-employee override: each employee record stores explicit allowed/denied feature flags.

- **Products & Inventory**
  - Product catalog with SKU, barcode, name, description, cost, price, tax rate, unit, category.
  - Stock entries: quantity on hand, location (single store in v1), low-stock threshold, last counted date.
  - Stock adjustments with audit trails (who, reason, before/after).

- **Sales / Checkout**
  - Add items by SKU/barcode or search.
  - Apply discounts (percentage or fixed) if permission allowed.
  - Accept payment types: cash (v1), card (manual entry v1), store credit (optional), voucher (optional)
  - Generate receipts (print / PDF / email optional later)
  - Transaction record persists with items, taxes, payment method, employee ID, shift ID.

- **Returns & Refunds**
  - Refund flow with permission checks and reason required. Partial refunds supported.

- **Reports**
  - Daily sales, top-sellers, inventory valuation, cash reports. Export to CSV.

- **User Management**
  - Create employees, set role, set individual permission flags, set PIN, activate/deactivate.

- **Backup & Export**
  - Manual export of DB dump or CSVs. Owners can create scheduled backup scripts in production.

## 4.2 Admin Features
- Role creation/editing (Owner). Default base role definitions included.
- Audit logs (who did what, when).

---

# 5. Non-Functional Requirements

- **Availability:** System should be available during business hours; single-VPS deployment target ~99% uptime.
- **Performance:** Checkout should process in <200ms for DB lookup on a local network; UI respond within 100ms for interactive actions.
- **Security:** Encrypt sensitive data in transit (HTTPS), hash passwords (bcrypt), protect JWTs, rate-limit auth endpoints.
- **Scalability:** Design to allow vertical scaling in v1 and easy migration to multi-instance setup in v2.
- **Localization:** Bahasa Indonesia language support and currency IDR formatting.

---

# 6. High-Level Architecture

Components:
- **Frontend (Next.js)**: React-based SPA with server-side rendering for public pages and client-side rendering for POS interface. Deploy to Vercel.
- **Backend (Go + Fiber)**: RESTful API + optional WebSocket for real-time inventory updates. Auth, business logic, DB access.
- **Database (PostgreSQL)**: Relational store for all transactional and master data.
- **Cloudflare Tunnel (Phase 2)**: Expose backend running on non-static IP securely to Vercel frontend.
- **Logging & Monitoring**: File logs + simple metrics (Prometheus or hosted alternative in production later).

Diagram (text):

```
[Next.js (Browser)] <----HTTPS----> [Vercel frontend] --(API calls over HTTPS)--> [Cloudflare Tunnel or direct] --> [Go Fiber backend on VPS/old laptop/localhost] --> [Postgres DB]
```

Real-time updates (optional): WebSocket or Server-Sent Events between backend and frontend for inventory/notifications.

---

# 7. Data Model (ER Summary)

Primary tables:
- `users` (employees + owner)
- `roles` (owner, manager, cashier)
- `permissions` (list of permission keys)
- `user_permissions` (overrides)
- `products`
- `inventory_items` (product_id, quantity_on_hand)
- `stock_audits` (adjustments history)
- `sales` (transactions)
- `sale_items` (lineitems)
- `payments` (sale_id -> payment details)
- `shifts` (employee shifts, cash drawer)
- `logs` (audit logs)

### Example table schemas (simplified)

```sql
-- users
CREATE TABLE users (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT UNIQUE,
  pin TEXT, -- hashed PIN for cashier quick-login
  password_hash TEXT,
  role_id UUID REFERENCES roles(id),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT now()
);

-- roles
CREATE TABLE roles (
  id UUID PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  description TEXT
);

-- permissions (each row is a permission key description)
CREATE TABLE permissions (
  id UUID PRIMARY KEY,
  key TEXT UNIQUE NOT NULL,
  description TEXT
);

-- user_permissions override table
CREATE TABLE user_permissions (
  user_id UUID REFERENCES users(id),
  permission_id UUID REFERENCES permissions(id),
  allowed BOOLEAN,
  PRIMARY KEY (user_id, permission_id)
);

-- products
CREATE TABLE products (
  id UUID PRIMARY KEY,
  sku TEXT UNIQUE,
  barcode TEXT UNIQUE,
  name TEXT,
  description TEXT,
  category TEXT,
  price NUMERIC(12,2),
  cost NUMERIC(12,2),
  tax_rate NUMERIC(5,2),
  created_at TIMESTAMP DEFAULT now()
);

-- inventory_items
CREATE TABLE inventory_items (
  product_id UUID REFERENCES products(id) PRIMARY KEY,
  quantity NUMERIC(12,3) DEFAULT 0,
  low_threshold NUMERIC(12,3) DEFAULT 0,
  last_counted_at TIMESTAMP
);

-- sales
CREATE TABLE sales (
  id UUID PRIMARY KEY,
  invoice_no TEXT UNIQUE,
  created_at TIMESTAMP DEFAULT now(),
  total NUMERIC(12,2),
  tax NUMERIC(12,2),
  payment_status TEXT,
  employee_id UUID REFERENCES users(id),
  shift_id UUID REFERENCES shifts(id)
);

-- sale_items
CREATE TABLE sale_items (
  id UUID PRIMARY KEY,
  sale_id UUID REFERENCES sales(id),
  product_id UUID REFERENCES products(id),
  quantity NUMERIC(12,3),
  unit_price NUMERIC(12,2),
  discount NUMERIC(12,2)
);
```

---

# 8. API Design (Representative endpoints)

All endpoints are prefixed with `/api/v1`.

**Auth**
- `POST /api/v1/auth/login` — { email, password } → returns `{ access_token, refresh_token, user }`
- `POST /api/v1/auth/pin-login` — { user_id, pin } → for cashier quick login
- `POST /api/v1/auth/refresh` — { refresh_token }

**Users & Roles**
- `GET /api/v1/users` — list (Owner/Manager)
- `POST /api/v1/users` — create user
- `PATCH /api/v1/users/:id/permissions` — set per-user permissions (Owner/Manager)
- `GET /api/v1/roles` — reference roles

**Products & Inventory**
- `GET /api/v1/products` — query products
- `POST /api/v1/products` — create product
- `PATCH /api/v1/products/:id` — edit
- `GET /api/v1/inventory/:product_id` — view stock
- `POST /api/v1/inventory/adjust` — adjust stock (records audit)

**Sales**
- `POST /api/v1/sales` — commit a sale (items[], payment details)
- `GET /api/v1/sales/:id` — sale details
- `POST /api/v1/sales/:id/refund` — refund (permission-checked)

**Reports**
- `GET /api/v1/reports/daily?date=YYYY-MM-DD` — daily summary
- `GET /api/v1/reports/top-sellers?start=...&end=...`

**Audit & Logs**
- `GET /api/v1/logs` — view audit logs (Owner)

**Health**
- `GET /api/v1/health` — quick health check

**Notes on API**
- Use standard HTTP status codes.
- Validate payloads using schema (e.g., go-playground/validator or equivalent).
- Return consistent error payload `{ code, message, details? }`.

---

# 9. Authentication & Authorization

- **Auth**: JWT access token (short-lived, e.g., 15–60 min) + refresh token (longer). HTTPS required.
- **Password/PIN**: Password hashed with bcrypt/argon2. PIN hashed as well and limited attempts.
- **RBAC**: Evaluate permissions by combining role defaults + per-user overrides. Permission resolution order:
  1. If `user_permissions` entry exists for permission → use it.
  2. Else use `role` default attributes stored as role-to-permission mapping.

- **Session revocation**: Store active refresh tokens to allow logout/revoke.

---

# 10. UI / UX

## 10.1 UI Goals (Hard Requirements)

The POS UI must feel **modern, fast, and premium**. This is a strict requirement.

**Must-haves**
- **Fully responsive** UI across:
  - Mobile (min width 360px)
  - Tablet (768px)
  - Desktop (>=1024px)
- **POS friendly layout**:
  - Large touch targets (for touch screens)
  - Keyboard-first support for desktop cashiers
- **Polished states**:
  - Loading states (skeletons)
  - Error states (friendly messages)
  - Empty states
  - Toast notifications
- **Accessibility basics**:
  - Focus-visible rings
  - ARIA labels where needed
  - High contrast for readability

## 10.2 UI Tech Stack Standards

The frontend must use the following libraries unless there is a strong technical reason not to:

- **Tailwind CSS** — main styling system
- **shadcn/ui** (Radix UI primitives) — reusable components
- **lucide-react** — icons
- **react-hook-form** — forms
- **zod** — schema validation
- **@hookform/resolvers** — RHF + Zod integration
- **tanstack/react-query** (or SWR) — data fetching, caching, invalidation
- **date-fns** — date handling

Optional / recommended:
- **sonner** or **react-hot-toast** — toasts
- **clsx + tailwind-merge** — class composition

## 10.3 Design System & Consistency

- Use consistent spacing, typography scale, and component sizing.
- Use a small set of design tokens:
  - `--radius` for rounding
  - consistent shadow levels
  - consistent button heights
- Pages should follow a common layout structure:
  - Top navigation / sidebar (admin screens)
  - Quick action bar (POS screen)

## 10.4 UX for POS speed

**Keyboard shortcuts (recommended defaults)**
- `F1` Search product
- `F2` Add manual item
- `F3` Apply discount (if allowed)
- `F4` Checkout
- `Esc` Close modal / clear focus

**Checkout screen requirements**
- Barcode input always available
- Cart visible at all times
- Payment panel is simple and fast
- After successful sale: show success + print/export receipt

## 10.5 Permission-Aware UI

The UI must:
- Hide or disable actions the user does not have access to
- Provide clear explanation tooltips for disabled features (optional)
- Prevent navigation to restricted pages (route guards)

---

# 11. Data Flows & Sequences

## 11.1 Sale creation (happy path)
1. Cashier scans item barcode or searches and adds to cart.
2. Frontend validates item existence; calculates line total (tax, discounts) locally.
3. On ``Checkout`` click, frontend sends `POST /api/v1/sales` with items, totals, payment.
4. Backend validates stock availability; begins DB transaction.
5. Insert `sales` record → insert `sale_items` → decrease `inventory_items.quantity` per product.
6. Commit transaction. If any step fails, rollback and return error.
7. Backend returns sale ID and invoice number; frontend prints receipt and shows success.

## 11.2 Stock adjustment
- Manager/Owner requests adjustment → `POST /api/v1/inventory/adjust` with reason.
- Backend writes `stock_audits` record and updates `inventory_items` quantity.

---

# 12. Concurrency & Consistency

- Use DB transactions for sales to ensure stock consistency.
- Use `SELECT ... FOR UPDATE` when adjusting inventory to prevent race conditions when multiple cashiers checkout simultaneously.
- Consider optimistic locking with `version` fields for product edits.

---

# 13. Testing Strategy

- Unit tests for core business logic (Go) and component tests for React components.
- Integration tests for sale flows and inventory updates (local Postgres test DB).
- End-to-end tests using Playwright or Cypress to run flows from UI.
- Load testing (simple): simulate bursts of checkout requests to ensure DB locking and API performance.

---

# 14. Deployment & Phase-specific Notes

## Phase 1 — Local (Dev / Simulation)
- Run Next.js locally with `next dev`.
- Backend: run `go run ./cmd/server` on MacBook.
- DB: local Postgres instance.
- Provide `docker-compose.yml` for local dev convenience (optional).

**Env variables (local)**
```
DATABASE_URL=postgres://user:pass@localhost:5432/pos_dev
JWT_SECRET=dev-secret
PORT=8080
```

## Phase 2 — Old laptop (simulation with non-static IP)
- Deploy Next.js to Vercel and configure environment to point to backend URL exposed via Cloudflare Tunnel.
- On old laptop run Cloudflare Tunnel (cloudflared) to expose backend:
  - `cloudflared tunnel --hostname pos.yourdomain.example http://localhost:8080`
  - If you don't have a domain, use a Cloudflare Tunnel random domain or use `ngrok` as alternative.
- Ensure TLS termination at Cloudflare; Backend still should require HTTPS via Cloudflare.
- Set up Postgres on the old laptop; ensure firewall allows local-only binds and Cloudflare handles external exposure.

Note: Cloudflare Tunnel removes the need for a static IP — it creates an outbound connection from your server to Cloudflare.

## Production
- Provision VPS with stable provider.
- Hardening: create non-root user, enable UFW, install Fail2Ban, secure Postgres (listen addresses), use managed Postgres if available.
- Use TLS certs (Let's Encrypt) or keep Cloudflare in front.
- Backups: schedule `pg_dump` and keep copies off-machine (S3 or remote storage).
- CI/CD: push Next.js to Vercel automatically; backend deploy with GitHub Actions to provisioned VPS or Docker container.

---

# 15. Backup & Recovery

- Daily DB dump (cron) and keep 7/30-day retention.
- Option to download on-demand export CSV for sales/inventory.
- For quick recovery, keep SQL migration scripts (use goose or golang-migrate).

---

# 16. Observability & Logging

- Structured logs (JSON) written to stdout to be captured by systemd or container log collectors.
- Simple health endpoint `/api/v1/health`.
- In production add Prometheus metrics (or third-party hosted solution) and alerting for DB down, disk full, tunnel down.

---

# 17. Security Considerations

- Enforce HTTPS at all times.
- Hash passwords with bcrypt/argon2.
- Limit login attempts (esp. PIN login); temporary lockout after failed attempts.
- Sanitize inputs and validate all payloads.
- Run DB backups to an external location.
- Secure access to VPS (SSH keys only).
- Minimize data held: avoid storing full card data; use tokenization if integrating card processors later.

---

# 18. Migration & Upgrade Path

- Use `golang-migrate` for DB migrations. Keep migrations in `migrations/` folder.
- For Phase 1 → Phase 2:
  - Swap DB connection string to remote Postgres on old laptop.
  - Expose backend via Cloudflare Tunnel and change `NEXT_PUBLIC_API_URL` in Vercel.
- For Phase 2 → Production:
  - Provision stable VPS, restore DB backups to production Postgres, move records, update DNS to production domain, retire Cloudflare Tunnel (or keep Cloudflare in front).

---

# 19. CI/CD & DevOps

- Frontend: Vercel (automatic from repo on push to `main`).
- Backend: GitHub Actions to build and push Docker image to registry, then deploy to VPS (SSH + `docker-compose`), or use systemd to run binary.
- DB schema migrations executed during deploy using `golang-migrate`.

---

# 20. Risks & Assumptions

- Assumes single store location (no multi-store sync in v1).
- Using old laptop with non-static IP: Cloudflare Tunnel required — if Cloudflare connectivity broken, backend becomes unreachable externally.
- Payment gateway integration is out of scope for v1.

---

# 21. Appendix

## 21.1 Sample `docker-compose.yml` for local dev

```yaml
version: "3.8"
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_USER: pos
      POSTGRES_PASSWORD: pos
      POSTGRES_DB: pos_dev
    ports:
      - "5432:5432"
    volumes:
      - ./pgdata:/var/lib/postgresql/data

  backend:
    build: ./backend
    command: ./server
    environment:
      DATABASE_URL: postgres://pos:pos@db:5432/pos_dev?sslmode=disable
      JWT_SECRET: dev-secret
    ports:
      - "8080:8080"
    depends_on:
      - db

  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080/api/v1
    depends_on:
      - backend
```

## 21.2 Sample permission keys (starter)
- `can_create_sale`
- `can_refund`
- `can_apply_discount`
- `can_edit_price`
- `can_adjust_stock`
- `can_view_reports`
- `can_manage_users`

---

# 22. Next Steps (Suggested)

1. Approve SDD content or request changes.
2. Create initial repo skeleton (frontend & backend) and seed migrations + sample data.
3. Implement core checkout + inventory flow and run Phase 1 locally.
4. After Phase 1 stable, deploy Phase 2 using Cloudflare Tunnel for remote testing.

---

*End of document.*

