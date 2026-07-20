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

for env_file in "${env_files[@]}"; do
  [[ "$(basename "$env_file")" == ".env.example" ]] && continue

  if [[ -n "$client_env_filter" ]]; then
    filter_name=$(basename "$client_env_filter")
    [[ "$(basename "$env_file")" == "$filter_name" ]] || continue
  fi

  project_name=$(read_env_value PROJECT_NAME "$env_file")
  validate_project_name "$project_name"

  docker compose \
    --env-file "$env_file" \
    -p "$project_name" \
    -f "$compose_file" \
    config --quiet

  docker compose \
    --env-file "$env_file" \
    -p "$project_name" \
    -f "$compose_file" \
    up -d --build

  selected_count=$((selected_count + 1))
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
