package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeDLQPublisher struct {
	key     []byte
	value   []byte
	writes  int
	close   int
	failErr error
}

func (f *fakeDLQPublisher) PublishRaw(_ context.Context, key, value []byte) error {
	f.writes++
	f.key = append([]byte(nil), key...)
	f.value = append([]byte(nil), value...)
	return f.failErr
}

func (f *fakeDLQPublisher) Close() error {
	f.close++
	return nil
}

func TestHandleJobPublishesExhaustedFailureToDLQ(t *testing.T) {
	dlq := &fakeDLQPublisher{}
	next := retryTestJob("event-dlq")
	next.message.Key = []byte("owner/repo")
	next.message.Value = []byte(`{"id":"event-dlq"}`)
	tracker := testOffsetTracker()
	tracker.Register(next.message)

	handleJob(
		context.Background(),
		1,
		next,
		dlq,
		tracker,
		func(context.Context, int, job) error {
			return errors.New("processing failed")
		},
		recordDelays(&[]time.Duration{}),
	)

	if dlq.writes != 1 {
		t.Fatalf("dlq writes = %d, want 1", dlq.writes)
	}
	if string(dlq.key) != "owner/repo" {
		t.Fatalf("dlq key = %q, want owner/repo", string(dlq.key))
	}
	if string(dlq.value) != `{"id":"event-dlq"}` {
		t.Fatalf("dlq value = %q, want original payload", string(dlq.value))
	}
}

func TestHandleJobSuccessDoesNotPublishToDLQ(t *testing.T) {
	dlq := &fakeDLQPublisher{}

	handleJob(
		context.Background(),
		1,
		retryTestJob("event-ok"),
		dlq,
		testOffsetTracker(),
		func(context.Context, int, job) error {
			return nil
		},
		recordDelays(&[]time.Duration{}),
	)

	if dlq.writes != 0 {
		t.Fatalf("dlq writes = %d, want 0", dlq.writes)
	}
}

func TestHandleJobSuccessCommitsOffset(t *testing.T) {
	committer := &fakeOffsetCommitter{}
	tracker := newOffsetTracker(committer)
	next := retryTestJob("event-ok")
	next.message = testMessage(0, 10)
	tracker.Register(next.message)

	handleJob(
		context.Background(),
		1,
		next,
		&fakeDLQPublisher{},
		tracker,
		func(context.Context, int, job) error {
			return nil
		},
		recordDelays(&[]time.Duration{}),
	)

	if got := lastCommit(committer).Offset; got != 10 {
		t.Fatalf("committed offset = %d, want 10", got)
	}
}

func TestHandleJobCommitsAfterDLQSuccess(t *testing.T) {
	committer := &fakeOffsetCommitter{}
	tracker := newOffsetTracker(committer)
	next := retryTestJob("event-dlq")
	next.message = testMessage(0, 11)
	tracker.Register(next.message)

	handleJob(
		context.Background(),
		1,
		next,
		&fakeDLQPublisher{},
		tracker,
		func(context.Context, int, job) error {
			return errors.New("processing failed")
		},
		recordDelays(&[]time.Duration{}),
	)

	if got := lastCommit(committer).Offset; got != 11 {
		t.Fatalf("committed offset = %d, want 11", got)
	}
}

func TestHandleJobDoesNotCommitAfterDLQFailure(t *testing.T) {
	committer := &fakeOffsetCommitter{}
	tracker := newOffsetTracker(committer)
	next := retryTestJob("event-dlq")
	next.message = testMessage(0, 12)
	tracker.Register(next.message)

	handleJob(
		context.Background(),
		1,
		next,
		&fakeDLQPublisher{failErr: errors.New("kafka unavailable")},
		tracker,
		func(context.Context, int, job) error {
			return errors.New("processing failed")
		},
		recordDelays(&[]time.Duration{}),
	)

	if len(committer.commits) != 0 {
		t.Fatalf("commits = %v, want none", committer.commits)
	}
}
