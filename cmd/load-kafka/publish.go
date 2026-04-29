package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

type rawPublisher interface {
	PublishRawBatch(context.Context, []kafka.RawMessage) error
}

func publishSyntheticEvents(ctx context.Context, producer rawPublisher, cfg config) error {
	started := time.Now()
	now := started.UTC()
	for start := 0; start < cfg.events; start += cfg.batchSize {
		end := min(start+cfg.batchSize, cfg.events)
		batch, err := syntheticBatch(cfg.prefix, start, end, now)
		if err != nil {
			return err
		}
		if err := producer.PublishRawBatch(ctx, batch); err != nil {
			return fmt.Errorf("publish synthetic batch %d-%d: %w", start, end, err)
		}
	}

	elapsed := time.Since(started)
	slog.Info("synthetic kafka load complete",
		"events", cfg.events,
		"topic", cfg.topic,
		"prefix", cfg.prefix,
		"duration_seconds", elapsed.Seconds(),
		"events_per_second", float64(cfg.events)/elapsed.Seconds(),
	)
	return nil
}

func syntheticBatch(prefix string, start, end int, now time.Time) ([]kafka.RawMessage, error) {
	batch := make([]kafka.RawMessage, 0, end-start)
	for i := start; i < end; i++ {
		key, value, err := syntheticKafkaEvent(prefix, i, now)
		if err != nil {
			return nil, err
		}
		batch = append(batch, kafka.RawMessage{Key: key, Value: value})
	}
	return batch, nil
}

var _ rawPublisher = (*kafka.Producer)(nil)
