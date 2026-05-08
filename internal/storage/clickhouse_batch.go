package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

type clickHouseBatchWriter interface {
	InsertAnalyticsBatch(context.Context, []events.GitHubEvent) error
}

// ClickHouseBatchSink batches analytics writes while still reporting per-event results.
type ClickHouseBatchSink struct {
	writer        clickHouseBatchWriter
	maxSize       int
	flushInterval time.Duration
	input         chan clickHouseBatchItem
	done          chan struct{}
	closeOnce     sync.Once
}

type clickHouseBatchItem struct {
	event  events.GitHubEvent
	result chan error
}

// NewClickHouseBatchSink starts the single goroutine that owns batch state.
func NewClickHouseBatchSink(
	writer clickHouseBatchWriter,
	maxSize int,
	flushInterval time.Duration,
) *ClickHouseBatchSink {
	sink := &ClickHouseBatchSink{
		writer:        writer,
		maxSize:       maxSize,
		flushInterval: flushInterval,
		input:         make(chan clickHouseBatchItem),
		done:          make(chan struct{}),
	}
	go sink.run()
	return sink
}

// InsertEvent waits until this event's batch flushes, so worker success is real.
func (s *ClickHouseBatchSink) InsertEvent(
	ctx context.Context,
	event events.GitHubEvent,
) error {
	item := clickHouseBatchItem{
		event:  event,
		result: make(chan error, 1),
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.input <- item:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-item.result:
		return err
	}
}

// Close stops accepting new events and flushes any pending batch before returning.
func (s *ClickHouseBatchSink) Close(ctx context.Context) error {
	s.closeOnce.Do(func() {
		close(s.input)
	})

	select {
	case <-ctx.Done():
		return fmt.Errorf("close clickhouse batch sink: %w", ctx.Err())
	case <-s.done:
		return nil
	}
}
