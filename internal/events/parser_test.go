package events

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestParseGitHubEventValid(t *testing.T) {
	data := []byte(`{
		"id": "123",
		"type": "PushEvent",
		"repo": {"name": "owner/repo"},
		"actor": {"login": "octocat"},
		"created_at": "2026-05-05T01:02:03Z"
	}`)

	event, err := ParseGitHubEvent(data)
	if err != nil {
		t.Fatalf("ParseGitHubEvent returned error: %v", err)
	}

	wantTime := time.Date(2026, 5, 5, 1, 2, 3, 0, time.UTC)
	if event.ID != "123" || event.Type != "PushEvent" {
		t.Fatalf("unexpected id/type: %#v", event)
	}
	if event.RepoName != "owner/repo" || event.ActorName != "octocat" {
		t.Fatalf("unexpected repo/actor: %#v", event)
	}
	if !event.CreatedAt.Equal(wantTime) {
		t.Fatalf("created_at = %v, want %v", event.CreatedAt, wantTime)
	}
	if !bytes.Equal(event.Payload, data) {
		t.Fatalf("payload was not preserved")
	}
}

func TestParseGitHubEventMissingRequiredField(t *testing.T) {
	data := []byte(`{
		"id": "123",
		"type": "PushEvent",
		"repo": {},
		"actor": {"login": "octocat"},
		"created_at": "2026-05-05T01:02:03Z"
	}`)

	_, err := ParseGitHubEvent(data)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "repo.name") {
		t.Fatalf("error = %q, want repo.name", err)
	}
}

func TestParseGitHubEventInvalidCreatedAt(t *testing.T) {
	data := []byte(`{
		"id": "123",
		"type": "PushEvent",
		"repo": {"name": "owner/repo"},
		"actor": {"login": "octocat"},
		"created_at": "not-a-time"
	}`)

	_, err := ParseGitHubEvent(data)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "created_at") {
		t.Fatalf("error = %q, want created_at", err)
	}
}
