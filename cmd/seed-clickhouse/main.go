package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/vivekspatil/gitstream/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	loadDotEnv(".env")
	cfg := loadConfig()
	if err := validateConfig(cfg); err != nil {
		slog.Error("invalid seed configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	store, err := storage.NewClickHouseStore(ctx, cfg.clickhouse)
	if err != nil {
		slog.Error("could not open clickhouse", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	if err := store.EnsureSchema(ctx); err != nil {
		slog.Error("could not ensure schema", "error", err)
		os.Exit(1)
	}
	if err := seedClickHouse(ctx, store, cfg); err != nil {
		slog.Error("could not seed clickhouse", "error", err)
		os.Exit(1)
	}
	slog.Info("seed complete", "rows", cfg.rows, "prefix", cfg.prefix)
}
