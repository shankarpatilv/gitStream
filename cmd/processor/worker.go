package main

import (
	"context"
	"log/slog"
	"sync"
)

// startWorkers starts a bounded set of processors and reports when all exit.
func startWorkers(
	ctx context.Context,
	count int,
	jobs <-chan job,
	dlq dlqPublisher,
	sink eventSink,
) <-chan struct{} {
	var wg sync.WaitGroup
	wg.Add(count)
	for id := 1; id <= count; id++ {
		go func(workerID int) {
			defer wg.Done()
			runWorker(ctx, workerID, jobs, dlq, sink)
		}(id)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

// runWorker drains queued jobs until shutdown cancellation or channel close.
func runWorker(ctx context.Context, id int, jobs <-chan job, dlq dlqPublisher, sink eventSink) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("processor worker stopped", "worker_id", id, "error", ctx.Err())
			return
		case next, ok := <-jobs:
			if !ok {
				slog.Info("processor worker drained", "worker_id", id)
				return
			}
			process := func(ctx context.Context, workerID int, next job) error {
				return processJob(ctx, workerID, next, sink)
			}
			handleJob(ctx, id, next, dlq, process, sleepWithContext)
		}
	}
}

// handleJob routes exhausted processing failures to the DLQ.
func handleJob(
	ctx context.Context,
	workerID int,
	next job,
	dlq dlqPublisher,
	process jobProcessor,
	sleep retrySleeper,
) {
	err := processJobWithRetry(ctx, workerID, next, process, sleep)
	if err == nil {
		return
	}

	slog.Error(
		"processor job failed",
		"worker_id", workerID,
		"event_id", next.event.ID,
		"error", err,
	)
	publishFailedJob(ctx, workerID, next, dlq)
}
