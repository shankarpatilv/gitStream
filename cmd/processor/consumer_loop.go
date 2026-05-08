package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

type messageConsumer interface {
	FetchMessage(context.Context) (kafka.Message, error)
	Close() error
}

// runConsumer fetches Kafka messages and enqueues decoded events for workers.
func runConsumer(
	ctx context.Context,
	consumer messageConsumer,
	jobs chan<- job,
	dlq dlqPublisher,
	offsets *offsetTracker,
) {
	for {
		message, err := consumer.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Info("kafka consumer stopped", "error", err)
				return
			}
			slog.Error("kafka fetch failed", "error", err)
			continue
		}
		offsets.Register(message)

		event, err := decodeMessageEvent(message)
		if err != nil {
			logMalformedMessage(message, err)
			handleMalformedMessage(ctx, message, dlq, offsets)
			continue
		}

		if ok := enqueueJob(ctx, jobs, job{message: message, event: event}); !ok {
			slog.Info("kafka consumer enqueue stopped", "error", ctx.Err())
			return
		}
	}
}

func handleMalformedMessage(
	ctx context.Context,
	message kafka.Message,
	dlq dlqPublisher,
	offsets *offsetTracker,
) {
	if !publishMalformedMessage(ctx, message, dlq) {
		return
	}
	if err := offsets.Complete(ctx, message); err != nil {
		slog.Error(
			"kafka offset commit failed",
			"topic", message.Topic,
			"partition", message.Partition,
			"offset", message.Offset,
			"error", err,
		)
	}
}

func enqueueJob(ctx context.Context, jobs chan<- job, next job) bool {
	select {
	case <-ctx.Done():
		return false
	case jobs <- next:
		return true
	}
}
