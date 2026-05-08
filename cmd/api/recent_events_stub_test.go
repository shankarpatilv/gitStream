package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/storage"
)

type stubRecentEventsStore struct {
	events []storage.RecentEvent
	err    error
	repo   string
	limit  int
}

func (s *stubRecentEventsStore) RecentEvents(
	_ context.Context,
	repo string,
	limit int,
) ([]storage.RecentEvent, error) {
	s.repo = repo
	s.limit = limit
	return s.events, s.err
}
