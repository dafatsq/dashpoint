#!/usr/bin/env bash

set -euo pipefail

app_dir=${APP_DIR:-$(pwd)}
client_env_dir=${CLIENT_ENV_DIR:-CLIENTS}
caddy_file=${CADDY_FILE:-$app_dir/Caddyfile}
client_env_filter=${CLIENT_ENV_FILTER:-}
compose_file=${COMPOSE_FILE:-docker-compose.prod.yml}
caddy_compose_file=${CADDY_COMPOSE_FILE:-docker-compose.caddy.yml}

resolve_path() {
  local path=$1
  if [[ "$path" = /* ]]; then
    printf '%s' "$path"
  else
    printf '%s/%s' "$app_dir" "$path"
  fi
}

client_env_dir=$(resolve_path "$client_env_dir")
compose_file=$(resolve_path "$compose_file")
caddy_compose_file=$(resolve_path "$caddy_compose_file")

if [[ ! -d "$client_env_dir" ]]; then
  printf 'Client env directory does not exist: %s\n' "$client_env_dir" >&2
  exit 1
fi

if [[ ! -x "$app_dir/scripts/render-caddyfile.sh" ]]; then
  printf 'Caddy renderer is missing or not executable: %s\n' "$app_dir/scripts/render-caddyfile.sh" >&2
  exit 1
fi

read_env_value() {
  local key=$1
  local file=$2
  local value

  value=$(awk -F= -v key="$key" '$1 == key {sub(/^[^=]*=/, ""); print; exit}' "$file")
  value=${value#\"}
  value=${value%\"}
  value=${value#\'}
  value=${value%\'}
  printf '%s' "$value"
}

validate_project_name() {
  local project_name=$1
  if [[ -z "$project_name" || ! "$project_name" =~ ^[a-z0-9][a-z0-9_-]*$ ]]; then
    printf 'Invalid PROJECT_NAME in client env.\n' >&2
    exit 1
  fi
}

shopt -s nullglob
env_files=("$client_env_dir"/.env.*)
selected_count=0
deployed_projects=()
deployed_sites=()

for env_file in "${env_files[@]}"; do
  [[ "$(basename "$env_file")" == ".env.example" ]] && continue

  if [[ -n "$client_env_filter" ]]; then
    filter_name=$(basename "$client_env_filter")
    [[ "$(basename "$env_file")" == "$filter_name" ]] || continue
  fi

  project_name=$(read_env_value PROJECT_NAME "$env_file")
  site_address=$(read_env_value CADDY_SITE_ADDRESS "$env_file")
  validate_project_name "$project_name"

  if [[ -z "$site_address" || ! "$site_address" =~ ^[A-Za-z0-9*._-]+$ ]]; then
    printf 'Invalid CADDY_SITE_ADDRESS in client env: %s\n' "$env_file" >&2
    exit 1
  fi

  docker compose \
    --env-file "$env_file" \
    -p "$project_name" \
    -f "$compose_file" \
    config --quiet

  # DEPLOY_TAG is set by CI when images were prebuilt and loaded onto this
  # host; plain `up -d` then uses those images. Without it, fall back to
  # building locally (not recommended on small VPS instances).
  up_flags="-d"
  if [[ -z "${DEPLOY_TAG:-}" ]]; then
    up_flags="$up_flags --build"
  fi

  docker compose \
    --env-file "$env_file" \
    -p "$project_name" \
    -f "$compose_file" \
    up $up_flags

  selected_count=$((selected_count + 1))
  deployed_projects+=("$project_name")
  deployed_sites+=("$site_address")
  printf 'Deployed client: %s\n' "$project_name"
done

if (( selected_count == 0 )); then
  if [[ -n "$client_env_filter" ]]; then
    printf 'No matching client env file found: %s\n' "$client_env_filter" >&2
  else
    printf 'No active client env files found in: %s\n' "$client_env_dir" >&2
  fi
  exit 1
fi

# Always render every active route, even when only one client was rebuilt.
"$app_dir/scripts/render-caddyfile.sh" "$client_env_dir" "$caddy_file"

if docker ps -aq --filter 'name=^global-caddy$' | grep -q .; then
  if ! docker ps -q --filter 'name=^global-caddy$' | grep -q .; then
    docker start global-caddy >/dev/null
  fi
else
  docker compose -f "$caddy_compose_file" up -d
fi

docker exec global-caddy caddy validate --config /etc/caddy/Caddyfile >/dev/null
docker exec global-caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null
printf 'Caddy routes validated and reloaded.\n'

if ! command -v curl >/dev/null 2>&1; then
  printf 'curl is required for deployment health checks.\n' >&2
  exit 1
fi

for index in "${!deployed_projects[@]}"; do
  project_name=${deployed_projects[$index]}
  site_address=${deployed_sites[$index]}
  backend_container="${project_name}-backend-prod"

  if [[ "$site_address" == *'*'* ]]; then
    backend_state=$(docker inspect -f '{{.State.Status}}' "$backend_container" 2>/dev/null || true)
    if [[ "$backend_state" != "running" ]]; then
      printf 'Backend health check failed for %s: container state is %s\n' "$project_name" "${backend_state:-missing}" >&2
      docker logs --tail=80 "$backend_container" >&2 || true
      exit 1
    fi
    continue
  fi

  health_url="https://${site_address}/api/v1/health"
  if ! curl --fail --silent --show-error --max-time 20 --retry 10 --retry-delay 2 \
    --resolve "${site_address}:443:127.0.0.1" "$health_url" >/dev/null; then
    printf 'Backend health check failed for %s (%s).\n' "$project_name" "$health_url" >&2
    docker logs --tail=80 "$backend_container" >&2 || true
    exit 1
  fi

  printf 'Health check passed: %s\n' "$health_url"
done
