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
		"kafka_dlq_topic", cfg.dlqTopic,
		"consumer_group", cfg.consumerGroup,
		"worker_count", cfg.workerCount,
	)
	slog.Info("processor configuration validated", "service", service)

	consumerCtx, stop := signal.NotifyContext(
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

	dlqProducer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:  cfg.kafkaBrokers,
		Topic:    cfg.dlqTopic,
		Username: cfg.kafkaUsername,
		Password: cfg.kafkaPassword,
	})
	if err != nil {
		slog.Error("could not create dlq producer", "service", service, "error", err)
		os.Exit(1)
	}

	jobs := make(chan job, jobCapacity)
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()
	workersDone := startWorkers(workerCtx, cfg.workerCount, jobs, dlqProducer)

	runConsumer(consumerCtx, consumer, jobs)
	close(jobs)
	waitForWorkers(workersDone, workerDrainTimeout, cancelWorkers)
	closeConsumer(consumer)
	closeDLQProducer(dlqProducer)
	slog.Info("service stopped", "service", service)
}
