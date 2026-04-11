package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("GITSTREAM_INTEGRATION") != "1" {
		t.Skip("set GITSTREAM_INTEGRATION=1 to run storage integration tests")
	}
}

func integrationPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     envOrDefault("POSTGRES_HOST", "localhost"),
		Port:     envOrDefault("POSTGRES_PORT", "5432"),
		Database: envOrDefault("POSTGRES_DB", "gitstream"),
		User:     envOrDefault("POSTGRES_USER", "gitstream"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
	}
}

func integrationClickHouseConfig() ClickHouseConfig {
	return ClickHouseConfig{
		Host:     envOrDefault("CLICKHOUSE_HOST", "localhost"),
		Port:     envOrDefault("CLICKHOUSE_NATIVE_PORT", "9000"),
		Database: envOrDefault("CLICKHOUSE_DB", "gitstream"),
		User:     envOrDefault("CLICKHOUSE_USER", "gitstream"),
		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func integrationEvent(id, eventType, repo string) events.GitHubEvent {
	return events.GitHubEvent{
		ID:        id,
		Type:      eventType,
		RepoName:  repo,
		ActorName: "integration",
		CreatedAt: time.Date(2026, 5, 8, 16, 30, 0, 0, time.UTC),
		Payload: []byte(`{
			"id":"` + id + `",
			"type":"` + eventType + `",
			"repo":{"name":"` + repo + `"},
			"actor":{"login":"integration"},
			"created_at":"2026-05-08T16:30:00Z"
		}`),
	}
}

func integrationContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}
