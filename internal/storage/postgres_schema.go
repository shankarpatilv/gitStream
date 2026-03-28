package storage

import (
	"context"
	"fmt"
)

// EnsureSchema creates the raw events table and query indexes if needed.
func (s *PostgresStore) EnsureSchema(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, createEventsTableSQL); err != nil {
		return fmt.Errorf("create postgres events table: %w", err)
	}
	if _, err := s.pool.Exec(ctx, createEventsCreatedAtIndexSQL); err != nil {
		return fmt.Errorf("create postgres created_at index: %w", err)
	}
	if _, err := s.pool.Exec(ctx, createEventsRepoIndexSQL); err != nil {
		return fmt.Errorf("create postgres repo index: %w", err)
	}
	return nil
}

// The raw events table keeps exact source payloads for fidelity/debugging.
const createEventsTableSQL = `
CREATE TABLE IF NOT EXISTS events (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	repo_name TEXT NOT NULL,
	actor_name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	payload JSONB NOT NULL
);`

// Recent-event API queries will scan newest events first.
const createEventsCreatedAtIndexSQL = `
CREATE INDEX IF NOT EXISTS events_created_at_idx
ON events (created_at DESC);`

// Repo-specific API queries need fast lookups by repo and recency.
const createEventsRepoIndexSQL = `
CREATE INDEX IF NOT EXISTS events_repo_created_at_idx
ON events (repo_name, created_at DESC);`
