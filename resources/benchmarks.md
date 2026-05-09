# Benchmarks

This guide separates three different checks:

```text
live smoke: GitHub API -> ingest -> Kafka -> processor -> storage -> API
worker load: synthetic Kafka -> processor -> Postgres/ClickHouse -> commits
read benchmark: seeded ClickHouse rows -> API query latency
```

Synthetic benchmark data is for load testing only. It does not replace the live
GitHub ingestion path.

## Requirements

- Docker Desktop
- Local `.env` copied from `.env.example`
- Kafka, Postgres, ClickHouse, and Prometheus from Docker Compose

Load local environment values before running benchmark commands:

```sh
set -a
. ./.env
set +a
```

## Start Dependencies

```sh
docker compose up -d kafka postgres clickhouse prometheus
docker compose ps kafka postgres clickhouse prometheus
```

Expected status:

```text
Up ... (healthy)
```

Create the normal topics from `.env`:

```sh
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

## Live GitHub Smoke Checklist

Run the three services in separate terminals:

```sh
go run ./cmd/ingest
go run ./cmd/processor
go run ./cmd/api
```

Check the live path:

```sh
curl -i "localhost:${API_PORT:-8090}/health"
curl -i "localhost:${API_PORT:-8090}/dashboard"
curl -i "localhost:${API_PORT:-8090}/api/trending?hours=24&limit=10"
```

Expected result:

- ingest logs accepted GitHub events and Kafka publishes succeed.
- processor logs processed events and committed Kafka offsets.
- API health returns `200`.
- dashboard shows trending repositories or clear empty states while data builds.

## Synthetic Worker Load

Use a dedicated benchmark topic and consumer group when you want clean numbers:

```sh
BENCH_TOPIC=github-events-bench
BENCH_GROUP=gitstream-benchmark-$(date +%Y%m%d%H%M%S)
BENCH_PREFIX=bench-worker-$(date +%Y%m%d%H%M%S)

docker compose exec -T kafka \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --if-not-exists \
  --topic "$BENCH_TOPIC" \
  --partitions 3 \
  --replication-factor 1
```

Run the processor against the benchmark topic:

```sh
KAFKA_TOPIC="$BENCH_TOPIC" \
KAFKA_CONSUMER_GROUP="$BENCH_GROUP" \
PROCESSOR_METRICS_PORT=8092 \
WORKER_COUNT=100 \
go run ./cmd/processor
```

Publish synthetic GitHub-like events into Kafka:

```sh
KAFKA_TOPIC="$BENCH_TOPIC" \
go run ./cmd/load-kafka \
  -events=1000 \
  -batch-size=100 \
  -prefix="$BENCH_PREFIX"
```

Confirm durable writes:

```sh
docker compose exec -T postgres sh -c \
  "PGPASSWORD=\"\$POSTGRES_PASSWORD\" psql -U \"\$POSTGRES_USER\" -d \"\$POSTGRES_DB\" -t -c \"SELECT COUNT(*) FROM events WHERE id LIKE '${BENCH_PREFIX}-%';\""

docker compose exec -T clickhouse sh -c \
  "clickhouse-client --user \"\$CLICKHOUSE_USER\" --password \"\$CLICKHOUSE_PASSWORD\" --database \"\$CLICKHOUSE_DB\" --query \"SELECT sum(count) FROM events_hourly WHERE repo_name LIKE '${BENCH_PREFIX}/%'\""
```

Check committed Kafka lag:

```sh
docker compose exec -T kafka \
  /opt/kafka/bin/kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe \
  --group "$BENCH_GROUP"
```

Expected result:

- Postgres count equals published event count.
- ClickHouse count equals published event count.
- Kafka `LAG` is `0` for every benchmark topic partition.
- `gitstream_events_failed_total` is `0`.
- `gitstream_dlq_depth` is `0`.
- Grafana `Processed Events Total` shows the exact processor counter.
- Grafana `Processed Events Last 15m Estimate` shows the recent benchmark
  window using Prometheus range math, so it can differ slightly from the exact
  counter around scrape boundaries.

## Worker Benchmark Result

Measured locally on 2026-05-09:

```text
events: 1000
publisher: 944.9 events/sec
processor workers: 100
processor durable count: 1000
processor first-to-last processed log span: about 0.91 sec
final Kafka lag: 0
failures: 0
DLQ depth: 0
Postgres rows: 1000
ClickHouse rows: 1000
processing p95 bucket: <= 100ms
Postgres write p95 bucket: <= 50ms
ClickHouse write p95 bucket: <= 25ms
```

With `WORKER_COUNT=10`, local processing drains much more slowly because each
worker waits for its ClickHouse batch result and the batch often flushes on the
5-second interval instead of the 100-row size threshold.

## ClickHouse/API Read Benchmark

Seed 1M analytics rows directly into ClickHouse:

```sh
go run ./cmd/seed-clickhouse \
  -rows=1000000 \
  -batch-size=5000 \
  -prefix=bench-read-$(date +%Y%m%d%H%M%S)
```

Run the API:

```sh
go run ./cmd/api
```

Measure trending query latency:

```sh
curl -s -o /tmp/gitstream-trending-bench.json \
  -w '%{time_total}\n' \
  "localhost:${API_PORT:-8090}/api/trending?hours=24&limit=10"
```

Measured locally on 2026-05-09:

```text
ClickHouse rows seeded: 1,000,000
endpoint: /api/trending?hours=24&limit=10
latency: 0.037935 seconds
```

This measures read/query performance only. It does not measure ingestion
capacity because it bypasses Kafka and the processor.

## Dashboard Links

API dashboard:

```text
http://localhost:8090/dashboard
```

Grafana dashboard:

```text
http://localhost:3000/d/gitstream-pipeline/gitstream-pipeline
```
