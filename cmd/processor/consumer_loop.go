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
func runConsumer(ctx context.Context, consumer messageConsumer, jobs chan<- job) {
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

		event, err := decodeMessageEvent(message)
		if err != nil {
			logMalformedMessage(message, err)
			continue
		}

		if ok := enqueueJob(ctx, jobs, job{message: message, event: event}); !ok {
			slog.Info("kafka consumer enqueue stopped", "error", ctx.Err())
			return
		}
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
