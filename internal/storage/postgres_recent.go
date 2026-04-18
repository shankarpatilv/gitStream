package storage

import (
	"context"
	"fmt"
	"time"
)

type RecentEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	RepoName  string    `json:"repo_name"`
	ActorName string    `json:"actor_name"`
	CreatedAt time.Time `json:"created_at"`
}

// RecentEvents returns newest raw events for one repository from Postgres.
func (s *PostgresStore) RecentEvents(
	ctx context.Context,
	repo string,
	limit int,
) ([]RecentEvent, error) {
	rows, err := s.pool.Query(ctx, recentEventsSQL, repo, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent events: %w", err)
	}
	defer rows.Close()

	var events []RecentEvent
	for rows.Next() {
		var event RecentEvent
		if err := rows.Scan(
			&event.ID,
			&event.Type,
			&event.RepoName,
			&event.ActorName,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan recent event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read recent events: %w", err)
	}
	return events, nil
}

const recentEventsSQL = `
SELECT id, type, repo_name, actor_name, created_at
FROM events
WHERE repo_name = $1
ORDER BY created_at DESC
LIMIT $2;`
