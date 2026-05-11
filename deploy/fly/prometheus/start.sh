#!/bin/sh
set -eu

sed "s|\${INGEST_METRICS_TARGET}|${INGEST_METRICS_TARGET}|g; \
s|\${PROCESSOR_METRICS_TARGET}|${PROCESSOR_METRICS_TARGET}|g; \
s|\${API_METRICS_TARGET}|${API_METRICS_TARGET}|g" \
  /etc/prometheus/prometheus.yml.template > /tmp/prometheus.yml

exec /bin/prometheus \
  --config.file=/tmp/prometheus.yml \
  --storage.tsdb.path=/prometheus \
  --web.listen-address=:9090
