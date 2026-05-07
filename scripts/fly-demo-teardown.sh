#!/usr/bin/env bash
set -euo pipefail

if ! command -v flyctl >/dev/null 2>&1; then
  echo "missing required command: flyctl" >&2
  exit 1
fi

if [[ -z "${GITSTREAM_FLY_PREFIX:-}" ]]; then
  echo "missing required env var: GITSTREAM_FLY_PREFIX" >&2
  exit 1
fi

if [[ "${FLY_DESTROY_CONFIRM:-}" != "destroy" ]]; then
  echo "set FLY_DESTROY_CONFIRM=destroy to remove demo apps and volumes" >&2
  exit 1
fi

apps=(
  "${GITSTREAM_FLY_PREFIX}-grafana"
  "${GITSTREAM_FLY_PREFIX}-prometheus"
  "${GITSTREAM_FLY_PREFIX}-api"
  "${GITSTREAM_FLY_PREFIX}-ingest"
  "${GITSTREAM_FLY_PREFIX}-processor"
  "${GITSTREAM_FLY_PREFIX}-clickhouse"
  "${GITSTREAM_FLY_PREFIX}-postgres"
)

for app in "${apps[@]}"; do
  if flyctl status -a "$app" >/dev/null 2>&1; then
    flyctl apps destroy "$app" --yes
  fi
done
