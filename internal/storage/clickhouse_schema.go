package storage

import (
	"context"
	"fmt"
)

// EnsureSchema creates analytics tables used by the API's ClickHouse queries.
func (s *ClickHouseStore) EnsureSchema(ctx context.Context) error {
	if err := s.conn.Exec(ctx, createEventsHourlySQL); err != nil {
		return fmt.Errorf("create clickhouse events_hourly table: %w", err)
	}
	if err := s.conn.Exec(ctx, createEventsTimeseriesSQL); err != nil {
		return fmt.Errorf("create clickhouse events_timeseries table: %w", err)
	}
	return nil
}

const createEventsHourlySQL = `
CREATE TABLE IF NOT EXISTS events_hourly (
	hour DateTime,
	repo_name String,
	event_type String,
	count UInt64
) ENGINE = SummingMergeTree()
ORDER BY (hour, repo_name, event_type);`

// The timeseries table keeps event-time rows without the large raw JSON payload.
const createEventsTimeseriesSQL = `
CREATE TABLE IF NOT EXISTS events_timeseries (
	timestamp DateTime,
	event_type String,
	repo_name String
) ENGINE = MergeTree()
ORDER BY (timestamp, repo_name);`
