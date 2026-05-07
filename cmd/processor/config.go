package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultKafkaTopic    = "github-events"
	defaultConsumerGroup = "gitstream-processors"
	defaultWorkerCount   = 10
)

type config struct {
	kafkaBrokers  []string
	kafkaTopic    string
	consumerGroup string
	workerCount   int
}

func loadConfig() (config, error) {
	workerCount, err := workerCountFromEnv()
	if err != nil {
		return config{}, err
	}

	cfg := config{
		kafkaBrokers:  envList("KAFKA_BROKERS"),
		kafkaTopic:    envOrDefault("KAFKA_TOPIC", defaultKafkaTopic),
		consumerGroup: envOrDefault("KAFKA_CONSUMER_GROUP", defaultConsumerGroup),
		workerCount:   workerCount,
	}
	if len(cfg.kafkaBrokers) == 0 {
		return config{}, fmt.Errorf("KAFKA_BROKERS is required")
	}
	if cfg.kafkaTopic == "" {
		return config{}, fmt.Errorf("KAFKA_TOPIC is required")
	}
	if cfg.consumerGroup == "" {
		return config{}, fmt.Errorf("KAFKA_CONSUMER_GROUP is required")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envList(key string) []string {
	parts := strings.Split(os.Getenv(key), ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func workerCountFromEnv() (int, error) {
	value := strings.TrimSpace(os.Getenv("WORKER_COUNT"))
	if value == "" {
		return defaultWorkerCount, nil
	}
	workerCount, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("WORKER_COUNT must be an integer: %w", err)
	}
	if workerCount < 1 {
		return 0, fmt.Errorf("WORKER_COUNT must be at least 1")
	}
	return workerCount, nil
}
