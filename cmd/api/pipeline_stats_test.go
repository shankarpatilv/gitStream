package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPipelineStatsSnapshotUsesCountersAndUptime(t *testing.T) {
	startedAt := time.Date(2026, 5, 8, 20, 0, 0, 0, time.UTC)
	stats := newPipelineStats(startedAt)
	stats.ingested.Add(3)
	stats.processed.Add(2)
	stats.failed.Add(1)

	snapshot := stats.Snapshot(startedAt.Add(10 * time.Second))

	if snapshot.Status != "ok" || snapshot.UptimeSeconds != 10 {
		t.Fatalf("snapshot status/uptime = %q/%d", snapshot.Status, snapshot.UptimeSeconds)
	}
	if snapshot.EventsIngested != 3 || snapshot.EventsProcessed != 2 || snapshot.EventsFailed != 1 {
		t.Fatalf("snapshot counters = %+v", snapshot)
	}
}

func TestPipelineStatsHandlerReturnsStableShape(t *testing.T) {
	startedAt := time.Date(2026, 5, 8, 20, 0, 0, 0, time.UTC)
	recorder := httptest.NewRecorder()

	pipelineStatsHandler(newPipelineStats(startedAt)).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/stats/pipeline", nil),
	)

	body := recorder.Body.String()
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	for _, field := range []string{
		`"status":"ok"`,
		`"started_at":`,
		`"uptime_seconds":`,
		`"events_ingested_total":0`,
		`"events_processed_total":0`,
		`"events_failed_total":0`,
	} {
		if !strings.Contains(body, field) {
			t.Fatalf("expected response to contain %s, got %s", field, body)
		}
	}
}
