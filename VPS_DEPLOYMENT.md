# DashPoint VPS Deployment Guide

This guide describes the target VPS-only deployment for DashPoint, where the frontend, backend, database, and Caddy reverse proxy all run on the same VPS through Docker Compose.

Use this guide for the first VPS deployment. For adding more clients to the same VPS later, reuse the same architecture with a separate Compose project and env file per client.

## Target Architecture

```text
Internet
  -> Caddy container on ports 80/443
    -> frontend container on private Docker port 3000
    -> backend container on private Docker port 8080
    -> postgres container on private Docker port 5432
```

Public traffic should enter through Caddy only.

Recommended URL shape:

```text
https://pos.example.com/        -> frontend
https://pos.example.com/api/v1  -> backend
```

The database must not be exposed publicly. It should be reachable only by the backend over Docker networking.

## What Changes From Vercel Hosting

Old deployment:

```text
Vercel serves frontend
VPS serves backend and database
Caddy proxies backend
```

New deployment:

```text
VPS serves frontend, backend, database, and Caddy
Caddy proxies frontend and backend
Vercel is no longer involved
```

Expected project changes:

- Add a production frontend Dockerfile.
- Add frontend service to the production Compose stack.
- Route frontend and backend through Caddy.
- Use same-origin frontend API config: `NEXT_PUBLIC_API_URL=/api/v1`.
- Keep backend and database private inside Docker networking.
- Replace Vercel auto-deploy with GitHub Actions or manual VPS deploy commands.

## Required Repo Files

Before this deployment can be run end-to-end, the repo should contain production VPS deployment files like these:

```text
frontend/Dockerfile
docker-compose.vps.yml
Caddyfile
.env.example
```

If these files are not implemented yet, treat this document as the deployment target and complete the deployment conversion first.

## Domain Requirement

Do not use the raw VPS IP as the long-term production URL.

Recommended:

```text
pos.example.com -> VPS public IP
```

Temporary testing options:

```text
165-245-184-75.nip.io
165.245.184.75.sslip.io
```

For production, use a real domain or client-specific subdomain.

## VPS Prerequisites

Use a clean Ubuntu/Debian VPS or equivalent.

Required:

- Docker installed.
- Docker Compose plugin installed.
- SSH access configured.
- Firewall allows only required public ports.
- A domain or temporary DNS name points to the VPS public IP.

Recommended public firewall:

```text
22/tcp    SSH
80/tcp    HTTP for Caddy/Let's Encrypt
443/tcp   HTTPS for Caddy
```

Do not expose:

```text
3000/tcp  frontend container
8080/tcp  backend container
5432/tcp  postgres container
```

Those should stay private inside Docker unless there is a deliberate debugging reason.

## One-Time VPS Setup

Create a deploy directory:

```bash
sudo mkdir -p /opt/dashpoint
sudo chown "$USER":"$USER" /opt/dashpoint
cd /opt/dashpoint
```

Clone the repo:

```bash
git clone <your-github-repo-url> .
```

Create the shared proxy network:

```bash
docker network create dashpoint_proxy
```

If the network already exists, Docker will report that. That is fine.

## Environment File Strategy

Never commit real production `.env` files.

For the first deployment, create a VPS-only env file:

```bash
nano .env.production
chmod 600 .env.production
```

Example:

```env
# Stack identity
PROJECT_NAME=dashpoint-demo
SITE_HOST=pos.example.com

# Backend app
ENVIRONMENT=production
PORT=8080
CORS_ORIGINS=https://pos.example.com
JWT_SECRET=replace_with_at_least_32_random_characters
JWT_EXPIRY_MINUTES=15
REFRESH_EXPIRY_HOURS=168
DEBUG=false

# Frontend app
NEXT_PUBLIC_API_URL=/api/v1
NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS=false

# Database
POSTGRES_USER=dashpoint_demo
POSTGRES_PASSWORD=replace_with_long_random_database_password
POSTGRES_DB=dashpoint_demo
DATABASE_URL=postgres://dashpoint_demo:replace_with_long_random_database_password@db:5432/dashpoint_demo?sslmode=disable

# Only use this if the backend supports it and the DB is private inside Docker/VPS networking.
DATABASE_ALLOW_LOCAL_INSECURE=true
```

Important:

- `JWT_SECRET` must be strong and at least 32 characters.
- `CORS_ORIGINS` must match the public site origin.
- `NEXT_PUBLIC_API_URL=/api/v1` is preferred for same-origin VPS hosting.
- `POSTGRES_PASSWORD` and the password inside `DATABASE_URL` must match.
- `DATABASE_URL` should point to Docker service host `db`, not `localhost`.

## Database SSL Decision

There are two valid production options:

### Option A: Private Docker DB Without TLS

This is practical for a single VPS where backend and DB communicate only inside Docker networking.

Use:

```env
DATABASE_URL=postgres://user:password@db:5432/dbname?sslmode=disable
DATABASE_ALLOW_LOCAL_INSECURE=true
```

Only use this if:

- Postgres is not exposed to the public internet.
- Backend and DB run on the same VPS/private Docker network.
- The backend explicitly supports `DATABASE_ALLOW_LOCAL_INSECURE=true`.

If this flag is not implemented yet, implement it before deployment or use Option B.

### Option B: TLS-Enabled Postgres

Use this if the database is remote or publicly reachable.

Use one of:

```env
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=verify-ca
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=verify-full
```

Remote production databases should not use `sslmode=disable`.

## Caddy Routing

For a single client/domain:

```caddyfile
pos.example.com {
  handle /api/v1/* {
    reverse_proxy backend:8080
  }

  handle {
    reverse_proxy frontend:3000
  }
}
```

For local testing without a real domain:

```caddyfile
:8083 {
  handle /api/v1/* {
    reverse_proxy backend:8080
  }

  handle {
    reverse_proxy frontend:3000
  }
}
```

Caddy should be the only public web entrypoint.

## Compose Stack Shape

The production Compose stack should include:

```text
caddy
frontend
backend
db
```

The backend should depend on `db`.

The frontend should depend on nothing or only on the backend if needed for health checks.

Caddy should depend on frontend and backend.

The DB should use a named volume or per-client host path that cannot be shared accidentally with another client.

Example volume names:

```text
${PROJECT_NAME}_postgres_data
${PROJECT_NAME}_uploads
caddy_data
caddy_config
```

Avoid static shared paths like these for multi-client deployments:

```text
./data/postgres
./data/uploads
```

Static paths can mix client data if multiple stacks run from one project folder.

## First Deployment

From the VPS project folder:

```bash
cd /opt/dashpoint
git pull
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml config
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml up -d --build
```

Check containers:

```bash
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml ps
```

Check logs:

```bash
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml logs --tail=100 backend
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml logs --tail=100 frontend
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml logs --tail=100 caddy
```

Check backend through Caddy:

```bash
curl -i https://pos.example.com/api/v1/ping
```

Expected:

```text
HTTP/2 200
{"message":"pong"}
```

Check frontend:

```bash
curl -I https://pos.example.com/
```

Expected:

```text
HTTP/2 200
```

Then open:

```text
https://pos.example.com
```

## Initial Database Seeding

Only seed the database intentionally.

For a demo environment:

```bash
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml exec backend ./server
```

If seeding is performed through Go scripts instead of app startup/migrations, run it deliberately from the backend environment:

```bash
cd /opt/dashpoint/backend
ENVIRONMENT=production \
ALLOW_PRODUCTION_SEED=true \
DATABASE_URL='postgres://user:password@db:5432/dbname?sslmode=disable' \
SEED_MODE=core \
go run ./scripts/seed_demo.go
```

Do not run demo seed data against a real client production database unless that is explicitly intended.

## Verification Checklist

After deployment, verify:

- Domain points to the VPS IP.
- Caddy obtained HTTPS certificate.
- `https://<domain>/` loads the frontend.
- `https://<domain>/api/v1/ping` returns `{"message":"pong"}`.
- Browser login works.
- API calls use the same domain under `/api/v1`.
- No browser CORS errors.
- Backend logs show no fatal config errors.
- Frontend logs show no build/start errors.
- Postgres container is not publicly exposed.
- Upload/image features work.
- Security headers are present:

```bash
curl -I https://pos.example.com/api/v1/ping
```

Look for:

```text
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: no-referrer
Permissions-Policy: camera=(), microphone=(), geolocation=(), payment=()
```

## Multi-Client Model On One VPS

Use one project folder and multiple Compose projects.

Example:

```text
/opt/dashpoint
  CLIENTS/.env.demo
  CLIENTS/.env.alzauk
  CLIENTS/.env.xenon
  docker-compose.vps.yml
  Caddyfile
  backend/
  frontend/
```

Deploy each stack separately:

```bash
docker compose --env-file CLIENTS/.env.demo -p demo -f docker-compose.vps.yml up -d --build
docker compose --env-file CLIENTS/.env.alzauk -p alzauk -f docker-compose.vps.yml up -d --build
docker compose --env-file CLIENTS/.env.xenon -p xenon -f docker-compose.vps.yml up -d --build
```

Each client must have:

- unique `PROJECT_NAME`
- unique `SITE_HOST`
- unique DB credentials
- unique DB name
- unique JWT secret
- isolated DB volume
- isolated upload volume

Do not share DB volumes or upload volumes between clients.

## Multi-Client Caddy Example

```caddyfile
demo.example.com {
  handle /api/v1/* {
    reverse_proxy demo-backend:8080
  }

  handle {
    reverse_proxy demo-frontend:3000
  }
}

alzauk.example.com {
  handle /api/v1/* {
    reverse_proxy alzauk-backend:8080
  }

  handle {
    reverse_proxy alzauk-frontend:3000
  }
}

xenon.example.com {
  handle /api/v1/* {
    reverse_proxy xenon-backend:8080
  }

  handle {
    reverse_proxy xenon-frontend:3000
  }
}
```

For this to work, Caddy and each client stack must share a Docker network such as:

```bash
docker network create dashpoint_proxy
```

## CI/CD Model

Caddy does not deploy code. Caddy only routes traffic.

To replace Vercel auto-deploy, use GitHub Actions:

```text
git push
  -> GitHub Actions runs backend tests
  -> GitHub Actions runs frontend tests/build
  -> GitHub Actions SSHs into VPS
  -> VPS runs git pull
  -> VPS runs docker compose up -d --build
```

GitHub Secrets should contain deploy access only:

```text
VPS_HOST
VPS_USER
VPS_SSH_KEY
VPS_APP_DIR
```

Production app secrets should stay on the VPS in `.env.production` or `CLIENTS/.env.<client>`.

Do not commit production env files.

## Manual Redeploy

For manual deployment:

```bash
ssh deploy@your-vps-ip
cd /opt/dashpoint
git pull origin main
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml up -d --build
docker image prune -f
```

For a specific client:

```bash
docker compose --env-file CLIENTS/.env.xenon -p xenon -f docker-compose.vps.yml up -d --build
```

## Rollback

If a deploy breaks:

```bash
cd /opt/dashpoint
git log --oneline -5
git checkout <known-good-commit>
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml up -d --build
```

For CI/CD image-based deployment, roll back by redeploying the previous image tag.

## Backups

Database backups are mandatory.

Example backup:

```bash
docker compose --env-file .env.production -p dashpoint-demo -f docker-compose.vps.yml exec db \
  pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > "backup-$(date +%F).sql"
```

Also back up uploads:

```bash
tar -czf "uploads-$(date +%F).tar.gz" ./data/uploads
```

For named volumes, use a temporary container or provider-level snapshot.

Recommended backup cadence:

- daily DB backup
- daily upload backup
- keep at least 7 daily backups
- test restore regularly

## Restore Test

Do not trust backups until a restore has been tested.

Basic restore flow:

```bash
createdb restored_dashpoint
psql restored_dashpoint < backup-YYYY-MM-DD.sql
```

For Dockerized Postgres, restore inside the DB container or through a temporary Postgres client container.

## Security Checklist

Before going live:

- Use HTTPS domain, not raw IP.
- Use strong unique `JWT_SECRET` per client.
- Use strong unique DB password per client.
- Set `NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS=false`.
- Do not expose Postgres port publicly.
- Do not expose backend/frontend host ports unless needed.
- Keep `.env.production` and `CLIENTS/.env.*` at `chmod 600`.
- Ensure `CORS_ORIGINS` is exact, not `*`.
- Ensure production frontend uses `/api/v1` or the correct HTTPS API URL.
- Ensure Caddy routes `/api/v1/*` before the frontend catch-all route.
- Keep Docker, OS packages, and Caddy updated.
- Configure VPS firewall.
- Configure backups.

## Common Failure Cases

### Frontend Loads But API Fails

Check:

```bash
curl -i https://pos.example.com/api/v1/ping
```

Likely causes:

- Caddy route order is wrong.
- `NEXT_PUBLIC_API_URL` is wrong.
- Backend container is not healthy.
- Backend and Caddy do not share a Docker network.

### CORS Error In Browser

With same-origin `/api/v1`, CORS should usually disappear.

If using a separate API domain, check:

```env
CORS_ORIGINS=https://frontend-domain.example.com
```

Do not use `*` in production.

### Backend Fails On Database SSL

If DB is private Docker-only:

```env
DATABASE_URL=postgres://user:pass@db:5432/dbname?sslmode=disable
DATABASE_ALLOW_LOCAL_INSECURE=true
```

If DB is remote:

```env
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
```

### Caddy Cannot Get Certificate

Check:

- DNS A record points to VPS IP.
- Ports 80 and 443 are open.
- No other service is occupying 80/443.
- Caddy logs:

```bash
docker compose -f docker-compose.vps.yml logs --tail=100 caddy
```

### Client Data Appears Mixed

Stop immediately.

Check:

- Compose project names.
- DB volume names.
- upload volume names.
- env file used for deployment.
- Caddy routes.

Each client must have isolated DB and upload storage.

## Final Go-Live Checklist

```text
[ ] Domain DNS points to VPS
[ ] Caddy serves HTTPS
[ ] Frontend loads through domain
[ ] Backend ping works through /api/v1
[ ] Login works
[ ] POS checkout works
[ ] Upload/image display works
[ ] DB is not public
[ ] Env files are chmod 600
[ ] Backups configured
[ ] Rollback command tested
[ ] GitHub Actions or manual deploy process documented
```

