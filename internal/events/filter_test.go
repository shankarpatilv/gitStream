package events

import "testing"

func TestIsAllowedEventTypeAcceptsLockedTypes(t *testing.T) {
	allowed := []string{
		"PushEvent",
		"PullRequestEvent",
		"IssuesEvent",
		"WatchEvent",
		"ForkEvent",
	}

	for _, eventType := range allowed {
		if !IsAllowedEventType(eventType) {
			t.Fatalf("%s should be allowed", eventType)
		}
	}
}

func TestIsAllowedEventTypeRejectsUnknownType(t *testing.T) {
	if IsAllowedEventType("CreateEvent") {
		t.Fatal("CreateEvent should not be allowed")
	}
}

func TestFilterAllowedEventsKeepsOnlyAllowedTypes(t *testing.T) {
	events := []GitHubEvent{
		{ID: "1", Type: "PushEvent"},
		{ID: "2", Type: "CreateEvent"},
		{ID: "3", Type: "WatchEvent"},
	}

	filtered := FilterAllowedEvents(events)
	if len(filtered) != 2 {
		t.Fatalf("len(filtered) = %d, want 2", len(filtered))
	}
	if filtered[0].ID != "1" || filtered[1].ID != "3" {
		t.Fatalf("filtered events = %#v", filtered)
	}
}
