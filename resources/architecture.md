# GitStream Architecture

GitStream is a real-time backend pipeline for GitHub public activity. It polls
GitHub events, publishes them to Kafka, processes them with a bounded Go worker
pool, writes durable raw events to Postgres, writes analytics rows to
ClickHouse, and exposes a small dashboard/API with Prometheus and Grafana
observability.

## System Flow

```text
GitHub Public Events API
  -> ingest service
  -> Kafka topic github-events
  -> processor consumer group gitstream-processors
  -> bounded worker pool
  -> Postgres raw event store
  -> ClickHouse analytics tables
  -> API service and dashboard
  -> Prometheus metrics
  -> Grafana dashboard
```

## Services

`cmd/ingest`

- Polls the GitHub Public Events API.
- Normalizes each event into the project event contract.
- Publishes events to Kafka with repository name as the key.
- Exposes Prometheus metrics for accepted events, poll duration, and errors.

`cmd/processor`

- Consumes `github-events` as the `gitstream-processors` group.
- Decodes event JSON and sends jobs through a bounded in-process queue.
- Runs a configurable worker pool through `WORKER_COUNT`.
- Writes raw events to Postgres.
- Writes analytics batches to ClickHouse.
- Commits Kafka offsets only after durable handling succeeds.
- Retries transient failures and publishes exhausted failures to
  `github-events-dlq`.

`cmd/api`

- Serves `/health`, `/metrics`, `/dashboard`, and read-only API endpoints.
- Reads recent raw events from Postgres.
- Reads trending/time-series analytics from ClickHouse.

## Event Contract

The pipeline centers on a GitHub event shape with these stable fields:

- `id`: GitHub event ID, used as the Postgres idempotency key.
- `type`: GitHub event type, such as `PushEvent` or `IssuesEvent`.
- `repo.name`: repository name, also used as the Kafka message key.
- `actor.login`: GitHub actor login.
- `created_at`: event timestamp.
- `payload`: preserved raw event payload for debugging and future features.

Kafka preserves all raw event JSON. Postgres stores the raw durable event, and
ClickHouse stores query-optimized analytics rows derived from that event.

## Kafka Design

Primary topic: `github-events`

DLQ topic: `github-events-dlq`

The Kafka key is the repository name. This keeps events for the same repository
ordered within a partition while still allowing parallelism across repositories.

The processor uses manual commits. An offset is committed only after both sinks
are handled successfully, or after the failed event is safely published to the
DLQ. This makes the processor at-least-once. Postgres idempotency with
`ON CONFLICT (id) DO NOTHING` protects the raw event store from duplicate rows
when a message is retried.

In the current metrics implementation, `gitstream_consumer_lag` means fetched
messages not yet durably completed by this processor process. For production
SLOs, broker-side consumer lag from Kafka or Confluent should be added.

## Storage Design

Postgres stores raw events for correctness and inspection:

```text
events(id, type, repo_name, actor_name, created_at, payload)
```

The primary key is `id`. Supporting indexes cover time and repository/time
queries.

ClickHouse stores analytics tables:

```text
events_hourly
events_timeseries
```

`events_hourly` supports fast trending repository queries. `events_timeseries`
supports event trends over time. This split keeps the API simple: Postgres is
the correctness store, ClickHouse is the analytics store.

## API Surface

- `GET /health`: dependency health for Postgres and ClickHouse.
- `GET /metrics`: Prometheus metrics for API requests and latency.
- `GET /dashboard`: embedded dashboard UI.
- `GET /api/trending?hours=24&limit=10`: top repositories by event count.
- `GET /api/events/recent?repo=owner/repo&limit=10`: recent raw events.

## Observability

Each Go service exposes Prometheus metrics. The Grafana dashboard tracks:

- ingested event throughput;
- processed event throughput;
- processor in-flight lag;
- DLQ depth;
- processing failures and retries;
- processor latency;
- Postgres and ClickHouse write latency;
- API request latency.

The temporary Fly.io demo keeps Prometheus private and exposes Grafana as the
public observability entry point. Grafana provisions the Prometheus datasource
and the committed `grafana/dashboard.json` dashboard.

## Deployment Model

Local development uses Docker Compose for Kafka, Postgres, ClickHouse,
Prometheus, and Grafana.

The temporary demo deployment uses Fly.io for:

- ingest;
- processor;
- API;
- Postgres with a 1GB volume;
- ClickHouse with a 5GB volume;
- optional Prometheus;
- optional Grafana with a 1GB volume.

Kafka is external in the Fly demo. The project was tested with Confluent Cloud
over SASL/TLS. A kafka-go TLS issue was fixed by deriving
`tls.Config.ServerName` from the broker hostname instead of leaving it empty.

The Fly setup is intentionally temporary for portfolio demos and screenshots.
The teardown script destroys apps and attached volumes to avoid ongoing cost.

## Measured Evidence

Local worker benchmark on 2026-05-09:

```text
events: 1000
publisher throughput: 944.9 events/sec
processor workers: 100
processor durable count: 1000
final Kafka lag: 0
failures: 0
DLQ depth: 0
Postgres rows: 1000
ClickHouse rows: 1000
processing p95 bucket: <= 100ms
Postgres write p95 bucket: <= 50ms
ClickHouse write p95 bucket: <= 25ms
```

ClickHouse/API read benchmark on 2026-05-09:

```text
ClickHouse rows seeded: 1,000,000
endpoint: /api/trending?hours=24&limit=10
latency: 0.037935 seconds
```

Fly.io demo verification on 2026-05-11:

```text
real GitHub events reached Kafka, processor, Postgres, ClickHouse, and API
Grafana showed ingest, processor, and API targets up
gitstream_events_ingested_total: 51
gitstream_events_processed_total: 57
kafka publish errors: 0
```

Fly synthetic load test on 2026-05-11:

```text
events published: 100,000
publisher throughput: about 570 events/sec
processor workers during test: 100
failed events: 0
DLQ depth: 0
```

## Production Gaps

The project is complete enough to defend as a backend system, but production
operation would need these next steps:

- CI and deploy workflows from GitHub Actions.
- Secret rotation and managed secret ownership after every public demo.
- Managed Postgres and ClickHouse instead of demo self-hosted Fly databases.
- Broker-side Kafka consumer lag metrics from Confluent or Kafka admin APIs.
- Authentication and rate limits for public API endpoints.
- Autoscaling rules based on lag, queue depth, and write latency.
- Longer soak tests with realistic GitHub event distributions.
