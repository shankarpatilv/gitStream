package main

import (
	"context"
	"testing"
	"time"
)

func TestWaitForWorkersReturnsWhenWorkersFinish(t *testing.T) {
	workersDone := make(chan struct{})
	close(workersDone)

	called := false
	ok := waitForWorkers(workersDone, time.Second, func() {
		called = true
	})

	if !ok {
		t.Fatal("waitForWorkers returned false, want true")
	}
	if called {
		t.Fatal("cancelWorkers was called after workers finished")
	}
}

func TestWaitForWorkersTimesOutAndCancelsWorkers(t *testing.T) {
	workersDone := make(chan struct{})
	called := false

	ok := waitForWorkers(workersDone, time.Millisecond, func() {
		called = true
	})

	if ok {
		t.Fatal("waitForWorkers returned true, want false")
	}
	if !called {
		t.Fatal("cancelWorkers was not called after timeout")
	}
}

func TestWorkersExitWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan job)
	workersDone := startWorkers(ctx, 1, jobs, &fakeDLQPublisher{}, &fakeEventSink{})

	cancel()

	select {
	case <-workersDone:
	case <-time.After(time.Second):
		t.Fatal("worker did not exit after context cancellation")
	}
}

func TestWorkersDrainClosedJobsChannel(t *testing.T) {
	ctx := context.Background()
	jobs := make(chan job, 2)
	jobs <- retryTestJob("event-1")
	jobs <- retryTestJob("event-2")
	close(jobs)

	workersDone := startWorkers(ctx, 1, jobs, &fakeDLQPublisher{}, &fakeEventSink{})

	select {
	case <-workersDone:
	case <-time.After(time.Second):
		t.Fatal("worker did not drain closed jobs channel")
	}
}
