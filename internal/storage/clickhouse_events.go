package storage

import (
	"context"
	"fmt"

	"github.com/vivekspatil/gitstream/internal/events"
)

// InsertAnalyticsBatch writes both ClickHouse analytics shapes for each event.
func (s *ClickHouseStore) InsertAnalyticsBatch(
	ctx context.Context,
	batch []events.GitHubEvent,
) error {
	if len(batch) == 0 {
		return nil
	}
	if err := s.insertHourlyBatch(ctx, batch); err != nil {
		return err
	}
	if err := s.insertTimeseriesBatch(ctx, batch); err != nil {
		return err
	}
	return nil
}

func (s *ClickHouseStore) insertHourlyBatch(
	ctx context.Context,
	events []events.GitHubEvent,
) error {
	batch, err := s.conn.PrepareBatch(ctx, insertEventsHourlySQL)
	if err != nil {
		return fmt.Errorf("prepare events_hourly batch: %w", err)
	}
	for _, event := range events {
		// Truncating to the hour gives the API a stable bucket for trending queries.
		hour := event.CreatedAt.UTC().Truncate(eventHour)
		if err := batch.Append(hour, event.RepoName, event.Type, uint64(1)); err != nil {
			return fmt.Errorf("append events_hourly row: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("send events_hourly batch: %w", err)
	}
	return nil
}

func (s *ClickHouseStore) insertTimeseriesBatch(
	ctx context.Context,
	events []events.GitHubEvent,
) error {
	batch, err := s.conn.PrepareBatch(ctx, insertEventsTimeseriesSQL)
	if err != nil {
		return fmt.Errorf("prepare events_timeseries batch: %w", err)
	}
	for _, event := range events {
		// This row is intentionally small; Postgres owns the full raw JSON payload.
		if err := batch.Append(event.CreatedAt.UTC(), event.Type, event.RepoName); err != nil {
			return fmt.Errorf("append events_timeseries row: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("send events_timeseries batch: %w", err)
	}
	return nil
}
