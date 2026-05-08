# API Service

This guide verifies the Phase 3 query API locally.

The API service:

```text
Postgres raw events
ClickHouse analytics rows
  -> API service on port 8090
  -> health, metrics, and read-only query endpoints
```

## Requirements

- Go 1.25+
- Docker Desktop
- Local Postgres from `docker-compose.yml`
- Local ClickHouse from `docker-compose.yml`

## Start Dependencies

```sh
docker compose up -d postgres clickhouse
docker compose ps postgres clickhouse
```

Expected status:

```text
Up ... (healthy)
```

If Compose Postgres is mapped to local port `15432`, pass
`POSTGRES_PORT=15432` when running the API.

## Run API

```sh
POSTGRES_PORT=15432 go run ./cmd/api
```

Expected startup log:

```json
{"msg":"starting service","service":"api","port":"8090"}
```

Stop the service with `ctrl+c`.

## Health

```sh
curl -i localhost:8090/health
```

Expected with both databases reachable:

```text
HTTP/1.1 200 OK
{"status":"ok","postgres":"ok","clickhouse":"ok"}
```

Expected when either required database is unavailable:

```text
HTTP/1.1 503 Service Unavailable
```

## Metrics

```sh
curl -i localhost:8090/metrics
```

Expected:

```text
HTTP/1.1 200 OK
```

with Prometheus text-format metrics in the response body.

## Trending Repositories

Seed synthetic ClickHouse data intentionally when you want a predictable API
test without waiting for live GitHub traffic:

```sh
go run ./cmd/seed-clickhouse
```

The default inserts `100000` synthetic events into ClickHouse analytics tables.
For a quick smoke test, use fewer rows:

```sh
go run ./cmd/seed-clickhouse -rows=1000 -batch-size=250 -prefix=smoke
```

Query ClickHouse-backed trending repositories:

```sh
curl -i 'localhost:8090/api/trending?hours=24&limit=5'
```

Expected:

```text
HTTP/1.1 200 OK
```

with JSON shaped like:

```json
{
  "hours": 24,
  "limit": 5,
  "repos": [
    {
      "repo_name": "owner/repo",
      "count": 10
    }
  ]
}
```

Bad query params should return `400`:

```sh
curl -i 'localhost:8090/api/trending?hours=bad&limit=5'
```

Expected:

```text
HTTP/1.1 400 Bad Request
{"error":"invalid hours"}
```

Measure endpoint latency:

```sh
curl -s -o /tmp/gitstream-trending.json \
  -w '%{time_total}\n' \
  'localhost:8090/api/trending?hours=24&limit=5'
```

The Phase 3 local check returned:

```text
0.014763
```

## Recent Events

Insert sample raw events for a repo directly into local Postgres:

```sh
docker compose exec -T postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"' <<'SQL'
INSERT INTO events (id, type, repo_name, actor_name, created_at, payload)
VALUES
  ('api-recent-1', 'PushEvent', 'codex/recent', 'codex', now() - interval '5 minutes', '{}'::jsonb),
  ('api-recent-2', 'IssuesEvent', 'codex/recent', 'codex', now(), '{}'::jsonb)
ON CONFLICT (id) DO NOTHING;
SQL
```

Query recent raw events from the API:

```sh
curl -i 'localhost:8090/api/events/recent?repo=codex/recent&limit=5'
```

Expected:

```text
HTTP/1.1 200 OK
```

with newest events first:

```json
{
  "repo": "codex/recent",
  "limit": 5,
  "events": [
    {
      "id": "api-recent-2",
      "type": "IssuesEvent",
      "repo_name": "codex/recent",
      "actor_name": "codex"
    }
  ]
}
```

Missing `repo` or bad `limit` values should return `400`:

```sh
curl -i 'localhost:8090/api/events/recent'
curl -i 'localhost:8090/api/events/recent?repo=codex/recent&limit=0'
```

## Event Breakdown

Query ClickHouse-backed event-type counts:

```sh
curl -i 'localhost:8090/api/stats/breakdown?hours=24'
```

Expected:

```text
HTTP/1.1 200 OK
```

with JSON shaped like:

```json
{
  "hours": 24,
  "breakdown": [
    {
      "event_type": "PushEvent",
      "count": 100
    }
  ]
}
```

Bad `hours` values should return `400`:

```sh
curl -i 'localhost:8090/api/stats/breakdown?hours=0'
curl -i 'localhost:8090/api/stats/breakdown?hours=bad'
```

## Tests

Run the normal suite:

```sh
go test ./...
```

Run the ClickHouse integration test for trending queries:

```sh
set -a; source .env; set +a; \
GITSTREAM_INTEGRATION=1 \
go test ./internal/storage -run ClickHouseStoreIntegrationTrendingRepos -count=1 -v
```

Run the ClickHouse integration test for event breakdown queries:

```sh
set -a; source .env; set +a; \
GITSTREAM_INTEGRATION=1 \
go test ./internal/storage -run ClickHouseStoreIntegrationEventBreakdown -count=1 -v
```

Run the seed command tests:

```sh
go test ./cmd/seed-clickhouse
```
