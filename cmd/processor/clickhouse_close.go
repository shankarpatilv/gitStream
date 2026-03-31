package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/vivekspatil/gitstream/internal/storage"
)

func closeClickHouseBatchSink(sink *storage.ClickHouseBatchSink) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sink.Close(ctx); err != nil {
		slog.Error("clickhouse batch sink close failed", "error", err)
		return
	}
	slog.Info("clickhouse batch sink closed")
}

func closeClickHouseStore(store *storage.ClickHouseStore) {
	if err := store.Close(); err != nil {
		slog.Error("clickhouse close failed", "error", err)
		return
	}
	slog.Info("clickhouse closed")
}
