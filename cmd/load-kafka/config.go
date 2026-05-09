package main

import (
	"flag"
	"os"
	"strings"
)

const (
	defaultEvents    = 1000
	defaultBatchSize = 100
	defaultPrefix    = "load"
)

type config struct {
	events    int
	batchSize int
	prefix    string
	topic     string
	brokers   []string
	username  string
	password  string
}

func loadConfig() config {
	events := flag.Int("events", defaultEvents, "synthetic Kafka events to publish")
	batchSize := flag.Int("batch-size", defaultBatchSize, "Kafka messages per producer write")
	prefix := flag.String("prefix", defaultPrefix, "synthetic event id/repo prefix")
	flag.Parse()

	return config{
		events:    *events,
		batchSize: *batchSize,
		prefix:    strings.TrimSpace(*prefix),
		topic:     envOrDefault("KAFKA_TOPIC", "github-events"),
		brokers:   envList("KAFKA_BROKERS"),
		username:  os.Getenv("KAFKA_USERNAME"),
		password:  os.Getenv("KAFKA_PASSWORD"),
	}
}

func envList(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
