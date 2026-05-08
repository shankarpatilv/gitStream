package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

const githubEventsURL = "https://api.github.com/events"

// runGitHubPoller polls once on startup, then repeats on the configured interval.
func runGitHubPoller(
	ctx context.Context,
	client *http.Client,
	token string,
	deduper *events.Deduper,
	publisher eventPublisher,
	interval time.Duration,
) {
	pollAndPublishGitHubEvents(ctx, client, token, deduper, publisher)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("github poller stopped", "error", ctx.Err())
			return
		case <-ticker.C:
			pollAndPublishGitHubEvents(ctx, client, token, deduper, publisher)
		}
	}
}

// pollAndPublishGitHubEvents handles one GitHub API poll cycle.
func pollAndPublishGitHubEvents(
	ctx context.Context,
	client *http.Client,
	token string,
	deduper *events.Deduper,
	publisher eventPublisher,
) {
	defer observePoll(time.Now())

	rawEvents, err := fetchGitHubEvents(ctx, client, token)
	if err != nil {
		ingestErrors.WithLabelValues("github_poll").Inc()
		slog.Error("github events poll failed", "error", err)
		return
	}

	for _, rawEvent := range rawEvents {
		handleAcceptedGitHubEvent(ctx, rawEvent, deduper, publisher)
	}
}
