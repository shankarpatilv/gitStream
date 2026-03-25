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

// runConsumer fetches Kafka messages without committing offsets yet.
func runConsumer(ctx context.Context, consumer messageConsumer) {
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

		logConsumedMessage(message)
	}
}

// logConsumedMessage records Kafka metadata before JSON decoding exists.
func logConsumedMessage(message kafka.Message) {
	slog.Info(
		"consumed kafka message",
		"topic", message.Topic,
		"partition", message.Partition,
		"offset", message.Offset,
		"key", string(message.Key),
	)
}

// closeConsumer releases the Kafka reader connection on processor exit.
func closeConsumer(consumer messageConsumer) {
	if err := consumer.Close(); err != nil {
		slog.Error("kafka consumer close failed", "error", err)
		return
	}
	slog.Info("kafka consumer closed")
}
