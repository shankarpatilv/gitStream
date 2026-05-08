package main

import (
	"context"
	"errors"
	"testing"
)

func TestDualEventSinkWritesBothStores(t *testing.T) {
	postgres := &fakeEventSink{}
	clickHouse := &fakeEventSink{}
	sink := dualEventSink{postgres: postgres, clickHouse: clickHouse}

	err := sink.InsertEvent(context.Background(), retryTestJob("dual-sink").event)
	if err != nil {
		t.Fatalf("InsertEvent returned error: %v", err)
	}
	if postgres.writes != 1 || clickHouse.writes != 1 {
		t.Fatalf("writes = postgres %d clickhouse %d, want 1 each", postgres.writes, clickHouse.writes)
	}
}

func TestDualEventSinkSkipsClickHouseAfterPostgresFailure(t *testing.T) {
	postgres := &fakeEventSink{failErr: errors.New("postgres unavailable")}
	clickHouse := &fakeEventSink{}
	sink := dualEventSink{postgres: postgres, clickHouse: clickHouse}

	err := sink.InsertEvent(context.Background(), retryTestJob("dual-sink").event)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if clickHouse.writes != 0 {
		t.Fatalf("clickhouse writes = %d, want 0", clickHouse.writes)
	}
}
