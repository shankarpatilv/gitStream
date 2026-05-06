package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/vivekspatil/gitstream/internal/events"
)

type eventPublisher interface {
	Publish(context.Context, events.GitHubEvent) error
}

// handleAcceptedGitHubEvent skips invalid events and publishes accepted ones.
func handleAcceptedGitHubEvent(
	ctx context.Context,
	rawEvent json.RawMessage,
	deduper *events.Deduper,
	publisher eventPublisher,
) {
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

	if err := publisher.Publish(ctx, event); err != nil {
		slog.Error(
			"kafka publish failed",
			"repo", event.RepoName,
			"type", event.Type,
			"actor", event.ActorName,
			"error", err,
		)
	}
}
