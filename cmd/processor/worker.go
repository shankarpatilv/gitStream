package main

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// startWorkers starts a bounded set of processors and reports when all exit.
func startWorkers(
	ctx context.Context,
	count int,
	jobs <-chan job,
	dlq dlqPublisher,
	sink eventSink,
	offsets *offsetTracker,
) <-chan struct{} {
	var wg sync.WaitGroup
	wg.Add(count)
	for id := 1; id <= count; id++ {
		go func(workerID int) {
			defer wg.Done()
			runWorker(ctx, workerID, jobs, dlq, sink, offsets)
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
func runWorker(
	ctx context.Context,
	id int,
	jobs <-chan job,
	dlq dlqPublisher,
	sink eventSink,
	offsets *offsetTracker,
) {
	processorActiveWorkers.Inc()
	defer processorActiveWorkers.Dec()

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
			start := time.Now()
			handleJob(ctx, id, next, dlq, offsets, process, sleepWithContext)
			processorProcessingDuration.Observe(time.Since(start).Seconds())
		}
	}
}

// handleJob routes exhausted processing failures to the DLQ.
func handleJob(
	ctx context.Context,
	workerID int,
	next job,
	dlq dlqPublisher,
	offsets *offsetTracker,
	process jobProcessor,
	sleep retrySleeper,
) {
	err := processJobWithRetry(ctx, workerID, next, process, sleep)
	if err == nil {
		if completeJobOffset(ctx, workerID, next, offsets) {
			processorProcessedEvents.Inc()
		}
		return
	}

	slog.Error(
		"processor job failed",
		"worker_id", workerID,
		"event_id", next.event.ID,
		"error", err,
	)
	if publishFailedJob(ctx, workerID, next, dlq) {
		completeJobOffset(ctx, workerID, next, offsets)
	}
}

func completeJobOffset(
	ctx context.Context,
	workerID int,
	next job,
	offsets *offsetTracker,
) bool {
	if err := offsets.Complete(ctx, next.message); err != nil {
		slog.Error(
			"kafka offset commit failed",
			"worker_id", workerID,
			"event_id", next.event.ID,
			"topic", next.message.Topic,
			"partition", next.message.Partition,
			"offset", next.message.Offset,
			"error", err,
		)
		return false
	}
	processorConsumerLag.Dec()
	return true
}
