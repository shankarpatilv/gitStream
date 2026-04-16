package main

import (
	"flag"
	"os"
	"strings"

	"github.com/vivekspatil/gitstream/internal/storage"
)

const (
	defaultRows      = 100000
	defaultBatchSize = 1000
	defaultPrefix    = "seed"
)

type config struct {
	rows       int
	batchSize  int
	prefix     string
	clickhouse storage.ClickHouseConfig
}

func loadConfig() config {
	rows := flag.Int("rows", defaultRows, "synthetic event rows to insert")
	batchSize := flag.Int("batch-size", defaultBatchSize, "ClickHouse insert batch size")
	prefix := flag.String("prefix", defaultPrefix, "synthetic repo/id prefix")
	flag.Parse()

	return config{
		rows:      *rows,
		batchSize: *batchSize,
		prefix:    strings.TrimSpace(*prefix),
		clickhouse: storage.ClickHouseConfig{
			Host:     envOrDefault("CLICKHOUSE_HOST", "localhost"),
			Port:     envOrDefault("CLICKHOUSE_NATIVE_PORT", "9000"),
			Database: envOrDefault("CLICKHOUSE_DB", "gitstream"),
			User:     envOrDefault("CLICKHOUSE_USER", "gitstream"),
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
