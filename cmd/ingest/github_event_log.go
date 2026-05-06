package main

import (
	"encoding/json"
	"log/slog"

	"github.com/vivekspatil/gitstream/internal/events"
)

// logAcceptedGitHubEvent skips invalid, unsupported, or duplicate events.
func logAcceptedGitHubEvent(rawEvent json.RawMessage, deduper *events.Deduper) {
	event, err := events.ParseGitHubEvent(rawEvent)
	if err != nil {
		slog.Warn("skipping invalid github event", "error", err)
		return
	}
	if !events.IsAllowedEventType(event.Type) || !deduper.IsNew(event.ID) {
		return
	}

	slog.Info(
		"accepted github event",
		"repo", event.RepoName,
		"type", event.Type,
		"actor", event.ActorName,
	)
}
