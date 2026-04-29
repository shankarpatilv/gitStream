# GitStream

GitStream is a real-time analytics pipeline for public GitHub activity.

It watches GitHub's public event stream, keeps the events that are useful for
open-source activity analysis, sends them through Kafka, processes them in Go,
and stores the results in PostgreSQL and ClickHouse. The API and dashboard then
make it easy to see what is happening across public repositories.

The project is intentionally built around real backend tradeoffs: polling a live
third-party API, handling duplicate events, partitioning Kafka messages,
processing with bounded concurrency, writing to two different databases, and
exposing enough metrics to understand whether the pipeline is healthy.

## What GitStream Answers

- Which repositories are trending right now?
- Who are the top contributors to a repository this week?
- Is a repository mostly getting pushes, pull requests, issues, stars, or forks?
- What public GitHub activity is happening live?
- Is the event pipeline keeping up, or is lag building?

## System Flow

```text
GitHub Public Events API
  -> ingest service
  -> Kafka topic github-events
  -> processor service
  -> PostgreSQL raw event storage
  -> ClickHouse analytics storage
  -> Go API
  -> live dashboard
```

The ingest service polls GitHub every 30 seconds. GitHub may return overlapping
events between polls, so the service deduplicates by event ID before publishing
to Kafka.

Kafka is partitioned by repository name. That keeps events for the same repo
ordered within a partition and gives the processor room to scale across
partitions.

The processor writes each event to PostgreSQL for raw fidelity and to ClickHouse
for fast analytical queries.

## Services

### ingest

Polls the GitHub Public Events API, parses the response, keeps only supported
event types, deduplicates overlapping events, and publishes normalized events to
Kafka.

### processor

Consumes events from Kafka with a bounded worker pool. A message is committed
only after both storage writes succeed. Failed writes are retried with backoff
before the event is sent to a dead-letter topic.

### api

Exposes read-only endpoints for the dashboard. PostgreSQL backs recent-event and
contributor queries. ClickHouse backs trending, breakdown, and time-series
queries.

## Event Scope

GitStream focuses on five event types:

- `PushEvent`
- `PullRequestEvent`
- `IssuesEvent`
- `WatchEvent`
- `ForkEvent`

Other GitHub event types are ignored instead of treated as failures. The goal is
to keep the v1 signal clear and avoid turning every GitHub payload shape into a
separate feature.

## Normalized Event

```go
type GitHubEvent struct {
	ID        string
	Type      string
	RepoName  string
	ActorName string
	CreatedAt time.Time
	Payload   []byte
}
```

The raw payload is preserved so the system can keep full event fidelity while
still exposing a small normalized shape to the rest of the pipeline.

## Kafka Design

| Setting | Value |
| --- | --- |
| Main topic | `github-events` |
| DLQ topic | `github-events-dlq` |
| Partitions | `3` |
| Partition key | `repo_name` |
| Producer acknowledgements | `acks=1` |
| Consumer commits | manual, after successful storage writes |
| Retry policy | `100ms -> 500ms -> 2s -> DLQ` |
| Main retention | 24 hours |
| DLQ retention | 7 days |

This is deliberately small enough to run locally, but it still uses the core
Kafka concepts that matter in production: partition keys, consumer groups,
manual commits, retries, and dead-letter handling.

## Storage Design

PostgreSQL stores raw events:

```sql
CREATE TABLE events (
    id          TEXT PRIMARY KEY,
    event_type  TEXT NOT NULL,
    repo_name   TEXT NOT NULL,
    actor_name  TEXT NOT NULL,
    payload     JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);
```

ClickHouse stores analytics-friendly tables:

```sql
CREATE TABLE events_hourly (
    hour        DateTime,
    repo_name   String,
    event_type  String,
    count       UInt64
) ENGINE = SummingMergeTree()
ORDER BY (hour, repo_name, event_type);
```

PostgreSQL is the right fit for exact raw records and recent lookups.
ClickHouse is the right fit for fast aggregate and time-series queries as event
volume grows.

## API

The API is intentionally read-only in v1. The dashboard is served by the API
service at `/dashboard`, and `/` redirects there.

```text
GET /api/trending?hours=1&limit=10
GET /api/events/recent?repo=X&limit=50
GET /api/stats/breakdown?hours=24
GET /api/contributors/top?repo=X&limit=10
GET /api/stats/pipeline
GET /health
GET /metrics
```

There is no auth, rate limiting, or complex pagination in v1. The project is
focused on the pipeline and analytics path first.

Example local queries:

```sh
curl -i localhost:8090/health
curl -i localhost:8090/metrics
curl -i localhost:8090/dashboard
curl -i 'localhost:8090/api/trending?hours=24&limit=5'
curl -i 'localhost:8090/api/events/recent?repo=owner/repo&limit=10'
curl -i 'localhost:8090/api/stats/breakdown?hours=24'
curl -i 'localhost:8090/api/contributors/top?repo=owner/repo&limit=10'
curl -i 'localhost:8090/api/stats/pipeline'
```

Bad query parameters return `400`, and database-backed endpoints return `503`
when their backing store is unavailable. ClickHouse backs trending and event
breakdown queries; PostgreSQL backs recent events and top contributors.

## Observability

Each service writes structured logs with Go's standard `slog` package and
exposes Prometheus metrics.

Core metrics:

- `gitstream_events_ingested_total`
- `gitstream_events_processed_total`
- `gitstream_events_failed_total`
- `gitstream_consumer_lag`
- `gitstream_processing_duration_seconds`
- `gitstream_postgres_write_duration_seconds`
- `gitstream_clickhouse_write_duration_seconds`
- `gitstream_kafka_poll_duration_seconds`
- `gitstream_dlq_depth`
- `gitstream_active_workers`

The Grafana dashboard focuses on the things that tell you whether the system is
healthy: event throughput, consumer lag, DLQ depth, processing latency, and
database write latency. The importable dashboard JSON lives at
`grafana/dashboard.json`.

## Tech Stack

| Layer | Technology |
| --- | --- |
| Language | Go 1.25+ |
| HTTP router | chi v5 |
| Message queue | Kafka |
| Kafka client | segmentio/kafka-go |
| OLTP database | PostgreSQL 15 |
| OLAP database | ClickHouse 24 |
| Logging | Go standard library `slog` |
| Metrics | Prometheus + Grafana |
| Local development | Docker Compose |
| Deployment target | Fly.io |
| CI/CD | GitHub Actions |

## Repository Layout

```text
cmd/
  ingest/       GitHub poller and Kafka producer
  processor/    Kafka consumer, worker pool, and storage writer
  api/          REST API and dashboard server
internal/
  events/       event models, parsing, filtering, and deduplication
  kafka/        producer and consumer wrappers
  storage/      PostgreSQL and ClickHouse clients
monitoring/
  prometheus.yml
grafana/
  dashboard.json
resources/
  observability.md
docker-compose.yml
.env.example
README.md
```

## Local Development

Create a local environment file:

```sh
cp .env.example .env
```

Run the ingest service:

```sh
go run ./cmd/ingest
```

Validate Go packages:

```sh
go list ./...
```

Validate local infrastructure:

```sh
docker compose config
```

Start local infrastructure:

```sh
docker compose up
```

Docker Compose reads `.env` automatically. The Go services also load `.env` for
local development, while real environment variables still take priority.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `INGEST_PORT` | `8080` | ingest service HTTP port |
| `KAFKA_PORT` | `9092` | local Kafka port |
| `PROCESSOR_METRICS_PORT` | `8091` | processor Prometheus metrics port |
| `POSTGRES_PORT` | `5432` | local PostgreSQL port |
| `CLICKHOUSE_HTTP_PORT` | `8123` | local ClickHouse HTTP port |
| `CLICKHOUSE_NATIVE_PORT` | `9000` | local ClickHouse native port |
| `PROMETHEUS_PORT` | `9090` | local Prometheus port |
| `GRAFANA_PORT` | `3000` | local Grafana port |
| `POSTGRES_DB` | `gitstream` | local PostgreSQL database |
| `POSTGRES_USER` | `gitstream` | local PostgreSQL user |
| `POSTGRES_PASSWORD` | `gitstream` | local PostgreSQL password |
| `CLICKHOUSE_DB` | `gitstream` | local ClickHouse database |
| `CLICKHOUSE_USER` | `gitstream` | local ClickHouse user |
| `CLICKHOUSE_PASSWORD` | `gitstream` | local ClickHouse password |
| `GRAFANA_ADMIN_USER` | `admin` | local Grafana admin user |
| `GRAFANA_ADMIN_PASSWORD` | `admin` | local Grafana admin password |

## Local Infrastructure

| Service | Port |
| --- | --- |
| Kafka | `9092` |
| PostgreSQL | `5432` |
| ClickHouse HTTP | `8123` |
| ClickHouse native | `9000` |
| Prometheus | `9090` |
| Grafana | `3000` |

## Build Plan

The project is built in four phases:

1. Ingest service: GitHub polling, filtering, deduplication, Kafka publishing,
   `/health`, `/metrics`, and graceful shutdown.
2. Processor service: Kafka consumer, bounded worker pool, retries, DLQ, and
   storage writes.
3. API and dashboard: read endpoints backed by PostgreSQL and ClickHouse, plus a
   live dashboard.
4. Operations and launch: Prometheus, Grafana, deployment, CI/CD, load testing,
   and final documentation.

The first priority is a working end-to-end path with real GitHub data. Polish
comes after the pipeline is reliable and measurable.
