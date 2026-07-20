#!/usr/bin/env bash

set -euo pipefail

clients_dir=${1:-CLIENTS}
output_file=${2:-Caddyfile}

if [[ ! -d "$clients_dir" ]]; then
  printf 'Client env directory does not exist: %s\n' "$clients_dir" >&2
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

validate_value() {
  local label=$1
  local value=$2
  local pattern=$3

  if [[ -z "$value" || ! "$value" =~ $pattern ]]; then
    printf 'Invalid %s in client env: %s\n' "$label" "$value" >&2
    exit 1
  fi
}

shopt -s nullglob
env_files=("$clients_dir"/.env.*)
if (( ${#env_files[@]} == 0 )); then
  printf 'No client env files found in %s\n' "$clients_dir" >&2
  exit 1
fi

temporary_file=$(mktemp "${TMPDIR:-/tmp}/dashpoint-caddy.XXXXXX")
cleanup() {
  unlink "$temporary_file" 2>/dev/null || true
}
trap cleanup EXIT

for env_file in "${env_files[@]}"; do
  [[ "$env_file" == "$clients_dir/.env.example" ]] && continue

  project_name=$(read_env_value PROJECT_NAME "$env_file")
  site_address=$(read_env_value CADDY_SITE_ADDRESS "$env_file")

  validate_value "PROJECT_NAME" "$project_name" '^[a-z0-9][a-z0-9_-]*$'
  validate_value "CADDY_SITE_ADDRESS" "$site_address" '^[A-Za-z0-9*._-]+$'

  cat >> "$temporary_file" <<EOF
# Generated from $env_file. Do not edit this block manually.
$site_address {
    handle /api/v1/* {
        reverse_proxy ${project_name}-backend-prod:8080
    }

    handle /uploads/* {
        reverse_proxy ${project_name}-backend-prod:8080
    }

    handle {
        reverse_proxy ${project_name}-frontend-prod:3000
    }
}

EOF
done

mv "$temporary_file" "$output_file"
trap - EXIT
