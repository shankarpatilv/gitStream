package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/vivekspatil/gitstream/internal/events"
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

		enqueueJob(ctx, jobs, job{message: message, event: event})
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

// logMalformedMessage records enough Kafka context to debug bad payloads.
func logMalformedMessage(message kafka.Message, err error) {
	slog.Warn(
		"skipping malformed kafka message",
		"topic", message.Topic,
		"partition", message.Partition,
		"offset", message.Offset,
		"key", string(message.Key),
		"error", err,
	)
}

// logDecodedMessage records Kafka metadata and normalized event fields.
func logDecodedMessage(message kafka.Message, event events.GitHubEvent, workerID int) {
	slog.Info(
		"processed kafka message",
		"worker_id", workerID,
		"topic", message.Topic,
		"partition", message.Partition,
		"offset", message.Offset,
		"key", string(message.Key),
		"event_id", event.ID,
		"event_type", event.Type,
		"repo", event.RepoName,
		"actor", event.ActorName,
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
