# Fly.io Demo Deployment Guide

This guide is for a temporary GitStream demo deployment. Use it to get a live
dashboard URL, take screenshots, then tear the apps down to avoid ongoing cost.

## What Gets Deployed

- `ingest`: polls GitHub Public Events and publishes to external Kafka.
- `processor`: consumes Kafka, writes Postgres and ClickHouse, and commits offsets.
- `api`: serves `/dashboard`, `/health`, `/metrics`, and read-only API endpoints.
- `postgres`: self-hosted Postgres 15 on a 1GB Fly Volume.
- `clickhouse`: self-hosted ClickHouse 24.8 on a 5GB Fly Volume.

Optional observability add-on:

- `prometheus`: scrapes deployed GitStream service metrics over Fly private
  networking.
- `grafana`: serves the provisioned GitStream Pipeline dashboard with a
  Prometheus datasource and a 1GB Fly Volume.

Kafka is not deployed by this repo. The Fly demo expects an external Kafka
broker with SASL/TLS credentials, such as Confluent Cloud.

## Prerequisites

1. A Fly.io account with billing enabled.
2. `flyctl` installed and authenticated.
3. An external Kafka cluster reachable from Fly.io.
4. Kafka topics:
   - `github-events`
   - `github-events-dlq`
5. Optional GitHub token for higher API rate limits.

Install and log in:

```sh
brew install flyctl
flyctl auth login
flyctl auth whoami
```

Fly trial apps may stop Machines after a short runtime window unless billing is
enabled. Add billing before the demo if you need the dashboard to stay live
long enough for screenshots or review.

## Prepare Demo Environment

Copy the template values into a private shell env file and source it:

```sh
cp deploy/fly/demo.env.example /tmp/gitstream-fly-demo.env
$EDITOR /tmp/gitstream-fly-demo.env
source /tmp/gitstream-fly-demo.env
```

Required values:

```sh
export GITSTREAM_FLY_PREFIX="gitstream-demo-yourname"
export FLY_REGION="iad"
export FLY_ORG="personal"
export POSTGRES_PASSWORD="..."
export CLICKHOUSE_PASSWORD="..."
export KAFKA_BROKERS="broker:9092"
export KAFKA_USERNAME="..."
export KAFKA_PASSWORD="..."
export GITHUB_TOKEN="..."
```

Optional observability values:

```sh
export GITSTREAM_FLY_OBSERVABILITY="1"
export GRAFANA_ADMIN_PASSWORD="..."
```

The prefix creates five app names:

```text
<prefix>-postgres
<prefix>-clickhouse
<prefix>-processor
<prefix>-ingest
<prefix>-api
```

With observability enabled, the same prefix also creates:

```text
<prefix>-prometheus
<prefix>-grafana
```

## Deploy

Run the deploy script from the repo root:

```sh
./scripts/fly-demo-deploy.sh
```

The script:

- creates the five Fly apps when missing;
- allocates private IPv6 addresses for Postgres and ClickHouse so `.internal`
  DNS works;
- creates a 1GB Postgres volume and a 5GB ClickHouse volume;
- imports runtime secrets without printing secret values;
- deploys Postgres and ClickHouse first;
- deploys processor, ingest, and API;
- when `GITSTREAM_FLY_OBSERVABILITY=1`, creates Prometheus and Grafana apps,
  creates a 1GB Grafana volume, provisions the datasource and dashboard, and
  deploys both observability apps;
- prints the dashboard URL.

Postgres uses `PGDATA=/var/lib/postgresql/data/pgdata` so the official
Postgres image initializes inside a subdirectory of the Fly volume instead of
the mount root.

## Verify

Check the public API app:

```sh
curl -i "https://${GITSTREAM_FLY_PREFIX}-api.fly.dev/health"
curl -i "https://${GITSTREAM_FLY_PREFIX}-api.fly.dev/dashboard"
curl -i "https://${GITSTREAM_FLY_PREFIX}-api.fly.dev/api/trending?hours=24&limit=5"
```

Check logs:

```sh
flyctl logs -a "${GITSTREAM_FLY_PREFIX}-ingest"
flyctl logs -a "${GITSTREAM_FLY_PREFIX}-processor"
flyctl logs -a "${GITSTREAM_FLY_PREFIX}-api"
```

Expected result:

- API `/health` returns HTTP 200 once both databases are reachable.
- Ingest logs accepted GitHub events and Kafka publish activity.
- Processor logs processed events.
- Dashboard loads at `https://<prefix>-api.fly.dev/dashboard`.

## Verify Observability

When `GITSTREAM_FLY_OBSERVABILITY=1`, Prometheus stays private and Grafana is
the public observability entry point:

```text
https://<prefix>-grafana.fly.dev
```

Log in with user `admin` and the private `GRAFANA_ADMIN_PASSWORD` value from
your local demo env file. The `GitStream / GitStream Pipeline` dashboard is
provisioned automatically from `grafana/dashboard.json`, with the datasource
pointing at `http://<prefix>-prometheus.internal:9090`.

Verify the private Prometheus datasource through Grafana's authenticated API:

```sh
GRAFANA_AUTH="admin:${GRAFANA_ADMIN_PASSWORD}"

curl -u "$GRAFANA_AUTH" \
  "https://${GITSTREAM_FLY_PREFIX}-grafana.fly.dev/api/datasources/proxy/uid/gitstream-prometheus/api/v1/query?query=up"

curl -u "$GRAFANA_AUTH" \
  "https://${GITSTREAM_FLY_PREFIX}-grafana.fly.dev/api/datasources/proxy/uid/gitstream-prometheus/api/v1/query?query=gitstream_events_ingested_total"
```

Prometheus target state should show the ingest, processor, and API jobs with
`up=1` once the services are running and exposing `/metrics`.

If a database Machine is stopped during a trial or after a failed deploy, start
it explicitly:

```sh
flyctl status -a "${GITSTREAM_FLY_PREFIX}-postgres"
flyctl status -a "${GITSTREAM_FLY_PREFIX}-clickhouse"
flyctl status -a "${GITSTREAM_FLY_PREFIX}-processor"
flyctl machines start <machine-id> -a "${GITSTREAM_FLY_PREFIX}-clickhouse"
flyctl machines start <machine-id> -a "${GITSTREAM_FLY_PREFIX}-processor"
```

## Teardown

Destroy the demo apps after screenshots or review:

```sh
export FLY_DESTROY_CONFIRM=destroy
./scripts/fly-demo-teardown.sh
```

Destroying the apps removes their Machines, secrets, public addresses, and
attached volumes. Do not run teardown if you need to keep the demo data.

## Cost Notes

This is a temporary demo setup. It runs five Fly apps plus two persistent
volumes. Expect ongoing cost while the apps and volumes exist. Tear down the
demo after viewing the dashboard unless you intentionally want to keep it live.

Enabling observability adds two more Fly apps and a 1GB Grafana volume.
Prometheus storage is ephemeral in this demo; Grafana uses a volume so local
UI state survives app restarts during the temporary demo. Tear down the whole
demo when screenshots or review are complete.

## Config Files

```text
deploy/fly/api.fly.toml
deploy/fly/ingest.fly.toml
deploy/fly/processor.fly.toml
deploy/fly/postgres.fly.toml
deploy/fly/clickhouse.fly.toml
deploy/fly/prometheus.fly.toml
deploy/fly/grafana.fly.toml
```

The checked-in `app` names are placeholders. The deploy script uses
`flyctl deploy -a` with `GITSTREAM_FLY_PREFIX` so each demo can use unique app
names without editing the TOML files.
