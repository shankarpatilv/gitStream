package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vivekspatil/gitstream/internal/kafka"
	"github.com/vivekspatil/gitstream/internal/storage"
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

	clickHouseStore, err := openClickHouseStore(consumerCtx, cfg)
	if err != nil {
		slog.Error("could not open clickhouse", "service", service, "error", err)
		os.Exit(1)
	}
	defer closeClickHouseStore(clickHouseStore)
	// ClickHouse prefers batches, but workers still wait for flush results.
	clickHouseSink := storage.NewClickHouseBatchSink(
		clickHouseStore,
		clickHouseBatchSize,
		clickHouseFlushInterval,
	)
	offsets := newOffsetTracker(consumer)

	jobs := make(chan job, jobCapacity)
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()
	workersDone := startWorkers(
		workerCtx,
		cfg.workerCount,
		jobs,
		dlqProducer,
		dualEventSink{postgres: postgresStore, clickHouse: clickHouseSink},
		offsets,
	)

	runConsumer(consumerCtx, consumer, jobs, dlqProducer, offsets)
	close(jobs)
	// Closing the ClickHouse sink after workers drain flushes the final partial batch.
	if waitForWorkers(workersDone, workerDrainTimeout, cancelWorkers) {
		closeClickHouseBatchSink(clickHouseSink)
	}
	closeConsumer(consumer)
	closeDLQProducer(dlqProducer)
	slog.Info("service stopped", "service", service)
}
