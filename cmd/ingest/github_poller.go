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
	interval time.Duration,
) {
	pollAndLogGitHubEvents(ctx, client, token, deduper)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("github poller stopped", "error", ctx.Err())
			return
		case <-ticker.C:
			pollAndLogGitHubEvents(ctx, client, token, deduper)
		}
	}
}

// pollAndLogGitHubEvents handles one GitHub API poll cycle.
func pollAndLogGitHubEvents(
	ctx context.Context,
	client *http.Client,
	token string,
	deduper *events.Deduper,
) {
	rawEvents, err := fetchGitHubEvents(ctx, client, token)
	if err != nil {
		slog.Error("github events poll failed", "error", err)
		return
	}

	for _, rawEvent := range rawEvents {
		logAcceptedGitHubEvent(rawEvent, deduper)
	}
}
