package storage

import (
	"testing"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

func TestClickHouseStoreIntegrationBatchInsertAndQuery(t *testing.T) {
	requireIntegration(t)
	ctx := integrationContext(t)

	store, err := NewClickHouseStore(ctx, integrationClickHouseConfig())
	if err != nil {
		t.Fatalf("NewClickHouseStore returned error: %v", err)
	}
	defer store.Close()
	if err := store.EnsureSchema(ctx); err != nil {
		t.Fatalf("EnsureSchema returned error: %v", err)
	}

	suffix := time.Now().UTC().Format("20060102150405.000000000")
	repo := "integration/clickhouse-" + suffix
	batch := []events.GitHubEvent{
		integrationEvent("integration-clickhouse-1-"+suffix, "PushEvent", repo),
		integrationEvent("integration-clickhouse-2-"+suffix, "IssuesEvent", repo),
	}
	if err := store.InsertAnalyticsBatch(ctx, batch); err != nil {
		t.Fatalf("InsertAnalyticsBatch returned error: %v", err)
	}

	var hourlyTotal uint64
	err = store.conn.QueryRow(
		ctx,
		`SELECT sum(count) FROM events_hourly WHERE repo_name = ?`,
		repo,
	).Scan(&hourlyTotal)
	if err != nil {
		t.Fatalf("query events_hourly returned error: %v", err)
	}
	if hourlyTotal != 2 {
		t.Fatalf("hourly total = %d, want 2", hourlyTotal)
	}

	var timeseriesRows uint64
	err = store.conn.QueryRow(
		ctx,
		`SELECT count() FROM events_timeseries WHERE repo_name = ?`,
		repo,
	).Scan(&timeseriesRows)
	if err != nil {
		t.Fatalf("query events_timeseries returned error: %v", err)
	}
	if timeseriesRows != 2 {
		t.Fatalf("timeseries rows = %d, want 2", timeseriesRows)
	}
}
