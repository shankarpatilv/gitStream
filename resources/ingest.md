# Ingest Service

This guide runs the ingest flow locally.

The ingest service:

```text
GitHub Public Events API
  -> parse events
  -> keep supported event types
  -> dedupe recent event IDs
  -> publish accepted events to Kafka topic github-events
```

## Requirements

- Go 1.25+
- Docker Desktop
- Local Kafka from `docker-compose.yml`

## Environment

Create a local env file if needed:

```sh
cp .env.example .env
```

Useful defaults:

```env
INGEST_PORT=8080
GITHUB_TOKEN=
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=github-events
KAFKA_USERNAME=
KAFKA_PASSWORD=
```

`GITHUB_TOKEN` is optional for local testing. Without it, the service polls the
public GitHub Events API anonymously with lower rate limits.

## Start Kafka

```sh
docker compose up -d kafka
```

Check Kafka health:

```sh
docker compose ps kafka
```

Expected status:

```text
Up ... (healthy)
```

## Run Ingest

```sh
INGEST_PORT=18080 \
KAFKA_BROKERS=localhost:9092 \
KAFKA_TOPIC=github-events \
go run ./cmd/ingest
```

Expected logs:

```text
INFO starting service service=ingest port=18080
INFO accepted github event repo=... type=PushEvent actor=...
```

Stop the service with `ctrl+c`.

Expected shutdown logs:

```text
INFO shutdown signal received
INFO github poller stopped error="context canceled"
INFO server stopped
INFO kafka producer closed
INFO service stopped service=ingest
```

## Check Health And Metrics

Run these while ingest is running:

```sh
curl -i localhost:18080/health
```

Expected:

```text
HTTP/1.1 200 OK
{"status":"ok"}
```

Check Prometheus metrics:

```sh
curl -i localhost:18080/metrics
```

Expected:

```text
HTTP/1.1 200 OK
```

with Prometheus text-format metrics in the response body.

## Check Kafka Messages

List topics:

```sh
docker exec gitstream-kafka \
  /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --list
```

Expected topic:

```text
github-events
```

Read one message:

```sh
docker exec gitstream-kafka \
  /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic github-events \
  --from-beginning \
  --max-messages 1 \
  --timeout-ms 5000
```

Expected output is one raw GitHub event JSON payload.

Check offsets:

```sh
docker exec gitstream-kafka \
  /opt/kafka/bin/kafka-get-offsets.sh \
  --bootstrap-server localhost:9092 \
  --topic github-events
```

Example:

```text
github-events:0:34
```

That means partition `0` has messages up to offset `34`.

## Run Tests

```sh
go test ./...
```

Check `internal/events` coverage:

```sh
go test ./internal/events -cover
```

Expected coverage is at least 70%.

If Go cache permissions are blocked in Codex:

```sh
GOCACHE=/tmp/gitstream-go-cache go test ./...
GOCACHE=/tmp/gitstream-go-cache go test ./internal/events -cover
```
