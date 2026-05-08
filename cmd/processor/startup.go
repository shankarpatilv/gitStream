package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vivekspatil/gitstream/internal/storage"
)

// logStartupConfig logs non-secret configuration for operational debugging.
func logStartupConfig(cfg config) {
	slog.Info(
		"starting service",
		"service", service,
		"kafka_brokers", cfg.kafkaBrokers,
		"kafka_topic", cfg.kafkaTopic,
		"kafka_dlq_topic", cfg.dlqTopic,
		"consumer_group", cfg.consumerGroup,
		"worker_count", cfg.workerCount,
		"postgres_host", cfg.postgresHost,
		"postgres_port", cfg.postgresPort,
		"postgres_db", cfg.postgresDB,
		"postgres_user", cfg.postgresUser,
		"clickhouse_host", cfg.clickHouseHost,
		"clickhouse_port", cfg.clickHousePort,
		"clickhouse_db", cfg.clickHouseDB,
		"clickhouse_user", cfg.clickHouseUser,
	)
}

// openPostgresStore connects to Postgres and prepares schema before workers run.
func openPostgresStore(ctx context.Context, cfg config) (*storage.PostgresStore, error) {
	store, err := storage.NewPostgresStore(ctx, storage.PostgresConfig{
		Host:     cfg.postgresHost,
		Port:     cfg.postgresPort,
		Database: cfg.postgresDB,
		User:     cfg.postgresUser,
		Password: cfg.postgresPass,
	})
	if err != nil {
		return nil, err
	}
	if err := store.EnsureSchema(ctx); err != nil {
		store.Close()
		return nil, fmt.Errorf("ensure postgres schema: %w", err)
	}
	return store, nil
}

// openClickHouseStore connects to ClickHouse and prepares analytics tables.
func openClickHouseStore(ctx context.Context, cfg config) (*storage.ClickHouseStore, error) {
	store, err := storage.NewClickHouseStore(ctx, storage.ClickHouseConfig{
		Host:     cfg.clickHouseHost,
		Port:     cfg.clickHousePort,
		Database: cfg.clickHouseDB,
		User:     cfg.clickHouseUser,
		Password: cfg.clickHousePass,
	})
	if err != nil {
		return nil, err
	}
	if err := store.EnsureSchema(ctx); err != nil {
		store.Close()
		return nil, fmt.Errorf("ensure clickhouse schema: %w", err)
	}
	return store, nil
}
