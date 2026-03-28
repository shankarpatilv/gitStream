package main

import (
	"context"
	"errors"
	"testing"

	"github.com/vivekspatil/gitstream/internal/events"
)

type fakeEventSink struct {
	event   events.GitHubEvent
	writes  int
	failErr error
}

func (f *fakeEventSink) InsertEvent(_ context.Context, event events.GitHubEvent) error {
	f.writes++
	f.event = event
	return f.failErr
}

func TestProcessJobWritesEventToSink(t *testing.T) {
	sink := &fakeEventSink{}
	next := retryTestJob("event-sink")

	err := processJob(context.Background(), 1, next, sink)
	if err != nil {
		t.Fatalf("processJob returned error: %v", err)
	}
	if sink.writes != 1 {
		t.Fatalf("sink writes = %d, want 1", sink.writes)
	}
	if sink.event.ID != "event-sink" {
		t.Fatalf("sink event ID = %q, want event-sink", sink.event.ID)
	}
}

func TestProcessJobReturnsSinkFailure(t *testing.T) {
	sink := &fakeEventSink{failErr: errors.New("postgres unavailable")}

	err := processJob(context.Background(), 1, retryTestJob("event-sink"), sink)
	if err == nil {
		t.Fatal("expected sink error, got nil")
	}
}
