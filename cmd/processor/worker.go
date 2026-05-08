package main

import (
	"context"
	"log/slog"
	"sync"
)

func startWorkers(ctx context.Context, count int, jobs <-chan job, dlq dlqPublisher) <-chan struct{} {
	var wg sync.WaitGroup
	wg.Add(count)
	for id := 1; id <= count; id++ {
		go func(workerID int) {
			defer wg.Done()
			runWorker(ctx, workerID, jobs, dlq)
		}(id)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

func runWorker(ctx context.Context, id int, jobs <-chan job, dlq dlqPublisher) {
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
			handleJob(ctx, id, next, dlq, processJob, sleepWithContext)
		}
	}
}

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

func processJob(ctx context.Context, workerID int, job job) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	logDecodedMessage(job.message, job.event, workerID)
	return nil
}
