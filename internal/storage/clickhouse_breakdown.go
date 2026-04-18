package storage

import (
	"context"
	"fmt"
	"time"
)

type EventBreakdown struct {
	EventType string `json:"event_type"`
	Count     uint64 `json:"count"`
}

// EventBreakdown returns activity counts grouped by event type.
func (s *ClickHouseStore) EventBreakdown(
	ctx context.Context,
	hours int,
) ([]EventBreakdown, error) {
	// Use the same API-owned time window style as trending queries.
	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	rows, err := s.conn.Query(ctx, eventBreakdownSQL, cutoff)
	if err != nil {
		return nil, fmt.Errorf("query event breakdown: %w", err)
	}
	defer rows.Close()

	var breakdown []EventBreakdown
	for rows.Next() {
		var item EventBreakdown
		if err := rows.Scan(&item.EventType, &item.Count); err != nil {
			return nil, fmt.Errorf("scan event breakdown: %w", err)
		}
		breakdown = append(breakdown, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read event breakdown: %w", err)
	}
	return breakdown, nil
}

const eventBreakdownSQL = `
SELECT event_type, sum(count) AS total
FROM events_hourly
WHERE hour >= ?
GROUP BY event_type
ORDER BY total DESC, event_type ASC`
