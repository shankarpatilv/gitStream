package main

import (
	"context"
	"log/slog"
)

type dlqPublisher interface {
	PublishRaw(context.Context, []byte, []byte) error
	Close() error
}

func publishFailedJob(ctx context.Context, workerID int, next job, dlq dlqPublisher) {
	if err := dlq.PublishRaw(ctx, next.message.Key, next.message.Value); err != nil {
		slog.Error(
			"dlq publish failed",
			"worker_id", workerID,
			"event_id", next.event.ID,
			"error", err,
		)
		return
	}

	processorFailedEvents.Inc()
	slog.Warn(
		"published failed event to dlq",
		"worker_id", workerID,
		"event_id", next.event.ID,
		"topic", next.message.Topic,
		"partition", next.message.Partition,
		"offset", next.message.Offset,
	)
}

func closeDLQProducer(producer dlqPublisher) {
	if err := producer.Close(); err != nil {
		slog.Error("dlq producer close failed", "error", err)
		return
	}
	slog.Info("dlq producer closed")
}
