package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/vivekspatil/gitstream/internal/storage"
)

const (
	defaultRecentLimit = 50
	maxRecentLimit     = 100
)

type recentEventsStore interface {
	RecentEvents(context.Context, string, int) ([]storage.RecentEvent, error)
}

func recentEventsHandler(store recentEventsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := strings.TrimSpace(r.URL.Query().Get("repo"))
		if repo == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "repo is required"})
			return
		}
		// Bound the result size because this endpoint reads row-oriented raw data.
		limit, ok := intQueryParam(w, r, "limit", defaultRecentLimit, 1, maxRecentLimit)
		if !ok {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		events, err := store.RecentEvents(ctx, repo, limit)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "postgres unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, recentEventsResponse{
			Repo:   repo,
			Limit:  limit,
			Events: events,
		})
	}
}

type recentEventsResponse struct {
	Repo   string                `json:"repo"`
	Limit  int                   `json:"limit"`
	Events []storage.RecentEvent `json:"events"`
}
