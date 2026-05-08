package main

import (
	"context"
	"log/slog"
	"time"
)

func waitForWorkers(
	workersDone <-chan struct{},
	timeout time.Duration,
	cancelWorkers context.CancelFunc,
) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-workersDone:
		slog.Info("processor workers drained")
		return true
	case <-timer.C:
		slog.Error("processor worker drain timed out", "timeout", timeout)
		cancelWorkers()
		return false
	}
}
