#!/usr/bin/env bash
set -euo pipefail

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "missing required env var: ${name}" >&2
    exit 1
  fi
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

ensure_app() {
  local app="$1"
  if flyctl status -a "$app" >/dev/null 2>&1; then
    return
  fi
  flyctl apps create "$app" --org "$FLY_ORG"
}

ensure_volume() {
  local app="$1"
  local name="$2"
  local size="$3"
  if flyctl volumes list -a "$app" | grep -q "$name"; then
    return
  fi
  flyctl volumes create "$name" -a "$app" --region "$FLY_REGION" --size "$size" --yes
}

ensure_private_ip() {
  local app="$1"
  if flyctl ips list -a "$app" | grep -q "private ingress"; then
    return
  fi
  flyctl ips allocate-v6 --private -a "$app"
}

import_secrets() {
  local app="$1"
  shift
  printf "%s\n" "$@" | flyctl secrets import -a "$app" --stage
}

require_cmd flyctl
require_env GITSTREAM_FLY_PREFIX
require_env POSTGRES_PASSWORD
require_env CLICKHOUSE_PASSWORD
require_env KAFKA_BROKERS
require_env KAFKA_USERNAME
require_env KAFKA_PASSWORD

FLY_REGION="${FLY_REGION:-iad}"
FLY_ORG="${FLY_ORG:-personal}"

POSTGRES_APP="${GITSTREAM_FLY_PREFIX}-postgres"
CLICKHOUSE_APP="${GITSTREAM_FLY_PREFIX}-clickhouse"
INGEST_APP="${GITSTREAM_FLY_PREFIX}-ingest"
PROCESSOR_APP="${GITSTREAM_FLY_PREFIX}-processor"
API_APP="${GITSTREAM_FLY_PREFIX}-api"

ensure_app "$POSTGRES_APP"
ensure_app "$CLICKHOUSE_APP"
ensure_app "$INGEST_APP"
ensure_app "$PROCESSOR_APP"
ensure_app "$API_APP"

ensure_private_ip "$POSTGRES_APP"
ensure_private_ip "$CLICKHOUSE_APP"

ensure_volume "$POSTGRES_APP" postgres_data 1
ensure_volume "$CLICKHOUSE_APP" clickhouse_data 5

import_secrets "$POSTGRES_APP" \
  "POSTGRES_PASSWORD=${POSTGRES_PASSWORD}"

import_secrets "$CLICKHOUSE_APP" \
  "CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD}"

import_secrets "$INGEST_APP" \
  "KAFKA_BROKERS=${KAFKA_BROKERS}" \
  "KAFKA_USERNAME=${KAFKA_USERNAME}" \
  "KAFKA_PASSWORD=${KAFKA_PASSWORD}" \
  "GITHUB_TOKEN=${GITHUB_TOKEN:-}"

import_secrets "$PROCESSOR_APP" \
  "KAFKA_BROKERS=${KAFKA_BROKERS}" \
  "KAFKA_USERNAME=${KAFKA_USERNAME}" \
  "KAFKA_PASSWORD=${KAFKA_PASSWORD}" \
  "POSTGRES_HOST=${POSTGRES_APP}.internal" \
  "POSTGRES_PASSWORD=${POSTGRES_PASSWORD}" \
  "CLICKHOUSE_HOST=${CLICKHOUSE_APP}.internal" \
  "CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD}"

import_secrets "$API_APP" \
  "POSTGRES_HOST=${POSTGRES_APP}.internal" \
  "POSTGRES_PASSWORD=${POSTGRES_PASSWORD}" \
  "CLICKHOUSE_HOST=${CLICKHOUSE_APP}.internal" \
  "CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD}"

flyctl deploy . --config deploy/fly/postgres.fly.toml -a "$POSTGRES_APP" --remote-only --ha=false
flyctl deploy . --config deploy/fly/clickhouse.fly.toml -a "$CLICKHOUSE_APP" --remote-only --ha=false
flyctl deploy . --config deploy/fly/processor.fly.toml -a "$PROCESSOR_APP" --remote-only --ha=false
flyctl deploy . --config deploy/fly/ingest.fly.toml -a "$INGEST_APP" --remote-only --ha=false
flyctl deploy . --config deploy/fly/api.fly.toml -a "$API_APP" --remote-only --ha=false

echo "API: https://${API_APP}.fly.dev/dashboard"
