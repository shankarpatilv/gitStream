package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/storage"
)

type clickHouseBreakdownStore struct {
	cfg config
}

// EventBreakdown opens ClickHouse for the request and runs the read-only query.
func (s clickHouseBreakdownStore) EventBreakdown(
	ctx context.Context,
	hours int,
) ([]storage.EventBreakdown, error) {
	store, err := storage.NewClickHouseStore(ctx, clickHouseConfig(s.cfg))
	if err != nil {
		return nil, err
	}
	defer store.Close()
	return store.EventBreakdown(ctx, hours)
}
