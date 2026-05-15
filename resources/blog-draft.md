# Building GitStream: A Small Real-Time Pipeline in Go

I built GitStream because I wanted a backend project that went beyond normal
CRUD work. The goal was simple: take public GitHub activity, move it through a
real event pipeline, and show the result in a small dashboard.

The flow looks like this:

```text
GitHub Public Events API
  -> Go ingest service
  -> Kafka
  -> Go processor worker pool
  -> Postgres and ClickHouse
  -> API/dashboard
  -> Prometheus and Grafana
```

The product is intentionally small. It shows trending repositories and recent
events. The real work is in the backend path: ingestion, Kafka, durable writes,
analytics queries, metrics, and deployment.

## Why I Built It This Way

Kafka sits between ingest and processing so the system does not depend on the
databases being ready at the exact moment an event arrives. If processing slows
down, Kafka gives the processor room to catch up.

Postgres stores the raw events. It is the correctness layer. Each GitHub event
ID is unique, so Postgres can ignore duplicates if Kafka redelivers a message.

ClickHouse stores the analytics rows. It is better suited for fast trending and
time-window queries than asking Postgres to do every aggregation.

## The Most Important Part

The processor is the core of the project. It consumes Kafka messages, sends
them through a bounded worker pool, writes to Postgres, writes to ClickHouse,
and only then commits the Kafka offset.

That gives the system at-least-once processing. If the processor crashes before
committing, Kafka can redeliver the event. Postgres idempotency protects the raw
event table from duplicate rows.

I also added retries and a DLQ so one bad event does not block the whole
consumer group.

## What I Measured

Local worker benchmark:

```text
1,000 synthetic Kafka events
944.9 published events/sec
100 processor workers
0 final Kafka lag
0 failures
0 DLQ depth
```

ClickHouse/API read benchmark:

```text
1,000,000 seeded ClickHouse rows
/api/trending?hours=24&limit=10
0.037935 seconds
```

Temporary Fly.io load smoke:

```text
100,000 synthetic events
about 570 published events/sec
0 failed events
0 DLQ depth
```

## What I Learned

The biggest lesson was that event pipelines are mostly about boundaries. Kafka
separates accepting work from processing work. Postgres separates correctness
from analytics. ClickHouse keeps read-heavy aggregation queries fast.
Prometheus and Grafana make the system easier to explain because the behavior
is visible instead of guessed.

For production, I would add CI/CD, managed databases, broker-side Kafka lag
metrics, alerts, authentication, and longer soak tests. For a portfolio
project, GitStream gives me a concrete system to discuss: Kafka keys, manual
commits, idempotency, DLQs, worker pools, backpressure, analytics schema design,
metrics, and deployment tradeoffs.
