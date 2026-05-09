package main

import (
	"testing"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

func TestSyntheticKafkaEventParsesAsGitHubEvent(t *testing.T) {
	key, value, err := syntheticKafkaEvent("bench", 7, time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("syntheticKafkaEvent returned error: %v", err)
	}
	if string(key) != "bench/repo-007" {
		t.Fatalf("key = %q, want bench/repo-007", key)
	}

	event, err := events.ParseGitHubEvent(value)
	if err != nil {
		t.Fatalf("ParseGitHubEvent returned error: %v", err)
	}
	if event.ID != "bench-000007" || event.RepoName != string(key) {
		t.Fatalf("unexpected event: %#v", event)
	}
}
