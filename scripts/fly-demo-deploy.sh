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
if [[ "${GITSTREAM_FLY_OBSERVABILITY:-}" == "1" ]]; then
  require_env GRAFANA_ADMIN_PASSWORD
fi

FLY_REGION="${FLY_REGION:-iad}"
FLY_ORG="${FLY_ORG:-personal}"

POSTGRES_APP="${GITSTREAM_FLY_PREFIX}-postgres"
CLICKHOUSE_APP="${GITSTREAM_FLY_PREFIX}-clickhouse"
INGEST_APP="${GITSTREAM_FLY_PREFIX}-ingest"
PROCESSOR_APP="${GITSTREAM_FLY_PREFIX}-processor"
API_APP="${GITSTREAM_FLY_PREFIX}-api"
PROMETHEUS_APP="${GITSTREAM_FLY_PREFIX}-prometheus"
GRAFANA_APP="${GITSTREAM_FLY_PREFIX}-grafana"

ensure_app "$POSTGRES_APP"
ensure_app "$CLICKHOUSE_APP"
ensure_app "$INGEST_APP"
ensure_app "$PROCESSOR_APP"
ensure_app "$API_APP"
if [[ "${GITSTREAM_FLY_OBSERVABILITY:-}" == "1" ]]; then
  ensure_app "$PROMETHEUS_APP"
  ensure_app "$GRAFANA_APP"
fi

ensure_private_ip "$POSTGRES_APP"
ensure_private_ip "$CLICKHOUSE_APP"
if [[ "${GITSTREAM_FLY_OBSERVABILITY:-}" == "1" ]]; then
  ensure_private_ip "$PROMETHEUS_APP"
fi

ensure_volume "$POSTGRES_APP" postgres_data 1
ensure_volume "$CLICKHOUSE_APP" clickhouse_data 5
if [[ "${GITSTREAM_FLY_OBSERVABILITY:-}" == "1" ]]; then
  ensure_volume "$GRAFANA_APP" grafana_data 1
fi

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

if [[ "${GITSTREAM_FLY_OBSERVABILITY:-}" == "1" ]]; then
  import_secrets "$PROMETHEUS_APP" \
    "INGEST_METRICS_TARGET=${INGEST_APP}.internal:8080" \
    "PROCESSOR_METRICS_TARGET=${PROCESSOR_APP}.internal:8091" \
    "API_METRICS_TARGET=${API_APP}.internal:8090"

  import_secrets "$GRAFANA_APP" \
    "GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}" \
    "PROMETHEUS_URL=http://${PROMETHEUS_APP}.internal:9090"

  flyctl deploy . --config deploy/fly/prometheus.fly.toml -a "$PROMETHEUS_APP" --remote-only --ha=false
  flyctl deploy . --config deploy/fly/grafana.fly.toml -a "$GRAFANA_APP" --remote-only --ha=false

  echo "Prometheus: https://${PROMETHEUS_APP}.fly.dev"
  echo "Grafana: https://${GRAFANA_APP}.fly.dev"
fi

echo "API: https://${API_APP}.fly.dev/dashboard"
