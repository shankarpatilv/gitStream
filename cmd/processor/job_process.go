package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/events"
)

// eventSink lets worker tests fake storage while production uses Postgres.
type eventSink interface {
	InsertEvent(context.Context, events.GitHubEvent) error
}

// processJob is the successful processing path protected by retry/DLQ logic.
func processJob(ctx context.Context, workerID int, job job, sink eventSink) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := sink.InsertEvent(ctx, job.event); err != nil {
		return err
	}
	logDecodedMessage(job.message, job.event, workerID)
	return nil
}
