package storage

import (
	"testing"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

func TestPostgresStoreIntegrationInsertAndRead(t *testing.T) {
	requireIntegration(t)
	ctx := integrationContext(t)

	store, err := NewPostgresStore(ctx, integrationPostgresConfig())
	if err != nil {
		t.Fatalf("NewPostgresStore returned error: %v", err)
	}
	defer store.Close()
	if err := store.EnsureSchema(ctx); err != nil {
		t.Fatalf("EnsureSchema returned error: %v", err)
	}

	id := "integration-postgres-" + time.Now().UTC().Format("20060102150405.000000000")
	event := integrationEvent(id, "PushEvent", "integration/postgres")
	if err := store.InsertEvent(ctx, event); err != nil {
		t.Fatalf("first InsertEvent returned error: %v", err)
	}
	if err := store.InsertEvent(ctx, event); err != nil {
		t.Fatalf("second InsertEvent returned error: %v", err)
	}

	var count int
	var eventType string
	err = store.pool.QueryRow(
		ctx,
		`SELECT COUNT(*), max(type) FROM events WHERE id = $1`,
		id,
	).Scan(&count, &eventType)
	if err != nil {
		t.Fatalf("query inserted event returned error: %v", err)
	}
	if count != 1 || eventType != "PushEvent" {
		t.Fatalf("stored event count/type = %d/%q, want 1/PushEvent", count, eventType)
	}
}

func TestPostgresStoreIntegrationRecentEvents(t *testing.T) {
	requireIntegration(t)
	ctx := integrationContext(t)

	store, err := NewPostgresStore(ctx, integrationPostgresConfig())
	if err != nil {
		t.Fatalf("NewPostgresStore returned error: %v", err)
	}
	defer store.Close()
	if err := store.EnsureSchema(ctx); err != nil {
		t.Fatalf("EnsureSchema returned error: %v", err)
	}

	suffix := time.Now().UTC().Format("20060102150405.000000000")
	repo := "integration/recent-" + suffix
	older := integrationEvent("integration-recent-old-"+suffix, "PushEvent", repo)
	newer := integrationEvent("integration-recent-new-"+suffix, "IssuesEvent", repo)
	newer.CreatedAt = older.CreatedAt.Add(10 * time.Minute)
	if err := store.InsertEvent(ctx, older); err != nil {
		t.Fatalf("insert older event: %v", err)
	}
	if err := store.InsertEvent(ctx, newer); err != nil {
		t.Fatalf("insert newer event: %v", err)
	}

	events, err := store.RecentEvents(ctx, repo, 10)
	if err != nil {
		t.Fatalf("RecentEvents returned error: %v", err)
	}
	if len(events) < 2 {
		t.Fatalf("len(events) = %d, want at least 2", len(events))
	}
	if events[0].ID != newer.ID || events[1].ID != older.ID {
		t.Fatalf("events ordered as %q then %q", events[0].ID, events[1].ID)
	}
}

func TestPostgresStoreIntegrationTopContributors(t *testing.T) {
	requireIntegration(t)
	ctx := integrationContext(t)

	store, err := NewPostgresStore(ctx, integrationPostgresConfig())
	if err != nil {
		t.Fatalf("NewPostgresStore returned error: %v", err)
	}
	defer store.Close()
	if err := store.EnsureSchema(ctx); err != nil {
		t.Fatalf("EnsureSchema returned error: %v", err)
	}

	suffix := time.Now().UTC().Format("20060102150405.000000000")
	repo := "integration/contributors-" + suffix
	first := integrationEvent("integration-contributor-1-"+suffix, "PushEvent", repo)
	second := integrationEvent("integration-contributor-2-"+suffix, "IssuesEvent", repo)
	third := integrationEvent("integration-contributor-3-"+suffix, "WatchEvent", repo)
	first.ActorName = "alice"
	second.ActorName = "alice"
	third.ActorName = "bob"
	for _, event := range []events.GitHubEvent{first, second, third} {
		if err := store.InsertEvent(ctx, event); err != nil {
			t.Fatalf("insert event %q: %v", event.ID, err)
		}
	}

	contributors, err := store.TopContributors(ctx, repo, 10)
	if err != nil {
		t.Fatalf("TopContributors returned error: %v", err)
	}
	if len(contributors) < 2 {
		t.Fatalf("len(contributors) = %d, want at least 2", len(contributors))
	}
	if contributors[0].ActorName != "alice" || contributors[0].Count != 2 {
		t.Fatalf("top contributor = %+v, want alice count 2", contributors[0])
	}
	if contributors[1].ActorName != "bob" || contributors[1].Count != 1 {
		t.Fatalf("second contributor = %+v, want bob count 1", contributors[1])
	}
}
