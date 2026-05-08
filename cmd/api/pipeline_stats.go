package main

import (
	"sync/atomic"
	"time"
)

type pipelineStats struct {
	startedAt time.Time
	ingested  atomic.Uint64
	processed atomic.Uint64
	failed    atomic.Uint64
}

func newPipelineStats(startedAt time.Time) *pipelineStats {
	return &pipelineStats{startedAt: startedAt.UTC()}
}

// Snapshot reads the in-memory counters without mutating pipeline state.
func (s *pipelineStats) Snapshot(now time.Time) pipelineStatsSnapshot {
	uptime := int64(now.UTC().Sub(s.startedAt).Seconds())
	if uptime < 0 {
		uptime = 0
	}
	return pipelineStatsSnapshot{
		Status:          "ok",
		StartedAt:       s.startedAt,
		UptimeSeconds:   uptime,
		EventsIngested:  s.ingested.Load(),
		EventsProcessed: s.processed.Load(),
		EventsFailed:    s.failed.Load(),
	}
}

type pipelineStatsSnapshot struct {
	Status          string    `json:"status"`
	StartedAt       time.Time `json:"started_at"`
	UptimeSeconds   int64     `json:"uptime_seconds"`
	EventsIngested  uint64    `json:"events_ingested_total"`
	EventsProcessed uint64    `json:"events_processed_total"`
	EventsFailed    uint64    `json:"events_failed_total"`
}
