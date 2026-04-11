package storage

import (
	"testing"
	"time"
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
