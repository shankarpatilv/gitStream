package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/storage"
)

type postgresContributorsStore struct {
	cfg config
}

// TopContributors opens Postgres for the request and runs the read-only query.
func (s postgresContributorsStore) TopContributors(
	ctx context.Context,
	repo string,
	limit int,
) ([]storage.TopContributor, error) {
	store, err := storage.NewPostgresStore(ctx, postgresConfig(s.cfg))
	if err != nil {
		return nil, err
	}
	defer store.Close()
	return store.TopContributors(ctx, repo, limit)
}
