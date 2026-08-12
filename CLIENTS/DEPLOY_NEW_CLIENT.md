# Deploy a New DashPoint Client

DashPoint is a reusable SaaS template. The application source, Docker Compose files, database schema, and Caddy configuration are shared. A client is configured with a private environment file, isolated storage, unique credentials, a domain, and a client-specific desktop executable.

This guide assumes one client instance on a new VPS. The same repository can also host multiple client env files on one VPS.

## What changes for a new client

Create or configure only these client-specific items:

- A private `CLIENTS/.env.<client-slug>` file on the VPS.
- A unique domain or API hostname pointing to the VPS.
- A unique `DATA_DIR`, database name, database credentials, and JWT secret.
- A private Windows executable built with that client's public API URL.
- A GitHub Actions deployment target if this VPS should receive automatic deployments. The current workflow supports one VPS target per run.

Do not edit the shared frontend, backend, database, Docker Compose, or generated Caddy routes for a normal client installation.

## 1. Prepare the VPS

Install Docker Engine, the Docker Compose plugin, OpenSSL, and `curl`. Allow only SSH, HTTP, and HTTPS through the firewall. Do not expose ports 3000, 8080, or 5432 publicly.

Clone the repository into the deployment directory:

```bash
sudo mkdir -p /opt/dashpoint
sudo chown "$USER":"$USER" /opt/dashpoint
git clone <private-repository-url> /opt/dashpoint
cd /opt/dashpoint
```

Create the external network used by the global Caddy container:

```bash
docker network create dashpoint_proxy
```

If the network already exists, keep using it.

## 2. Configure the client environment

Create the private client file on the VPS. Do not commit it to GitHub:

```bash
cd /opt/dashpoint
cp CLIENTS/.env.example CLIENTS/.env.acme
chmod 600 CLIENTS/.env.acme
nano CLIENTS/.env.acme
```

Set values similar to these:

```env
PROJECT_NAME=acme
CADDY_SITE_ADDRESS=pos.acme.example.com
CADDY_API_ONLY=false
CORS_ORIGINS=https://pos.acme.example.com,wails://wails,http://wails.localhost

DATA_DIR=/opt/dashpoint/clients/acme/data

POSTGRES_USER=acme_db_user
POSTGRES_PASSWORD=<unique-long-random-password>
POSTGRES_DB=acme_prod
JWT_SECRET=<unique-random-secret-at-least-32-characters>

JWT_EXPIRY_MINUTES=15
REFRESH_EXPIRY_HOURS=168
NEXT_PUBLIC_API_URL=/api/v1
NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS=false
PROXY_NETWORK=dashpoint_proxy
```

Every client must have a different `PROJECT_NAME`, `DATA_DIR`, database credentials, and `JWT_SECRET`. The client env file is the main per-client configuration file and should remain on the VPS outside version control.

If the client only uses the desktop application, `CADDY_API_ONLY=true` can be used with a dedicated API hostname. The hostname still needs DNS and HTTPS:

```env
CADDY_SITE_ADDRESS=api.acme.example.com
CADDY_API_ONLY=true
CORS_ORIGINS=wails://wails,http://wails.localhost
```

## 3. Configure DNS and client storage

Create an A record for the configured hostname pointing to the VPS. Caddy uses this hostname to obtain and renew the HTTPS certificate.

Create a separate PostgreSQL TLS directory for the client:

```bash
mkdir -p /opt/dashpoint/clients/acme/data/postgres-tls
openssl req -x509 -newkey rsa:2048 -nodes -days 3650 \
  -keyout /opt/dashpoint/clients/acme/data/postgres-tls/server.key \
  -out /opt/dashpoint/clients/acme/data/postgres-tls/server.crt \
  -subj "/CN=acme-postgres"
chmod 600 /opt/dashpoint/clients/acme/data/postgres-tls/server.key
chmod 644 /opt/dashpoint/clients/acme/data/postgres-tls/server.crt
```

Do not reuse another client's data directory or PostgreSQL certificate directory.

## 4. Deploy the client stack

Validate and start only this client:

```bash
cd /opt/dashpoint

docker compose \
  --env-file CLIENTS/.env.acme \
  -p acme \
  -f docker-compose.prod.yml \
  config --quiet

docker compose \
  --env-file CLIENTS/.env.acme \
  -p acme \
  -f docker-compose.prod.yml \
  up -d --build
```

The client containers are named `acme-db-prod`, `acme-backend-prod`, and `acme-frontend-prod`.

Generate and load the shared Caddy routes:

```bash
./scripts/render-caddyfile.sh CLIENTS Caddyfile
docker compose -f docker-compose.caddy.yml up -d
docker exec global-caddy caddy validate --config /etc/caddy/Caddyfile
docker exec global-caddy caddy reload --config /etc/caddy/Caddyfile
```

The deployment script can perform these steps together:

```bash
APP_DIR=/opt/dashpoint \
CLIENT_ENV_DIR=/opt/dashpoint/CLIENTS \
CLIENT_ENV_FILTER=.env.acme \
CADDY_FILE=/opt/dashpoint/Caddyfile \
./scripts/deploy-vps.sh
```

## 5. Verify the installation

```bash
docker ps --filter name=acme
docker logs --tail=100 acme-backend-prod
curl -fsS https://pos.acme.example.com/api/v1/health
```

Then verify browser login, desktop login if applicable, refresh-token behavior, uploads, and one representative POS or inventory flow. Confirm PostgreSQL is not publicly exposed.

## 6. Configure GitHub Actions for this VPS

The current `.github/workflows/deploy.yml` has one deployment target. It receives the target VPS through repository or GitHub Environment secrets:

```text
VPS_HOST
VPS_USER
VPS_SSH_KEY
VPS_KNOWN_HOSTS
VPS_APP_DIR=/opt/dashpoint
VPS_CLIENT_ENV_DIR=/opt/dashpoint/CLIENTS
VPS_CADDY_FILE=/opt/dashpoint/Caddyfile
```

Configure these under the repository's **Settings → Secrets and variables → Actions**. Never place SSH keys, database passwords, JWT secrets, or client env contents in the repository.

When a successful push to `main` changes deployment-relevant paths, GitHub Actions:

1. Runs CI.
2. SSHs into the configured VPS.
3. Copies the exact commit to `VPS_APP_DIR`.
4. Preserves the ignored client env files and client data.
5. Runs `scripts/deploy-vps.sh`.

An empty deployment filter rebuilds every active `CLIENTS/.env.*` file on that VPS. A manual workflow dispatch can target one file, such as `.env.acme`.

For a second client on the same VPS, add another private env file; no workflow change is required. For a client on a separate VPS, use the manual deployment procedure unless you first extend `.github/workflows/deploy.yml` with a deployment matrix. Do not replace the existing `VPS_*` secrets, or future deployments will be redirected away from the existing client.

Desktop-only changes do not deploy the VPS. Shared `frontend/`, `backend/`, migration, Docker, Caddy, or deployment changes are VPS-relevant.

## 7. Build the client's desktop application

On a private Windows build machine, update the repository and build with the client's public HTTPS API URL:

```powershell
git pull
.\desktop\scripts\build-windows.ps1 `
  -ApiBaseUrl "https://pos.acme.example.com/api/v1"
```

The output is always:

```text
build\bin\DashPoint.exe
```

Transfer the executable to the client through a private distribution channel. Do not commit it to GitHub or publish it in a public release. The executable contains the shared frontend and public API URL only; it never contains database credentials, JWT secrets, SSH keys, or production `.env` files.

The production desktop build has demo access disabled by default. Only use `-EnableQuickDemoAccess:$true` for an intentional demo build. Localhost builds require `-AllowLocalApi`.

## 8. Updates and data safety

Shared application updates are made once in the main repository. After a successful merge to `main`, configured VPS targets receive the commit through GitHub Actions. The VPS does not poll GitHub.

The current desktop distribution is private and local: a frontend or backend update does not automatically replace an executable already distributed to a client. Rebuild and privately distribute a new executable when the desktop client needs the update.

Rebuild one client without affecting other client stacks:

```bash
docker compose --env-file CLIENTS/.env.acme -p acme -f docker-compose.prod.yml up -d --build
```

Stop the client while preserving data:

```bash
docker compose --env-file CLIENTS/.env.acme -p acme -f docker-compose.prod.yml down
```

Never use `down --volumes` or delete `DATA_DIR` unless the client's database and uploads are intentionally being destroyed. Back up each client's database and uploads independently.

## Go-live checklist

- [ ] VPS has Docker, Compose, OpenSSL, `curl`, and ports 22/80/443 configured.
- [ ] `dashpoint_proxy` external Docker network exists.
- [ ] `CLIENTS/.env.<client>` exists on the VPS and is mode 600.
- [ ] `PROJECT_NAME`, `DATA_DIR`, database credentials, and `JWT_SECRET` are unique.
- [ ] DNS points the client hostname to the VPS.
- [ ] PostgreSQL TLS files exist under the client's `DATA_DIR`.
- [ ] Caddy routes and HTTPS certificate are working.
- [ ] `/api/v1/health`, login, uploads, and one business flow work.
- [ ] PostgreSQL is not publicly exposed.
- [ ] Database and uploads backups exist.
- [ ] GitHub Actions secrets are configured for the correct VPS, if automatic deployment is required.
- [ ] A private desktop executable was built with the correct client API URL, if required.
