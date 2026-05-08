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

func TestClickHouseStoreIntegrationTrendingRepos(t *testing.T) {
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
	topRepo := "integration/trending-top-" + suffix
	nextRepo := "integration/trending-next-" + suffix
	batch := []events.GitHubEvent{
		integrationEvent("integration-trending-1-"+suffix, "PushEvent", topRepo),
		integrationEvent("integration-trending-2-"+suffix, "PushEvent", topRepo),
		integrationEvent("integration-trending-3-"+suffix, "IssuesEvent", topRepo),
		integrationEvent("integration-trending-4-"+suffix, "WatchEvent", nextRepo),
	}
	if err := store.InsertAnalyticsBatch(ctx, batch); err != nil {
		t.Fatalf("InsertAnalyticsBatch returned error: %v", err)
	}

	repos, err := store.TrendingRepos(ctx, 24, 1000)
	if err != nil {
		t.Fatalf("TrendingRepos returned error: %v", err)
	}
	topCount, topOK := findTrendingCount(repos, topRepo)
	nextCount, nextOK := findTrendingCount(repos, nextRepo)
	if !topOK || !nextOK {
		t.Fatalf("expected both inserted repos in trending result: %#v", repos)
	}
	if topCount != 3 || nextCount != 1 {
		t.Fatalf("counts = top:%d next:%d, want 3 and 1", topCount, nextCount)
	}
}

func findTrendingCount(repos []TrendingRepo, name string) (uint64, bool) {
	for _, repo := range repos {
		if repo.RepoName == name {
			return repo.Count, true
		}
	}
	return 0, false
}
