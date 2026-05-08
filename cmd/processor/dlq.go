package main

import (
	"context"
	"log/slog"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

type dlqPublisher interface {
	PublishRaw(context.Context, []byte, []byte) error
	Close() error
}

func publishFailedJob(ctx context.Context, workerID int, next job, dlq dlqPublisher) bool {
	if err := dlq.PublishRaw(ctx, next.message.Key, next.message.Value); err != nil {
		slog.Error(
			"dlq publish failed",
			"worker_id", workerID,
			"event_id", next.event.ID,
			"error", err,
		)
		return false
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
	return true
}

func publishMalformedMessage(
	ctx context.Context,
	message kafka.Message,
	dlq dlqPublisher,
) bool {
	if err := dlq.PublishRaw(ctx, message.Key, message.Value); err != nil {
		slog.Error(
			"malformed message dlq publish failed",
			"topic", message.Topic,
			"partition", message.Partition,
			"offset", message.Offset,
			"error", err,
		)
		return false
	}

	processorFailedEvents.Inc()
	slog.Warn(
		"published malformed message to dlq",
		"topic", message.Topic,
		"partition", message.Partition,
		"offset", message.Offset,
	)
	return true
}

func closeDLQProducer(producer dlqPublisher) {
	if err := producer.Close(); err != nil {
		slog.Error("dlq producer close failed", "error", err)
		return
	}
	slog.Info("dlq producer closed")
}
