# DashPoint VPS Deployment

DashPoint is a reusable SaaS application template. One VPS can run multiple isolated client instances from one checked-out project folder.

## Architecture

    Internet
      -> one global Caddy container on ports 80/443
        -> client frontend container
        -> client backend container
          -> client PostgreSQL container

Each client gets one private env file, one Compose project name, one database and uploads directory, one hostname, and unique credentials. The database is never published to the internet. Client stacks share only the external Docker network used by Caddy.

## Deployment Files

    docker-compose.prod.yml       reusable client stack
    docker-compose.caddy.yml      one shared Caddy container
    Caddyfile                     generated routes or single-client template
    scripts/render-caddyfile.sh   builds routes from client env files
    CLIENTS/.env.example          safe client configuration template
    CLIENTS/DEPLOY_NEW_CLIENT.md  client deployment procedure

Real client env files and production secrets stay on the VPS and are ignored by Git.

For the complete new-client installation procedure, see [CLIENTS/DEPLOY_NEW_CLIENT.md](CLIENTS/DEPLOY_NEW_CLIENT.md). This document describes the shared VPS architecture and operational rules; the client guide covers first-time provisioning, CI/CD targeting, and desktop distribution.

## One-Time VPS Setup

Install Docker and the Compose plugin, configure the firewall for only SSH, HTTP, and HTTPS, then clone the repository once:

    mkdir -p /opt/dashpoint
    cd /opt/dashpoint
    git clone <repository-url> .
    docker network create dashpoint_proxy

If the network already exists, keep using it. Do not expose ports 3000, 8080, or 5432 publicly.

After the first client env file exists, generate its routes and start the shared Caddy service from the project folder:

    cd /opt/dashpoint
    ./scripts/render-caddyfile.sh CLIENTS Caddyfile
    docker compose -f docker-compose.caddy.yml up -d

## Client Configuration

Create one private env file per client:

    cp CLIENTS/.env.example CLIENTS/.env.acme
    chmod 600 CLIENTS/.env.acme

Set unique values for:

    PROJECT_NAME=acme
    CADDY_SITE_ADDRESS=acme.example.com
    CORS_ORIGINS=https://acme.example.com,wails://wails,http://wails.localhost
    DATA_DIR=/opt/dashpoint/clients/acme/data
    POSTGRES_USER=acme_db_user
    POSTGRES_PASSWORD=<long-random-password>
    POSTGRES_DB=acme_prod
    JWT_SECRET=<at-least-32-random-characters>
    NEXT_PUBLIC_API_URL=/api/v1
    NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS=false
    PROXY_NETWORK=dashpoint_proxy

### Desktop API host

For a desktop-only client deployment, use an API hostname instead of a website hostname:

    CADDY_SITE_ADDRESS=api.example.com
    CADDY_API_ONLY=true
    CORS_ORIGINS=wails://wails,http://wails.localhost

This generates HTTPS routes for `/api/v1/*` and `/uploads/*` while returning 404 for the site root. DNS for the API hostname must point to the VPS so Caddy can issue its certificate. The desktop executable still calls the same backend container and PostgreSQL database; it does not create a separate service. Keep HTTPS enabled because the Wails refresh cookie requires `Secure` and `SameSite=None`.

DATA_DIR must be a unique absolute path for every client. It holds that client's PostgreSQL files, PostgreSQL TLS files, and backend uploads.

The root `.env`, `backend/.env`, and `frontend/.env.local` support local or legacy standalone development. Do not configure them for a new multi-client VPS. Use `CLIENTS/.env.<client>` as the production client configuration.

## Client Database TLS Files

The production Compose file expects server.crt and server.key inside DATA_DIR/postgres-tls. For a private same-VPS database, a per-client self-signed certificate is sufficient when the backend uses sslmode=require:

    mkdir -p /opt/dashpoint/clients/acme/data/postgres-tls
    openssl req -x509 -newkey rsa:2048 -nodes -days 3650 \
      -keyout /opt/dashpoint/clients/acme/data/postgres-tls/server.key \
      -out /opt/dashpoint/clients/acme/data/postgres-tls/server.crt \
      -subj "/CN=acme-postgres"
    chmod 600 /opt/dashpoint/clients/acme/data/postgres-tls/server.key
    chmod 644 /opt/dashpoint/clients/acme/data/postgres-tls/server.crt

Use a separate certificate directory for each client.

## Deploy A Client

Validate the selected client env before creating containers:

    docker compose \
      --env-file CLIENTS/.env.acme \
      -p acme \
      -f docker-compose.prod.yml \
      config --quiet

Start or rebuild only that client:

    docker compose \
      --env-file CLIENTS/.env.acme \
      -p acme \
      -f docker-compose.prod.yml \
      up -d --build

The containers are named acme-db-prod, acme-backend-prod, and acme-frontend-prod. The Compose project name and DATA_DIR keep the client isolated.

## Update Shared Caddy Routes

Generate all routes from the client env files after adding or changing a client:

    ./scripts/render-caddyfile.sh CLIENTS Caddyfile
    docker exec global-caddy caddy reload --config /etc/caddy/Caddyfile

The Caddyfile is generated output in shared multi-client mode. Do not manually copy a backend container name from another client. The renderer validates project names and hostnames before replacing the file.

DNS A records for every client hostname must point to the VPS before Caddy can obtain certificates.

## Verify A Client

    docker ps --filter name=acme
    docker logs --tail=100 acme-backend-prod
    curl -fsS https://acme.example.com/api/v1/health

Then verify browser login, API calls, image uploads, and one representative sale or inventory flow. Check that PostgreSQL has no public listener:

    docker ps --format 'table {{.Names}}\t{{.Ports}}'

## Client Data Operations

Rebuild a client without affecting other clients:

    docker compose --env-file CLIENTS/.env.acme -p acme -f docker-compose.prod.yml up -d --build

Stop a client while preserving data:

    docker compose --env-file CLIENTS/.env.acme -p acme -f docker-compose.prod.yml down

Never use down --volumes or delete DATA_DIR unless the client's data is intentionally being destroyed.

## Backups

Back up every client's database and uploads independently. Store backups outside the active DATA_DIR and restrict them to mode 600:

    docker exec acme-db-prod sh -c \
      'pg_dump -Fc --no-owner --no-acl -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
      > /opt/dashpoint/backups/acme-$(date +%F).dump
    chmod 600 /opt/dashpoint/backups/acme-*.dump
    tar -czf /opt/dashpoint/backups/acme-uploads-$(date +%F).tar.gz \
      /opt/dashpoint/clients/acme/data/uploads

Keep multiple restore points and test restoring them before relying on them. Database restoration is a deliberate operation and is not part of normal code deployment.

## Deployment Boundary

Normal code deployment must not replace client databases. It should fetch source code, validate the selected client env file, run backend migrations, rebuild the selected frontend/backend services, run health checks, and reload Caddy after route validation.

Database imports, seed data, and destructive resets require a separate explicit operation.

## CI/CD

`.github/workflows/deploy.yml` runs backend tests, frontend tests, and a production build before deployment. On `main`, a successful run streams the exact pushed commit to the VPS, runs `scripts/deploy-vps.sh`, rebuilds the active client stacks, runs backend migrations during startup, and reloads Caddy after route validation.

The workflow transfers a Git archive over SSH instead of requiring Git credentials on the VPS. It does not overwrite ignored client env files, database directories, uploads, or backups. A manual workflow dispatch can provide `client_env=.env.acme` to rebuild one client; an empty value deploys every active top-level `CLIENTS/.env.*` file.

Configure these GitHub Actions secrets:

    VPS_HOST             VPS address
    VPS_USER             restricted deployment SSH user
    VPS_SSH_KEY          private SSH key for that user
    VPS_KNOWN_HOSTS      pinned known_hosts entry for the VPS
    VPS_APP_DIR          project directory, such as /opt/dashpoint
    VPS_CLIENT_ENV_DIR   usually CLIENTS
    VPS_CADDY_FILE       usually /opt/dashpoint/Caddyfile

The workflow currently has one VPS target per run. A new client on the same VPS only needs another private `CLIENTS/.env.<client>` file. For a client on a separate VPS, use the manual deployment command unless you first extend `.github/workflows/deploy.yml` with a deployment matrix. Do not overwrite the existing target secrets.

Client env files, JWT secrets, database passwords, uploads, and Caddy certificate data stay on the VPS. The deployment user should have only the permissions required to update the project directory and run the approved Docker Compose deployment. Do not recursively change ownership of a client's `DATA_DIR` to the deployment user: PostgreSQL data and TLS files must remain readable by the PostgreSQL container account, and uploads must remain writable by the backend container account.

The deployment script validates the Compose configuration, rebuilds the selected client stack, reloads Caddy, and then checks `https://<client-domain>/api/v1/health` through the local Caddy listener. A deployment fails if the backend is crash-looping or the public route cannot reach a healthy database-backed backend.

For the existing demo during its migration, use `VPS_APP_DIR=/opt/dashpoint-demo` and `VPS_CADDY_FILE=/opt/caddy/Caddyfile`. A fresh deployment should use the single project-folder layout shown above.

## Go-Live Checklist

    [ ] One client env file exists and is chmod 600
    [ ] PROJECT_NAME and DATA_DIR are unique
    [ ] Database credentials and JWT_SECRET are unique
    [ ] Domain DNS points to the VPS
    [ ] Caddy routes were generated and reloaded
    [ ] HTTPS certificate is valid
    [ ] Frontend and /api/v1/health load
    [ ] PostgreSQL is not publicly exposed
    [ ] Login and one business flow work
    [ ] Database and uploads backups exist
    [ ] Restore procedure is documented and tested
    [ ] CI/CD or the manual deployment command is selected
