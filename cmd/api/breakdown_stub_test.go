package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/storage"
)

type stubBreakdownStore struct {
	breakdown []storage.EventBreakdown
	err       error
	hours     int
}

func (s *stubBreakdownStore) EventBreakdown(
	_ context.Context,
	hours int,
) ([]storage.EventBreakdown, error) {
	s.hours = hours
	return s.breakdown, s.err
}
