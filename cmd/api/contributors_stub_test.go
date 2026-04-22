package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/storage"
)

type stubContributorsStore struct {
	contributors []storage.TopContributor
	err          error
	repo         string
	limit        int
}

func (s *stubContributorsStore) TopContributors(
	_ context.Context,
	repo string,
	limit int,
) ([]storage.TopContributor, error) {
	s.repo = repo
	s.limit = limit
	return s.contributors, s.err
}
