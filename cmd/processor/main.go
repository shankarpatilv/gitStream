package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vivekspatil/gitstream/internal/kafka"
)

func main() {
	const (
		service     = "processor"
		jobCapacity = 100
	)

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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	consumer, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers:       cfg.kafkaBrokers,
		Topic:         cfg.kafkaTopic,
		ConsumerGroup: cfg.consumerGroup,
		Username:      cfg.kafkaUsername,
		Password:      cfg.kafkaPassword,
	})
	if err != nil {
		slog.Error("could not create kafka consumer", "service", service, "error", err)
		os.Exit(1)
	}

	jobs := make(chan job, jobCapacity)
	startWorkers(ctx, cfg.workerCount, jobs)

	runConsumer(ctx, consumer, jobs)
	closeConsumer(consumer)
	slog.Info("service stopped", "service", service)
}
