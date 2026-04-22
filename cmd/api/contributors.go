package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/vivekspatil/gitstream/internal/storage"
)

const (
	defaultContributorsLimit = 10
	maxContributorsLimit     = 100
)

type contributorsStore interface {
	TopContributors(context.Context, string, int) ([]storage.TopContributor, error)
}

func contributorsHandler(store contributorsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := strings.TrimSpace(r.URL.Query().Get("repo"))
		if repo == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "repo is required"})
			return
		}
		// Bound the grouped Postgres scan to keep response size predictable.
		limit, ok := intQueryParam(w, r, "limit", defaultContributorsLimit, 1, maxContributorsLimit)
		if !ok {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		contributors, err := store.TopContributors(ctx, repo, limit)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "postgres unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, contributorsResponse{
			Repo:         repo,
			Limit:        limit,
			Contributors: contributors,
		})
	}
}

type contributorsResponse struct {
	Repo         string                   `json:"repo"`
	Limit        int                      `json:"limit"`
	Contributors []storage.TopContributor `json:"contributors"`
}
