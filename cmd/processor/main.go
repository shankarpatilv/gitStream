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

	loadDotEnv(".env")
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("invalid processor configuration", "service", service, "error", err)
		os.Exit(1)
	}

	logStartupConfig(cfg)
	slog.Info("processor configuration validated", "service", service)

	consumerCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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

	postgresStore, err := openPostgresStore(consumerCtx, cfg)
	if err != nil {
		slog.Error("could not open postgres", "service", service, "error", err)
		os.Exit(1)
	}
	defer postgresStore.Close()

	jobs := make(chan job, jobCapacity)
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()
	workersDone := startWorkers(
		workerCtx,
		cfg.workerCount,
		jobs,
		dlqProducer,
		postgresStore,
	)

	runConsumer(consumerCtx, consumer, jobs)
	close(jobs)
	waitForWorkers(workersDone, workerDrainTimeout, cancelWorkers)
	closeConsumer(consumer)
	closeDLQProducer(dlqProducer)
	slog.Info("service stopped", "service", service)
}
