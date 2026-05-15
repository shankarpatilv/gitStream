# Launch Copy

## LinkedIn

I built GitStream, a real-time GitHub events pipeline in Go.

It ingests public GitHub activity, publishes events to Kafka, processes them
with a bounded worker pool, writes raw events to Postgres, writes analytics
rows to ClickHouse, and exposes a dashboard/API with Prometheus and Grafana.

The point of the project was to practice production backend concerns:

- Kafka producer/consumer design
- manual offset commits after durable writes
- retry and DLQ behavior
- Postgres idempotency
- ClickHouse analytics queries
- Prometheus metrics and Grafana dashboards
- Docker and temporary Fly.io deployment

Measured checks:

- local worker benchmark: 1,000 events at 944.9 published events/sec, 0 final
  lag, 0 failures, 0 DLQ depth
- ClickHouse read benchmark: 1M rows queried by `/api/trending` in 0.037935s
- Fly.io demo: real GitHub events flowed through ingest, Confluent Kafka,
  processor, Postgres, ClickHouse, API, and Grafana
- Fly load smoke: 100k synthetic Kafka events published at about 570 events/sec
  with 0 failed events and 0 DLQ depth

## Show HN

Show HN: GitStream - a Go/Kafka/ClickHouse pipeline for live GitHub events

I built GitStream as a backend systems project. It polls GitHub Public Events,
publishes them to Kafka, processes them with a Go worker pool, stores raw events
in Postgres, stores analytics rows in ClickHouse, and serves a dashboard/API.

It includes Prometheus metrics, a Grafana dashboard, Dockerfiles, and temporary
Fly.io deployment config. The processor uses manual Kafka commits after durable
handling, retries, and a DLQ.

Measured checks include a 1k local worker benchmark, a 1M-row ClickHouse read
benchmark, and a temporary Fly demo using Confluent Kafka.

## GitHub Repo Description

Real-time GitHub events pipeline in Go with Kafka, Postgres, ClickHouse,
Prometheus, Grafana, Docker, and Fly.io demo deployment.

## Short Recruiter Pitch

GitStream is a backend portfolio project that demonstrates streaming ingestion,
Kafka processing, idempotent storage, analytics queries, metrics, dashboards,
and deployment. It is intentionally built with Go and production-style
operational boundaries instead of a simple CRUD stack.

## Resume Bullets

- Built GitStream, a Go event pipeline that ingests GitHub Public Events,
  publishes to Kafka, processes events with a bounded worker pool, and serves
  analytics through an HTTP API and dashboard.
- Implemented at-least-once Kafka processing with manual offset commits,
  Postgres idempotency, retry handling, and a DLQ for exhausted failures.
- Modeled storage with Postgres for raw durable events and ClickHouse for
  analytics tables powering trending repository queries.
- Added Prometheus metrics and a Grafana dashboard covering throughput, lag,
  DLQ depth, failures, retries, database write latency, and API latency.
- Containerized services with production Dockerfiles and deployed a temporary
  Fly.io demo using Confluent Kafka, Postgres, ClickHouse, Prometheus, and
  Grafana.
- Measured 944.9 events/sec in a local Kafka worker benchmark and queried 1M
  ClickHouse analytics rows through the API in 0.037935s.

## Interview Opener

GitStream is a real-time GitHub events pipeline I built in Go. The system polls
GitHub Public Events, publishes events to Kafka, processes them with a bounded
worker pool, writes raw events to Postgres, writes analytics rows to
ClickHouse, and exposes the data through an API, dashboard, Prometheus metrics,
and Grafana.

The core design decision is that the processor commits Kafka offsets only after
durable handling succeeds. That lets me talk concretely about at-least-once
processing, idempotency, DLQs, worker pools, backpressure, analytics schema
design, and production gaps.
