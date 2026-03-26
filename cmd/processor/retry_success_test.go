package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

func TestProcessJobWithRetrySucceedsFirstAttempt(t *testing.T) {
	attempts := 0
	delays := []time.Duration{}

	err := processJobWithRetry(
		context.Background(),
		1,
		retryTestJob("event-1"),
		func(context.Context, int, job) error {
			attempts++
			return nil
		},
		recordDelays(&delays),
	)

	if err != nil {
		t.Fatalf("processJobWithRetry returned error: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
	if len(delays) != 0 {
		t.Fatalf("delays = %v, want none", delays)
	}
}

func TestProcessJobWithRetrySucceedsAfterRetry(t *testing.T) {
	attempts := 0
	delays := []time.Duration{}

	err := processJobWithRetry(
		context.Background(),
		1,
		retryTestJob("event-2"),
		func(context.Context, int, job) error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary failure")
			}
			return nil
		},
		recordDelays(&delays),
	)

	if err != nil {
		t.Fatalf("processJobWithRetry returned error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	want := []time.Duration{100 * time.Millisecond, 500 * time.Millisecond}
	if !reflect.DeepEqual(delays, want) {
		t.Fatalf("delays = %v, want %v", delays, want)
	}
}

func recordDelays(delays *[]time.Duration) retrySleeper {
	return func(_ context.Context, delay time.Duration) error {
		*delays = append(*delays, delay)
		return nil
	}
}

func retryTestJob(id string) job {
	return job{event: events.GitHubEvent{ID: id}}
}
