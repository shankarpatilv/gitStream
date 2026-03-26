package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

var retryDelays = []time.Duration{
	100 * time.Millisecond,
	500 * time.Millisecond,
	2 * time.Second,
}

type jobProcessor func(context.Context, int, job) error
type retrySleeper func(context.Context, time.Duration) error

func processJobWithRetry(
	ctx context.Context,
	workerID int,
	next job,
	process jobProcessor,
	sleep retrySleeper,
) error {
	maxAttempts := len(retryDelays) + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := process(ctx, workerID, next)
		if err == nil {
			return nil
		}

		logFailedAttempt(workerID, next, attempt, err)
		if attempt == maxAttempts {
			return fmt.Errorf(
				"process event %q after %d attempts: %w",
				next.event.ID,
				attempt,
				err,
			)
		}
		if err := sleep(ctx, retryDelays[attempt-1]); err != nil {
			return fmt.Errorf("wait before retrying event %q: %w", next.event.ID, err)
		}
	}
	return nil
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func logFailedAttempt(workerID int, next job, attempt int, err error) {
	slog.Warn(
		"processor job attempt failed",
		"worker_id", workerID,
		"event_id", next.event.ID,
		"attempt", attempt,
		"error", err,
	)
}
