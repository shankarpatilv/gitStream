package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/storage"
)

type postgresRecentEventsStore struct {
	cfg config
}

// RecentEvents opens Postgres for the request and runs the read-only query.
func (s postgresRecentEventsStore) RecentEvents(
	ctx context.Context,
	repo string,
	limit int,
) ([]storage.RecentEvent, error) {
	store, err := storage.NewPostgresStore(ctx, postgresConfig(s.cfg))
	if err != nil {
		return nil, err
	}
	defer store.Close()
	return store.RecentEvents(ctx, repo, limit)
}
