package main

import (
	"context"
	"log/slog"
	"sync"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

type offsetCommitter interface {
	CommitMessage(context.Context, kafka.Message) error
}

type offsetTracker struct {
	committer  offsetCommitter
	mu         sync.Mutex
	partitions map[offsetPartition]*partitionOffsets
}

type offsetPartition struct {
	topic     string
	partition int
}

type partitionOffsets struct {
	next      int64
	started   bool
	completed map[int64]kafka.Message
}

func newOffsetTracker(committer offsetCommitter) *offsetTracker {
	return &offsetTracker{
		committer:  committer,
		partitions: make(map[offsetPartition]*partitionOffsets),
	}
}

// Register records that an offset is now in processor-owned work.
func (t *offsetTracker) Register(message kafka.Message) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.partitionState(message)
	if !state.started {
		state.next = message.Offset
		state.started = true
	}
}

// Complete commits the highest contiguous completed offset for this partition.
func (t *offsetTracker) Complete(ctx context.Context, message kafka.Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.partitionState(message)
	if !state.started || message.Offset < state.next {
		return nil
	}

	state.completed[message.Offset] = message
	return t.commitReady(ctx, state)
}

func (t *offsetTracker) partitionState(message kafka.Message) *partitionOffsets {
	key := offsetPartition{topic: message.Topic, partition: message.Partition}
	state, ok := t.partitions[key]
	if ok {
		return state
	}

	state = &partitionOffsets{
		completed: make(map[int64]kafka.Message),
	}
	t.partitions[key] = state
	return state
}

func (t *offsetTracker) commitReady(
	ctx context.Context,
	state *partitionOffsets,
) error {
	var last kafka.Message
	ready := make([]int64, 0)
	for {
		message, ok := state.completed[state.next]
		if !ok {
			break
		}
		last = message
		ready = append(ready, state.next)
		state.next++
	}
	if len(ready) == 0 {
		return nil
	}
	if err := t.committer.CommitMessage(ctx, last); err != nil {
		state.next = ready[0]
		return err
	}
	for _, offset := range ready {
		delete(state.completed, offset)
	}
	slog.Info(
		"committed kafka offset",
		"topic", last.Topic,
		"partition", last.Partition,
		"offset", last.Offset,
	)
	return nil
}
