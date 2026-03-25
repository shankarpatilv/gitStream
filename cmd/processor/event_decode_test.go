package main

import (
	"strings"
	"testing"
	"time"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

func TestDecodeMessageEventValid(t *testing.T) {
	message := kafka.Message{
		Value: []byte(`{
			"id": "processor-1",
			"type": "PushEvent",
			"repo": {"name": "owner/repo"},
			"actor": {"login": "octocat"},
			"created_at": "2026-05-06T21:30:00Z"
		}`),
	}

	event, err := decodeMessageEvent(message)
	if err != nil {
		t.Fatalf("decodeMessageEvent returned error: %v", err)
	}

	wantTime := time.Date(2026, 5, 6, 21, 30, 0, 0, time.UTC)
	if event.ID != "processor-1" || event.Type != "PushEvent" {
		t.Fatalf("unexpected id/type: %#v", event)
	}
	if event.RepoName != "owner/repo" || event.ActorName != "octocat" {
		t.Fatalf("unexpected repo/actor: %#v", event)
	}
	if !event.CreatedAt.Equal(wantTime) {
		t.Fatalf("created_at = %v, want %v", event.CreatedAt, wantTime)
	}
}

func TestDecodeMessageEventMalformedJSON(t *testing.T) {
	message := kafka.Message{Value: []byte(`{"id":`)}

	_, err := decodeMessageEvent(message)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "decode kafka event value") {
		t.Fatalf("error = %q, want decode context", err)
	}
}

func TestDecodeMessageEventMissingRequiredField(t *testing.T) {
	message := kafka.Message{
		Value: []byte(`{
			"id": "processor-1",
			"type": "PushEvent",
			"repo": {},
			"actor": {"login": "octocat"},
			"created_at": "2026-05-06T21:30:00Z"
		}`),
	}

	_, err := decodeMessageEvent(message)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "repo.name") {
		t.Fatalf("error = %q, want repo.name", err)
	}
}
