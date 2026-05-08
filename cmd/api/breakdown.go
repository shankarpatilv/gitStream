package main

import (
	"context"
	"net/http"
	"time"

	"github.com/vivekspatil/gitstream/internal/storage"
)

const (
	defaultBreakdownHours = 24
	maxBreakdownHours     = 168
)

type breakdownStore interface {
	EventBreakdown(context.Context, int) ([]storage.EventBreakdown, error)
}

func breakdownHandler(store breakdownStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Bound the time window so one request cannot trigger an unbounded scan.
		hours, ok := intQueryParam(w, r, "hours", defaultBreakdownHours, 1, maxBreakdownHours)
		if !ok {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		breakdown, err := store.EventBreakdown(ctx, hours)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "clickhouse unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, breakdownResponse{
			Hours:     hours,
			Breakdown: breakdown,
		})
	}
}

type breakdownResponse struct {
	Hours     int                      `json:"hours"`
	Breakdown []storage.EventBreakdown `json:"breakdown"`
}
