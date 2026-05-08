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
  -> publish exhausted failures to github-events-dlq
```

At this stage the processor does not write to ClickHouse or commit offsets.

## Requirements

- Go 1.23+
- Docker Desktop
- Local Kafka from `docker-compose.yml`
- Local Postgres from `docker-compose.yml`

## Start Kafka

```sh
docker compose up -d kafka postgres
docker compose ps kafka postgres
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

## Create Topic

```sh
docker compose exec -T kafka \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --if-not-exists \
  --topic github-events \
  --partitions 3 \
  --replication-factor 1
```

## Produce A Test Message

```sh
docker compose exec kafka \
  /opt/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic github-events \
  --property parse.key=true \
  --property key.separator=:
```

Paste one keyed message, then press `ctrl+c`:

```text
codex/task2:{"id":"task2-1","type":"PushEvent","repo":{"name":"codex/task2"},"actor":{"login":"codex"},"created_at":"2026-05-06T21:30:00Z"}
```

The key before `:` is the Kafka message key. GitStream uses repo name as the
key so events for the same repository stay in the same Kafka partition.

## Run Processor

Use a fresh test consumer group so the run starts from the beginning of the
topic:

```sh
KAFKA_BROKERS=localhost:9092 \
KAFKA_CONSUMER_GROUP=gitstream-task2-check \
go run ./cmd/processor
```

Expected logs include:

```json
{"msg":"consumed kafka message","topic":"github-events","partition":0,"offset":0,"key":"codex/task2"}
```

The exact partition and offset may differ.

Stop the processor with `ctrl+c`.

Expected shutdown logs:

```text
kafka consumer stopped
kafka consumer closed
service stopped
```

## Verify Offsets Are Not Committed

Task 2 intentionally fetches messages without committing offsets. Check the
test consumer group:

```sh
docker compose exec -T kafka \
  /opt/kafka/bin/kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe \
  --group gitstream-task2-check
```

Expected after shutdown:

```text
Consumer group 'gitstream-task2-check' has no active members.
```

There should be no committed offset rows listed for this test group.

## Why This Matters

Fetching a Kafka message means:

```text
read this record into the processor
```

Committing a Kafka offset means:

```text
tell Kafka this record is finished
```

Those are different operations. The processor uses `FetchMessage` so later
tasks can commit only after processing succeeds.

Future flow:

```text
fetch Kafka message
decode GitHub event
write Postgres
write ClickHouse
commit offset only after both writes succeed
```

This is the base for at-least-once processing.

## Run Tests

```sh
go test ./...
```

If Go cache permissions are blocked in Codex:

```sh
GOCACHE=/tmp/gitstream-go-cache go test ./...
```
