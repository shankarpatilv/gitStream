package storage

import (
	"context"
	"testing"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

type fakeClickHouseWriter struct {
	batches [][]events.GitHubEvent
	err     error
}

func (f *fakeClickHouseWriter) InsertAnalyticsBatch(
	_ context.Context,
	batch []events.GitHubEvent,
) error {
	copied := append([]events.GitHubEvent(nil), batch...)
	f.batches = append(f.batches, copied)
	return f.err
}

func TestClickHouseBatchSinkFlushesAtMaxSize(t *testing.T) {
	writer := &fakeClickHouseWriter{}
	sink := NewClickHouseBatchSink(writer, 2, time.Hour)
	defer sink.Close(context.Background())

	first := insertAsync(sink, testEvent("one"))
	assertStillWaiting(t, first)

	second := insertAsync(sink, testEvent("two"))
	assertNoError(t, <-first)
	assertNoError(t, <-second)

	if len(writer.batches) != 1 || len(writer.batches[0]) != 2 {
		t.Fatalf("batches = %#v, want one batch with two events", writer.batches)
	}
}

func TestClickHouseBatchSinkFlushesOnInterval(t *testing.T) {
	writer := &fakeClickHouseWriter{}
	sink := NewClickHouseBatchSink(writer, 100, 10*time.Millisecond)
	defer sink.Close(context.Background())

	result := insertAsync(sink, testEvent("one"))
	assertNoError(t, <-result)

	if len(writer.batches) != 1 || len(writer.batches[0]) != 1 {
		t.Fatalf("batches = %#v, want one batch with one event", writer.batches)
	}
}

func TestClickHouseBatchSinkCloseFlushesPendingEvents(t *testing.T) {
	writer := &fakeClickHouseWriter{}
	sink := NewClickHouseBatchSink(writer, 100, time.Hour)

	result := insertAsync(sink, testEvent("one"))
	assertStillWaiting(t, result)
	assertNoError(t, sink.Close(context.Background()))
	assertNoError(t, <-result)

	if len(writer.batches) != 1 || len(writer.batches[0]) != 1 {
		t.Fatalf("batches = %#v, want one batch with one event", writer.batches)
	}
}

func insertAsync(sink *ClickHouseBatchSink, event events.GitHubEvent) <-chan error {
	result := make(chan error, 1)
	go func() {
		result <- sink.InsertEvent(context.Background(), event)
	}()
	return result
}

func assertStillWaiting(t *testing.T, result <-chan error) {
	t.Helper()
	select {
	case err := <-result:
		t.Fatalf("InsertEvent returned early with error %v", err)
	case <-time.After(20 * time.Millisecond):
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func testEvent(id string) events.GitHubEvent {
	return events.GitHubEvent{
		ID:        id,
		Type:      "PushEvent",
		RepoName:  "owner/repo",
		ActorName: "octocat",
		CreatedAt: time.Date(2026, 5, 8, 12, 30, 0, 0, time.UTC),
	}
}
