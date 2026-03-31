package storage

import (
	"context"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

// run owns the pending batch; callers only communicate through channels.
func (s *ClickHouseBatchSink) run() {
	defer close(s.done)

	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	items := make([]clickHouseBatchItem, 0, s.maxSize)
	for {
		select {
		case item, ok := <-s.input:
			if !ok {
				s.flush(context.Background(), items)
				return
			}
			items = append(items, item)
			if len(items) >= s.maxSize {
				s.flush(context.Background(), items)
				items = items[:0]
			}
		case <-ticker.C:
			s.flush(context.Background(), items)
			items = items[:0]
		}
	}
}

// flush writes one ClickHouse batch and broadcasts the same result to its callers.
func (s *ClickHouseBatchSink) flush(
	ctx context.Context,
	items []clickHouseBatchItem,
) {
	if len(items) == 0 {
		return
	}

	batch := make([]events.GitHubEvent, 0, len(items))
	for _, item := range items {
		batch = append(batch, item.event)
	}
	err := s.writer.InsertAnalyticsBatch(ctx, batch)
	for _, item := range items {
		item.result <- err
	}
}
