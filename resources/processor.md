# Processor Service

This guide verifies the Phase 2 processor consumer flow locally.

The current processor flow is:

```text
Kafka topic github-events
  -> processor consumer group gitstream-processors
  -> decode GitHub event
  -> bounded worker pool
  -> retry failures
  -> write raw event to Postgres
  -> write analytics batches to ClickHouse
  -> publish exhausted failures to github-events-dlq
  -> commit Kafka offsets after durable handling
```

The processor commits Kafka offsets only after Postgres and ClickHouse both
succeed, or after an exhausted failure is successfully published to the DLQ.

## Requirements

- Go 1.25+
- Docker Desktop
- Local Kafka from `docker-compose.yml`
- Local Postgres from `docker-compose.yml`
- Local ClickHouse from `docker-compose.yml`

## Start Kafka

```sh
docker compose up -d kafka postgres clickhouse
docker compose ps kafka postgres clickhouse
```

Expected status:

```text
Up ... (healthy)
```

## Verify Postgres

Use Docker Compose environment variables inside the Postgres container. This
avoids putting real local passwords in commands or in this public file.

```sh
docker compose exec postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT current_user, current_database();"'
```

## Check The Events Table

After the processor has started at least once, it should create the `events`
table automatically:

```sh
docker compose exec postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "\dt"'
```

Expected table:

```text
events
```

Check indexes:

```sh
docker compose exec postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "\di events*"'
```

Expected indexes include:

```text
events_pkey
events_created_at_idx
events_repo_created_at_idx
```

## Query Stored Events

Show recent raw events:

```sh
docker compose exec postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT id, type, repo_name, actor_name, created_at FROM events ORDER BY created_at DESC LIMIT 10;"'
```

Check one event by ID:

```sh
docker compose exec postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT id, type, repo_name, actor_name FROM events WHERE id = '\''task8-postgres-1'\'';"'
```

Check duplicate protection:

```sh
docker compose exec postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT id, COUNT(*) FROM events GROUP BY id HAVING COUNT(*) > 1;"'
```

Expected result:

```text
0 rows
```

The processor inserts with `ON CONFLICT (id) DO NOTHING`, so reprocessed Kafka
messages should not create duplicate raw event rows.

## Verify ClickHouse

Use Docker Compose environment variables inside the ClickHouse container. This
keeps real local credentials out of the command and this public file.

```sh
docker compose exec clickhouse sh -c \
  'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$CLICKHOUSE_DB" --query "SELECT currentDatabase()"'
```

After the processor has started at least once, it should create both analytics
tables automatically:

```sh
docker compose exec clickhouse sh -c \
  'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$CLICKHOUSE_DB" --query "SHOW TABLES"'
```

Expected tables:

```text
events_hourly
events_timeseries
```

Query hourly aggregates:

```sh
docker compose exec clickhouse sh -c \
  'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$CLICKHOUSE_DB" --query "SELECT hour, repo_name, event_type, sum(count) AS total FROM events_hourly GROUP BY hour, repo_name, event_type ORDER BY hour DESC LIMIT 10"'
```

Query time-series rows:

```sh
docker compose exec clickhouse sh -c \
  'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$CLICKHOUSE_DB" --query "SELECT timestamp, repo_name, event_type FROM events_timeseries ORDER BY timestamp DESC LIMIT 10"'
```

## End-To-End Offset Commit Check

This check proves the full processor behavior:

```text
Kafka message
  -> Postgres write
  -> ClickHouse write
  -> Kafka offset commit
```

Use a dedicated topic and consumer group so the result is easy to inspect.

## Create Test Topic

```sh
docker compose exec -T kafka \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --if-not-exists \
  --topic github-events-offset-check \
  --partitions 1 \
  --replication-factor 1
```

## Produce Test Messages

```sh
docker compose exec -T kafka sh -c 'printf "%s\n%s\n" \
  "codex/offset:{\"id\":\"task10-offset-1\",\"type\":\"PushEvent\",\"repo\":{\"name\":\"codex/offset\"},\"actor\":{\"login\":\"codex\"},\"created_at\":\"2026-05-08T15:00:00Z\"}" \
  "codex/offset:{\"id\":\"task10-offset-2\",\"type\":\"IssuesEvent\",\"repo\":{\"name\":\"codex/offset\"},\"actor\":{\"login\":\"codex\"},\"created_at\":\"2026-05-08T15:05:00Z\"}" \
  | /opt/kafka/bin/kafka-console-producer.sh \
      --bootstrap-server localhost:9092 \
      --topic github-events-offset-check \
      --property parse.key=true \
      --property key.separator=:'
```

The key before `:` is the Kafka message key. GitStream uses repo name as the
key so events for the same repository stay in the same Kafka partition.

## Run Processor Against Test Topic

Use a fresh test consumer group so this run starts at the beginning of the
temporary topic.

If Compose Postgres is mapped to local port `5432`, omit `POSTGRES_PORT`.

```sh
KAFKA_BROKERS=localhost:9092 \
KAFKA_TOPIC=github-events-offset-check \
KAFKA_CONSUMER_GROUP=gitstream-task10-offset-check \
WORKER_COUNT=2 \
POSTGRES_PORT=15432 \
GOCACHE=/tmp/gitstream-go-cache \
go run ./cmd/processor
```

Expected logs include:

```json
{"msg":"processed kafka message","topic":"github-events-offset-check","partition":0,"offset":0,"key":"codex/offset"}
{"msg":"processed kafka message","topic":"github-events-offset-check","partition":0,"offset":1,"key":"codex/offset"}
{"msg":"committed kafka offset","topic":"github-events-offset-check","partition":0,"offset":0}
{"msg":"committed kafka offset","topic":"github-events-offset-check","partition":0,"offset":1}
```

Stop the processor with `ctrl+c` after both commit logs appear.

Expected shutdown logs:

```text
kafka consumer stopped
kafka consumer closed
service stopped
```

## Verify Postgres Received The Events

```sh
docker compose exec postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT id, type FROM events WHERE repo_name = '\''codex/offset'\'' ORDER BY id;"'
```

Expected rows:

```text
task10-offset-1 | PushEvent
task10-offset-2 | IssuesEvent
```

## Verify ClickHouse Received The Events

```sh
docker compose exec clickhouse sh -c \
  'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$CLICKHOUSE_DB" --query "SELECT event_type, sum(count) FROM events_hourly WHERE repo_name = '\''codex/offset'\'' GROUP BY event_type ORDER BY event_type"'
```

Expected rows:

```text
IssuesEvent  1
PushEvent    1
```

## Verify Offsets Are Committed

Check the test consumer group after the processor handles a valid message:

```sh
docker compose exec -T kafka \
  /opt/kafka/bin/kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe \
  --group gitstream-task10-offset-check
```

Expected after shutdown: the group has no active members, and Kafka lists a
committed offset for the topic partition.

```text
GROUP                         TOPIC                      PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG
gitstream-task10-offset-check github-events-offset-check 0          2               2               0
```

Kafka stores the next offset to read, so processing message offsets `0` and `1`
shows `CURRENT-OFFSET` as `2`.

## Optional: Verify Malformed Messages Go To DLQ

Produce a malformed message:

```sh
docker compose exec -T kafka sh -c 'printf "%s\n" \
  "codex/offset:{\"id\":" \
  | /opt/kafka/bin/kafka-console-producer.sh \
      --bootstrap-server localhost:9092 \
      --topic github-events-offset-check \
      --property parse.key=true \
      --property key.separator=:'
```

Run the processor again with a fresh group:

```sh
KAFKA_BROKERS=localhost:9092 \
KAFKA_TOPIC=github-events-offset-check \
KAFKA_CONSUMER_GROUP=gitstream-task10-malformed-check \
WORKER_COUNT=2 \
POSTGRES_PORT=15432 \
GOCACHE=/tmp/gitstream-go-cache \
go run ./cmd/processor
```

Expected logs include:

```text
skipping malformed kafka message
published malformed message to dlq
committed kafka offset
```

The malformed payload is not written to Postgres or ClickHouse. It is preserved
in `github-events-dlq`, and its source offset is committed only after DLQ
publish succeeds.

## Week 2 Storage Integration Tests

The storage integration tests are skipped by default. Run them only when local
Postgres and ClickHouse are available.

Load ignored local credentials from `.env` and point Postgres at the Compose
port used on this machine:

```sh
set -a
source .env
set +a

GITSTREAM_INTEGRATION=1 \
POSTGRES_PORT=15432 \
GOCACHE=/tmp/gitstream-go-cache \
go test ./internal/storage -run Integration -count=1 -v
```

Expected result:

```text
TestClickHouseStoreIntegrationBatchInsertAndQuery ... PASS
TestPostgresStoreIntegrationInsertAndRead ... PASS
```

If Compose Postgres is mapped to local port `5432`, omit `POSTGRES_PORT`.

## Week 2 Fault Checkpoint

This check records a simple processor interruption/restart boundary after the
full processor path exists.

Create a dedicated topic:

```sh
docker compose exec kafka \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --if-not-exists \
  --topic github-events-week2-fault-check \
  --partitions 1 \
  --replication-factor 1
```

Produce two events:

```sh
docker compose exec -T kafka sh -c 'printf "%s\n%s\n" \
  "codex/fault:{\"id\":\"week2-fault-1\",\"type\":\"PushEvent\",\"repo\":{\"name\":\"codex/fault\"},\"actor\":{\"login\":\"codex\"},\"created_at\":\"2026-05-08T17:00:00Z\"}" \
  "codex/fault:{\"id\":\"week2-fault-2\",\"type\":\"WatchEvent\",\"repo\":{\"name\":\"codex/fault\"},\"actor\":{\"login\":\"codex\"},\"created_at\":\"2026-05-08T17:05:00Z\"}" \
  | /opt/kafka/bin/kafka-console-producer.sh \
      --bootstrap-server localhost:9092 \
      --topic github-events-week2-fault-check \
      --property parse.key=true \
      --property key.separator=:'
```

Run the processor with a fresh consumer group:

```sh
set -a
source .env
set +a

KAFKA_BROKERS=localhost:9092 \
KAFKA_TOPIC=github-events-week2-fault-check \
KAFKA_CONSUMER_GROUP=gitstream-week2-fault-check \
WORKER_COUNT=2 \
POSTGRES_PORT=15432 \
GOCACHE=/tmp/gitstream-go-cache \
go run ./cmd/processor
```

After processing and commit logs appear, interrupt the processor with `ctrl+c`
or stop the process.

Verify Postgres:

```sh
docker compose exec postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT id, type FROM events WHERE repo_name = '\''codex/fault'\'' ORDER BY id;"'
```

Expected rows:

```text
week2-fault-1 | PushEvent
week2-fault-2 | WatchEvent
```

Verify ClickHouse:

```sh
docker compose exec clickhouse sh -c \
  'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$CLICKHOUSE_DB" --query "SELECT event_type, sum(count) FROM events_hourly WHERE repo_name = '\''codex/fault'\'' GROUP BY event_type ORDER BY event_type"'
```

Expected rows:

```text
PushEvent   1
WatchEvent  1
```

Verify the Kafka group has no lag:

```sh
docker compose exec -T kafka \
  /opt/kafka/bin/kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe \
  --group gitstream-week2-fault-check
```

Expected result:

```text
Consumer group 'gitstream-week2-fault-check' has no active members.

GROUP                       TOPIC                           PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG
gitstream-week2-fault-check github-events-week2-fault-check 0          2               2               0
```

The checkpoint passes when both databases contain the expected rows and Kafka
shows `LAG=0`. That means the processor completed durable writes and committed
the source offsets before the interruption boundary.

Restart the processor with the same topic and consumer group:

```sh
set -a
source .env
set +a

KAFKA_BROKERS=localhost:9092 \
KAFKA_TOPIC=github-events-week2-fault-check \
KAFKA_CONSUMER_GROUP=gitstream-week2-fault-check \
WORKER_COUNT=2 \
POSTGRES_PORT=15432 \
GOCACHE=/tmp/gitstream-go-cache \
go run ./cmd/processor
```

Expected restart behavior:

```text
starting service
processor configuration validated
```

There should be no new `processed kafka message` logs for `week2-fault-1` or
`week2-fault-2`, because the committed offset is already at the topic end.
Interrupt the restarted processor after confirming it does not replay those
messages.

## Why This Matters

Fetching a Kafka message means:

```text
read this record into the processor
```

Committing a Kafka offset means:

```text
tell Kafka this record is finished
```

Those are different operations. The processor uses `FetchMessage` so it can
commit only after processing succeeds.

Current flow:

```text
fetch Kafka message
decode GitHub event
write Postgres
write ClickHouse
commit offset only after both writes succeed
```

With parallel workers, offsets can finish out of order. The processor tracks
offset completion per partition and commits only the highest contiguous
completed offset. That avoids committing past an unfinished earlier message.

This is the base for at-least-once processing.

## Run Tests

```sh
go test ./...
```

If Go cache permissions are blocked in Codex:

```sh
GOCACHE=/tmp/gitstream-go-cache go test ./...
```
