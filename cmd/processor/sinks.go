package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/events"
)

// dualEventSink keeps the worker path explicit: raw storage first, analytics second.
type dualEventSink struct {
	postgres   eventSink
	clickHouse eventSink
}

// InsertEvent succeeds only after both storage systems accept the event.
func (s dualEventSink) InsertEvent(ctx context.Context, event events.GitHubEvent) error {
	if err := s.postgres.InsertEvent(ctx, event); err != nil {
		return err
	}
	if err := s.clickHouse.InsertEvent(ctx, event); err != nil {
		return err
	}
	return nil
}
