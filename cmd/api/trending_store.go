package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/storage"
)

type clickHouseTrendingStore struct {
	cfg config
}

// TrendingRepos opens ClickHouse for the request and runs the read-only query.
func (s clickHouseTrendingStore) TrendingRepos(
	ctx context.Context,
	hours int,
	limit int,
) ([]storage.TrendingRepo, error) {
	store, err := storage.NewClickHouseStore(ctx, clickHouseConfig(s.cfg))
	if err != nil {
		return nil, err
	}
	defer store.Close()
	return store.TrendingRepos(ctx, hours, limit)
}
