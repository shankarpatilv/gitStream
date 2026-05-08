package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/storage"
)

type stubTrendingStore struct {
	repos []storage.TrendingRepo
	err   error
	hours int
	limit int
}

func (s *stubTrendingStore) TrendingRepos(
	_ context.Context,
	hours int,
	limit int,
) ([]storage.TrendingRepo, error) {
	s.hours = hours
	s.limit = limit
	return s.repos, s.err
}
