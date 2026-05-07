package main

import (
	"log/slog"
	"os"
)

func main() {
	const service = "processor"

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := loadConfig()
	if err != nil {
		slog.Error("invalid processor configuration", "service", service, "error", err)
		os.Exit(1)
	}

	slog.Info(
		"starting service",
		"service", service,
		"kafka_brokers", cfg.kafkaBrokers,
		"kafka_topic", cfg.kafkaTopic,
		"consumer_group", cfg.consumerGroup,
		"worker_count", cfg.workerCount,
	)
	slog.Info("processor configuration validated", "service", service)
}
