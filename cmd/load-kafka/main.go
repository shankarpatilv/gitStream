package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	loadDotEnv(".env")
	cfg := loadConfig()
	if err := validateConfig(cfg); err != nil {
		slog.Error("invalid load configuration", "error", err)
		os.Exit(1)
	}

	producer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:  cfg.brokers,
		Topic:    cfg.topic,
		Username: cfg.username,
		Password: cfg.password,
	})
	if err != nil {
		slog.Error("could not open kafka producer", "error", err)
		os.Exit(1)
	}
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := publishSyntheticEvents(ctx, producer, cfg); err != nil {
		slog.Error("could not publish synthetic load", "error", err)
		os.Exit(1)
	}
}
