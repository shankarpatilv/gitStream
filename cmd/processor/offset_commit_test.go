package main

import (
	"context"
	"errors"
	"testing"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

type fakeOffsetCommitter struct {
	commits []kafka.Message
	err     error
}

func (f *fakeOffsetCommitter) CommitMessage(_ context.Context, message kafka.Message) error {
	if f.err != nil {
		return f.err
	}
	f.commits = append(f.commits, message)
	return nil
}

func TestOffsetTrackerCommitsContiguousOffsets(t *testing.T) {
	committer := &fakeOffsetCommitter{}
	tracker := newOffsetTracker(committer)

	registerOffsets(tracker, 10, 11, 12)

	assertNoCommitAfterComplete(t, tracker, committer, 12)
	assertCommitAfterComplete(t, tracker, committer, 10, 10)
	assertCommitAfterComplete(t, tracker, committer, 11, 12)
}

func TestOffsetTrackerTracksPartitionsIndependently(t *testing.T) {
	committer := &fakeOffsetCommitter{}
	tracker := newOffsetTracker(committer)

	first := testMessage(0, 5)
	second := testMessage(1, 20)
	tracker.Register(first)
	tracker.Register(second)

	if err := tracker.Complete(context.Background(), second); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if got := lastCommit(committer).Offset; got != 20 {
		t.Fatalf("last commit offset = %d, want 20", got)
	}
}

func TestOffsetTrackerKeepsOffsetAfterCommitFailure(t *testing.T) {
	committer := &fakeOffsetCommitter{err: errors.New("kafka unavailable")}
	tracker := newOffsetTracker(committer)
	message := testMessage(0, 10)
	tracker.Register(message)

	if err := tracker.Complete(context.Background(), message); err == nil {
		t.Fatal("expected commit error, got nil")
	}

	committer.err = nil
	if err := tracker.Complete(context.Background(), message); err != nil {
		t.Fatalf("Complete returned error after retry: %v", err)
	}
	if got := lastCommit(committer).Offset; got != 10 {
		t.Fatalf("last commit offset = %d, want 10", got)
	}
}

func registerOffsets(tracker *offsetTracker, offsets ...int64) {
	for _, offset := range offsets {
		tracker.Register(testMessage(0, offset))
	}
}

func assertNoCommitAfterComplete(
	t *testing.T,
	tracker *offsetTracker,
	committer *fakeOffsetCommitter,
	offset int64,
) {
	t.Helper()
	if err := tracker.Complete(context.Background(), testMessage(0, offset)); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if len(committer.commits) != 0 {
		t.Fatalf("commits = %v, want none", committer.commits)
	}
}

func assertCommitAfterComplete(
	t *testing.T,
	tracker *offsetTracker,
	committer *fakeOffsetCommitter,
	offset int64,
	want int64,
) {
	t.Helper()
	if err := tracker.Complete(context.Background(), testMessage(0, offset)); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if got := lastCommit(committer).Offset; got != want {
		t.Fatalf("last commit offset = %d, want %d", got, want)
	}
}

func lastCommit(committer *fakeOffsetCommitter) kafka.Message {
	return committer.commits[len(committer.commits)-1]
}

func testMessage(partition int, offset int64) kafka.Message {
	return kafka.Message{
		Topic:     "github-events",
		Partition: partition,
		Offset:    offset,
	}
}
