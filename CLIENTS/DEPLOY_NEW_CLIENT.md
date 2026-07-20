# Deploy A New Client

DashPoint is deployed as one reusable project template. Each client is defined by one private env file; the Compose file and application source are shared.

## 1. Create The Client Env

On the VPS, from the project root:

```bash
cp CLIENTS/.env.example CLIENTS/.env.acme
chmod 600 CLIENTS/.env.acme
```

Edit only `CLIENTS/.env.acme` and set:

- `PROJECT_NAME`: unique lowercase container identity, such as `acme`
- `CADDY_SITE_ADDRESS`: the client's DNS hostname
- `CORS_ORIGINS`: the exact HTTPS origin for that hostname
- `DATA_DIR`: unique absolute storage path for this client's database, TLS files, and uploads
- `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`: unique database credentials
- `JWT_SECRET`: a unique random secret of at least 32 characters

Do not commit the client env file. Do not reuse another client's `DATA_DIR`, database credentials, or JWT secret.

Existing legacy client env files must be migrated to this shape before deployment. In particular, add that client's `CADDY_SITE_ADDRESS`, exact `CORS_ORIGINS`, unique `DATA_DIR`, and the current application settings; do not run the route renderer until those values are present.

## 2. Prepare Client Storage

The production Compose file expects a TLS certificate and key under the client's `DATA_DIR`:

```bash
mkdir -p /opt/dashpoint/clients/acme/data/postgres-tls
openssl req -x509 -newkey rsa:2048 -nodes -days 3650 \
  -keyout /opt/dashpoint/clients/acme/data/postgres-tls/server.key \
  -out /opt/dashpoint/clients/acme/data/postgres-tls/server.crt \
  -subj "/CN=acme-postgres"
chmod 600 /opt/dashpoint/clients/acme/data/postgres-tls/server.key
chmod 644 /opt/dashpoint/clients/acme/data/postgres-tls/server.crt
```

Use a different storage path and certificate per client.

## 3. Deploy The Client Stack

The stack has three client-specific containers: database, backend, and frontend. They use the shared source tree but isolated storage and internal Docker networks.

```bash
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

The client containers will be named `acme-db-prod`, `acme-backend-prod`, and `acme-frontend-prod`.

## 4. Update Shared Caddy

The VPS uses one global Caddy container for all clients. Generate its routes from every client env file; do not edit routes by hand:

```bash
./scripts/render-caddyfile.sh CLIENTS Caddyfile
docker exec global-caddy caddy reload --config /etc/caddy/Caddyfile
```

On the first VPS deployment, start `global-caddy` with `docker compose -f docker-compose.caddy.yml up -d` after generating the first route instead of reloading an existing container.

The client stack and `global-caddy` must share the external `dashpoint_proxy` network. DNS for the client's hostname must point to the VPS before Caddy can issue HTTPS.

## 5. Verify

```bash
docker ps --filter name=acme
docker logs --tail=100 acme-backend-prod
curl -fsS https://acme.example.com/api/v1/health
```

Then verify browser login, API requests, uploads, and one representative business flow.

## 6. Update Or Remove A Client

Rebuild the same client without affecting other clients:

```bash
docker compose --env-file CLIENTS/.env.acme -p acme -f docker-compose.prod.yml up -d --build
```

Stop the client while preserving its database and uploads:

```bash
docker compose --env-file CLIENTS/.env.acme -p acme -f docker-compose.prod.yml down
```

Do not use `down --volumes` unless the client's database and uploads are intentionally being destroyed.
