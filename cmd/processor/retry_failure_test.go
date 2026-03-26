package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestProcessJobWithRetryExhaustsAttempts(t *testing.T) {
	attempts := 0
	delays := []time.Duration{}

	err := processJobWithRetry(
		context.Background(),
		1,
		retryTestJob("event-3"),
		func(context.Context, int, job) error {
			attempts++
			return errors.New("still failing")
		},
		recordDelays(&delays),
	)

	if err == nil {
		t.Fatal("expected retry exhaustion error, got nil")
	}
	if attempts != 4 {
		t.Fatalf("attempts = %d, want 4", attempts)
	}
	want := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		2 * time.Second,
	}
	if !reflect.DeepEqual(delays, want) {
		t.Fatalf("delays = %v, want %v", delays, want)
	}
}

func TestSleepWithContextStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sleepWithContext(ctx, time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("sleepWithContext error = %v, want context.Canceled", err)
	}
}
