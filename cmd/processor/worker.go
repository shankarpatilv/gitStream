package main

import (
	"context"
	"log/slog"
)

func startWorkers(ctx context.Context, count int, jobs <-chan job) {
	for id := 1; id <= count; id++ {
		go runWorker(ctx, id, jobs)
	}
}

func runWorker(ctx context.Context, id int, jobs <-chan job) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("processor worker stopped", "worker_id", id, "error", ctx.Err())
			return
		case job := <-jobs:
			logProcessedJob(id, job)
		}
	}
}

func logProcessedJob(workerID int, job job) {
	logDecodedMessage(job.message, job.event, workerID)
}
