package main

import (
	"context"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

// dualEventSink keeps the worker path explicit: raw storage first, analytics second.
type dualEventSink struct {
	postgres   eventSink
	clickHouse eventSink
}

// InsertEvent succeeds only after both storage systems accept the event.
func (s dualEventSink) InsertEvent(ctx context.Context, event events.GitHubEvent) error {
	postgresStart := time.Now()
	if err := s.postgres.InsertEvent(ctx, event); err != nil {
		processorPostgresWriteDuration.Observe(time.Since(postgresStart).Seconds())
		return err
	}
	processorPostgresWriteDuration.Observe(time.Since(postgresStart).Seconds())

	clickHouseStart := time.Now()
	if err := s.clickHouse.InsertEvent(ctx, event); err != nil {
		processorClickHouseWriteDuration.Observe(time.Since(clickHouseStart).Seconds())
		return err
	}
	processorClickHouseWriteDuration.Observe(time.Since(clickHouseStart).Seconds())
	return nil
}
