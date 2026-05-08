package main

import "testing"

func TestSyntheticBatchUsesRequestedRange(t *testing.T) {
	cfg := config{prefix: "test"}
	batch := syntheticBatch(cfg, 10, 13)

	if len(batch) != 3 {
		t.Fatalf("len(batch) = %d, want 3", len(batch))
	}
	if batch[0].ID != "test-000010" || batch[2].ID != "test-000012" {
		t.Fatalf("unexpected ids: %q %q", batch[0].ID, batch[2].ID)
	}
}

func TestSyntheticEventUsesSeedPrefix(t *testing.T) {
	event := syntheticEvent("bench", 7)

	if event.RepoName != "bench/repo-007" {
		t.Fatalf("RepoName = %q, want bench/repo-007", event.RepoName)
	}
	if event.ActorName != "bench-actor-007" {
		t.Fatalf("ActorName = %q, want bench-actor-007", event.ActorName)
	}
}
