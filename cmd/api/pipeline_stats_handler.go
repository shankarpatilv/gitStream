package main

import (
	"net/http"
	"time"
)

type pipelineStatsReader interface {
	Snapshot(time.Time) pipelineStatsSnapshot
}

func pipelineStatsHandler(stats pipelineStatsReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, stats.Snapshot(time.Now()))
	}
}
