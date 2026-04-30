# Observability

This guide verifies the local Prometheus and Grafana path.

## Requirements

- Docker Desktop
- Local `.env` copied from `.env.example`
- Local services exposing metrics:
  - ingest on `localhost:8080`
  - processor on `localhost:8091`
  - API on `localhost:8090`

## Local Environment

Create a local environment file if needed:

```sh
cp .env.example .env
```

Keep local overrides in `.env`, not in the commands below. On this machine,
`.env` uses `POSTGRES_PORT=15432` because another local Postgres process uses
`5432`.

Expected local values include:

```env
INGEST_PORT=8080
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=github-events
KAFKA_DLQ_TOPIC=github-events-dlq
KAFKA_CONSUMER_GROUP=gitstream-processors
PROCESSOR_METRICS_PORT=8091
POSTGRES_HOST=localhost
POSTGRES_PORT=15432
CLICKHOUSE_HOST=localhost
CLICKHOUSE_NATIVE_PORT=9000
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
```

## Start The Full Local Stack

```sh
docker compose up -d kafka postgres clickhouse prometheus grafana
```

Check container status:

```sh
docker compose ps kafka postgres clickhouse prometheus grafana
```

Expected status:

```text
Up ... (healthy)
```

Create the Kafka topics from `.env` before starting ingest:

```sh
set -a
. ./.env
set +a
docker compose exec -T kafka \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --if-not-exists \
  --topic "$KAFKA_TOPIC" \
  --partitions 3 \
  --replication-factor 1
docker compose exec -T kafka \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --if-not-exists \
  --topic "$KAFKA_DLQ_TOPIC" \
  --partitions 1 \
  --replication-factor 1
```

Run each Go service in a separate terminal from the repo root.

Terminal 1, ingest:

```sh
set -a
. ./.env
set +a
go run ./cmd/ingest
```

Terminal 2, processor:

```sh
set -a
. ./.env
set +a
go run ./cmd/processor
```

Terminal 3, API:

```sh
set -a
. ./.env
set +a
go run ./cmd/api
```

These commands read runtime values from `.env`; do not add local secrets,
ports, Kafka brokers, or database credentials inline.

## Dashboard Links

API dashboard:

```text
http://localhost:8090/dashboard
```

Grafana dashboard:

```text
http://localhost:3000/d/gitstream-pipeline/gitstream-pipeline
```

Prometheus targets:

```text
http://localhost:9090/targets
```

Prometheus graph:

```text
http://localhost:9090/graph
```

Raw service metrics:

```text
http://localhost:8080/metrics
http://localhost:8091/metrics
http://localhost:8090/metrics
```

## Check Prometheus Targets

Open Prometheus:

```text
http://localhost:9090/targets
```

Expected GitStream targets:

```text
gitstream-ingest
gitstream-processor
gitstream-api
```

If a GitStream service is not running locally, its target can be down until the
matching Go service starts.

## Import The Grafana Dashboard

Open Grafana:

```text
http://localhost:3000
```

Use the local admin credentials from `.env`, or the Compose defaults from
`.env.example` when no local override exists.

Add Prometheus as a data source if it does not already exist:

```text
http://prometheus:9090
```

Import the dashboard JSON:

```text
grafana/dashboard.json
```

Expected result: Grafana imports a dashboard named `GitStream Pipeline` with
eight panels:

```text
Event Throughput
Consumer Lag
DLQ Depth
Errors And Retries
Pipeline Latency P95
Storage And API Latency P95
Processed Events Total
Processed Events Last 15m Estimate
```

The panels render live data when the local pipeline is running and Prometheus
has scraped the service `/metrics` endpoints. The total-events panel shows the
exact current processor counter. The 15-minute estimate uses Prometheus range
math, so it can be slightly above or below the exact count around scrape
boundaries.

## Dashboard Metrics

The dashboard is wired to these committed metric names:

```text
gitstream_events_ingested_total
gitstream_events_processed_total
gitstream_events_failed_total
gitstream_consumer_lag
gitstream_dlq_depth
gitstream_ingest_errors_total
gitstream_processor_retries_total
gitstream_processing_duration_seconds
gitstream_kafka_poll_duration_seconds
gitstream_postgres_write_duration_seconds
gitstream_clickhouse_write_duration_seconds
gitstream_api_requests_total
gitstream_api_request_duration_seconds
```

## Validate Dashboard JSON

```sh
jq empty grafana/dashboard.json
```

Expected result: the command exits successfully with no output.

## Stop The Local Stack

Stop the Go services with `ctrl+c` in each service terminal.

Then stop the Docker services:

```sh
docker compose down
```

Expected result: Kafka, Postgres, ClickHouse, Prometheus, and Grafana stop.
Named volumes are preserved, so stored database and Grafana data are not
deleted.
