package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/vivekspatil/gitstream/internal/storage"
)

const (
	defaultTrendingHours = 1
	defaultTrendingLimit = 10
	maxTrendingHours     = 168
	maxTrendingLimit     = 100
)

type trendingStore interface {
	TrendingRepos(context.Context, int, int) ([]storage.TrendingRepo, error)
}

func trendingHandler(store trendingStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Keep API limits explicit so one request cannot trigger a huge scan.
		hours, ok := intQueryParam(w, r, "hours", defaultTrendingHours, 1, maxTrendingHours)
		if !ok {
			return
		}
		limit, ok := intQueryParam(w, r, "limit", defaultTrendingLimit, 1, maxTrendingLimit)
		if !ok {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		repos, err := store.TrendingRepos(ctx, hours, limit)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "clickhouse unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, trendingResponse{
			Hours: hours,
			Limit: limit,
			Repos: repos,
		})
	}
}

type trendingResponse struct {
	Hours int                    `json:"hours"`
	Limit int                    `json:"limit"`
	Repos []storage.TrendingRepo `json:"repos"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// intQueryParam parses a bounded positive integer query parameter.
func intQueryParam(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	defaultValue int,
	minValue int,
	maxValue int,
) (int, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return defaultValue, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < minValue || value > maxValue {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid " + name})
		return 0, false
	}
	return value, true
}
