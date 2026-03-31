package main

import (
	"context"
	"testing"
	"time"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

type fakeConsumer struct {
	messages []kafka.Message
	index    int
}

func (f *fakeConsumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if f.index >= len(f.messages) {
		<-ctx.Done()
		return kafka.Message{}, ctx.Err()
	}
	message := f.messages[f.index]
	f.index++
	return message, nil
}

func (f *fakeConsumer) Close() error {
	return nil
}

func TestRunConsumerEnqueuesValidMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan job, 1)
	consumer := &fakeConsumer{
		messages: []kafka.Message{validMessage("event-1")},
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		runConsumer(ctx, consumer, jobs, &fakeDLQPublisher{}, testOffsetTracker())
	}()

	received := <-jobs
	if received.event.ID != "event-1" {
		t.Fatalf("event ID = %q, want event-1", received.event.ID)
	}

	cancel()
	<-done
}

func TestRunConsumerSkipsMalformedMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan job, 1)
	dlq := &fakeDLQPublisher{}
	committer := &fakeOffsetCommitter{}
	consumer := &fakeConsumer{
		messages: []kafka.Message{
			{
				Topic:  "github-events",
				Offset: 0,
				Key:    []byte("owner/repo"),
				Value:  []byte(`{"id":`),
			},
			validMessage("event-2"),
		},
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		runConsumer(ctx, consumer, jobs, dlq, newOffsetTracker(committer))
	}()

	received := <-jobs
	if received.event.ID != "event-2" {
		t.Fatalf("event ID = %q, want event-2", received.event.ID)
	}
	if dlq.writes != 1 {
		t.Fatalf("dlq writes = %d, want 1", dlq.writes)
	}
	if len(committer.commits) != 1 || committer.commits[0].Offset != 0 {
		t.Fatalf("commits = %#v, want offset 0", committer.commits)
	}

	cancel()
	<-done
}

func TestEnqueueJobBlocksWhenQueueIsFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := make(chan job, 1)
	jobs <- job{}

	result := make(chan bool, 1)
	go func() {
		result <- enqueueJob(ctx, jobs, job{})
	}()

	select {
	case <-result:
		t.Fatal("enqueueJob returned while the jobs channel was full")
	case <-time.After(20 * time.Millisecond):
	}

	cancel()
	if ok := <-result; ok {
		t.Fatal("enqueueJob returned true after context cancellation")
	}
}

func validMessage(id string) kafka.Message {
	return kafka.Message{
		Topic:     "github-events",
		Partition: 0,
		Offset:    0,
		Key:       []byte("owner/repo"),
		Value: []byte(`{
			"id": "` + id + `",
			"type": "PushEvent",
			"repo": {"name": "owner/repo"},
			"actor": {"login": "octocat"},
			"created_at": "2026-05-07T12:00:00Z"
		}`),
	}
}

func testOffsetTracker() *offsetTracker {
	return newOffsetTracker(&fakeOffsetCommitter{})
}

var _ messageConsumer = (*fakeConsumer)(nil)
