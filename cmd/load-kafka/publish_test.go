package main

import (
	"context"
	"testing"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

type capturePublisher struct {
	keys   [][]byte
	values [][]byte
}

func (p *capturePublisher) PublishRawBatch(_ context.Context, batch []kafka.RawMessage) error {
	for _, next := range batch {
		p.keys = append(p.keys, append([]byte(nil), next.Key...))
		p.values = append(p.values, append([]byte(nil), next.Value...))
	}
	return nil
}

func TestPublishSyntheticEventsPublishesConfiguredCount(t *testing.T) {
	publisher := &capturePublisher{}
	cfg := config{events: 3, batchSize: 2, prefix: "load", topic: "github-events"}

	if err := publishSyntheticEvents(context.Background(), publisher, cfg); err != nil {
		t.Fatalf("publishSyntheticEvents returned error: %v", err)
	}
	if len(publisher.values) != 3 {
		t.Fatalf("published %d events, want 3", len(publisher.values))
	}
	if string(publisher.keys[0]) != "load/repo-000" {
		t.Fatalf("first key = %q, want load/repo-000", publisher.keys[0])
	}
}
