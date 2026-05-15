# Interview Checklist

Use this checklist to practice explaining GitStream clearly under interview
pressure.

## One-Minute System Pitch

GitStream is a real-time GitHub events pipeline written in Go. It polls GitHub
Public Events, publishes them to Kafka, processes them with a bounded worker
pool, stores raw events in Postgres, stores analytics rows in ClickHouse, and
serves a dashboard/API with Prometheus and Grafana observability.

The main backend topics are Kafka processing, at-least-once delivery,
idempotency, DLQs, worker pools, backpressure, storage tradeoffs, analytics
queries, metrics, and deployment.

## Architecture Questions

Why Kafka instead of writing directly to the database?

Kafka decouples event acceptance from database write speed. Ingest can publish
events even if processing is temporarily slower. It also lets the processor
scale independently and gives a replayable boundary for failures.

Why Postgres and ClickHouse?

Postgres is the correctness store for raw events. It has a primary key on the
GitHub event ID and protects against duplicate raw rows. ClickHouse is the
analytics store for fast time-window and aggregation queries.

Why use repository name as the Kafka key?

It keeps events for the same repository ordered within a partition while still
allowing parallel processing across repositories.

Why manual offset commits?

The processor commits only after Postgres and ClickHouse succeed, or after an
exhausted failure is published to the DLQ. That prevents acknowledging work
before it is durably handled.

What delivery guarantee does the processor provide?

At-least-once. Kafka messages can be redelivered after a crash, so sinks must
handle duplicates. Postgres does this with `ON CONFLICT (id) DO NOTHING`.

## Failure Handling Questions

What happens if Postgres succeeds and ClickHouse fails?

The message is not committed yet. The processor retries the event. On retry,
Postgres sees the same event ID and does nothing, while ClickHouse gets another
chance to write the analytics row.

What happens after retries are exhausted?

The processor publishes the original failed event to `github-events-dlq`. Only
after that DLQ publish succeeds does it commit the original Kafka offset.

Why have a DLQ?

A permanently bad event should not block a partition forever. The DLQ preserves
the failed payload for inspection while allowing the main consumer group to
continue.

What is consumer lag?

Consumer lag is how far the consumer is behind the broker's latest offsets. In
GitStream, `gitstream_consumer_lag` currently measures messages fetched by the
processor but not yet durably completed. For production, broker-side group lag
from Kafka or Confluent should be added.

## Worker Pool And Backpressure

Why a bounded worker pool?

It limits concurrency so a traffic spike does not create unbounded goroutines
or memory growth. The queue and worker count form the backpressure boundary.

Can you set `WORKER_COUNT=1000`?

You can, but it is not automatically better. More workers can overwhelm
Postgres, ClickHouse, Kafka, or network connections. Worker count should be
raised with metrics: lag, processing latency, DB write latency, failures, and
CPU/memory.

What made local processing slower with 10 workers?

Each worker waits for durable storage handling, and ClickHouse batches can
flush on the time interval instead of the size threshold. With too few workers,
the processor drains Kafka more slowly even if the publisher is fast.

## Observability Questions

Which metrics matter most?

- `gitstream_events_ingested_total`
- `gitstream_events_processed_total`
- `gitstream_events_failed_total`
- `gitstream_consumer_lag`
- `gitstream_dlq_depth`
- `gitstream_processing_duration_seconds`
- `gitstream_postgres_write_duration_seconds`
- `gitstream_clickhouse_write_duration_seconds`
- `gitstream_api_request_duration_seconds`

What would you alert on?

Alert on sustained broker-side consumer lag growth, DLQ depth growth,
processor failure rate, database write latency, ingest publish failures, API
5xx rate, and API latency.

Why is Prometheus private on Fly?

Prometheus exposes internal service topology and metrics. In the demo, Grafana
is the public authenticated entry point and Prometheus stays reachable only over
Fly private networking.

## Deployment Questions

What was deployed?

The temporary Fly.io demo deployed ingest, processor, API, Postgres,
ClickHouse, optional Prometheus, and optional Grafana. Kafka was external
through Confluent Cloud.

What deployment issue did you debug?

The services were up, but real events were not flowing through Kafka correctly.
The fix was deriving `tls.Config.ServerName` from the Confluent broker hostname
for kafka-go TLS connections, then sharing that transport setup across producer
and consumer code.

Why tear down Fly apps?

The Fly setup is a temporary portfolio demo. Postgres, ClickHouse, Grafana, and
apps can create ongoing cost, so the project includes teardown automation.

What secret hygiene matters after a demo?

Rotate any exposed Kafka, database, Fly, Grafana, or GitHub credentials. Public
repos should contain only templates and env var names, never live secret
values.

## Benchmark Questions

What numbers can you defend?

- Local worker benchmark: 1,000 events, 944.9 published events/sec, 100
  workers, 0 final lag, 0 failures, 0 DLQ depth.
- ClickHouse read benchmark: 1M rows queried by `/api/trending` in 0.037935s.
- Fly smoke/load test: 100k synthetic events published to Confluent at about
  570 events/sec, processor at 100 workers, 0 failed events, 0 DLQ depth.

Why call the Fly test a smoke/load test instead of a production benchmark?

It proved the deployed path could accept and process a large burst, but it was
not a controlled benchmark with isolated bottleneck analysis, long duration,
autoscaling, or production-grade managed storage.

## Production Readiness Questions

What would you add next?

- GitHub Actions CI for tests and builds.
- Main-branch deploy workflow after tests pass.
- Managed Postgres and ClickHouse.
- Broker-side Kafka lag metrics.
- Alerting.
- API authentication and rate limits.
- Longer soak tests.
- Autoscaling based on lag and write latency.

What is the strongest part of the project?

The processor path is the strongest part: manual commits after durable writes,
idempotent raw storage, retries, DLQ behavior, bounded concurrency, and metrics.

What is the weakest production gap?

CI/CD and managed infrastructure are still on hold. The demo deployment proves
the system runs outside local Docker, but production would need automated
deploys, stronger secret rotation, managed data services, and alerting.
