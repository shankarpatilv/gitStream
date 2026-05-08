package storage

import (
	"context"
	"fmt"

	"github.com/vivekspatil/gitstream/internal/events"
)

// InsertEvent stores the raw event idempotently by GitHub event ID.
func (s *PostgresStore) InsertEvent(ctx context.Context, event events.GitHubEvent) error {
	_, err := s.pool.Exec(
		ctx,
		insertEventSQL,
		event.ID,
		event.Type,
		event.RepoName,
		event.ActorName,
		event.CreatedAt,
		event.Payload,
	)
	if err != nil {
		return fmt.Errorf("insert postgres event %q: %w", event.ID, err)
	}
	return nil
}

// Reprocessed Kafka messages should not duplicate raw Postgres rows.
const insertEventSQL = `
INSERT INTO events (
	id,
	type,
	repo_name,
	actor_name,
	created_at,
	payload
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO NOTHING;`
