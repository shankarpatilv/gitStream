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

	handleJob(
		context.Background(),
		1,
		next,
		dlq,
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
		func(context.Context, int, job) error {
			return nil
		},
		recordDelays(&[]time.Duration{}),
	)

	if dlq.writes != 0 {
		t.Fatalf("dlq writes = %d, want 0", dlq.writes)
	}
}
