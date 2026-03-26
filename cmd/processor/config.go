package main

import (
	"fmt"
)

const (
	defaultKafkaTopic    = "github-events"
	defaultDLQTopic      = "github-events-dlq"
	defaultConsumerGroup = "gitstream-processors"
	defaultWorkerCount   = 10
)

type config struct {
	kafkaBrokers  []string
	kafkaTopic    string
	dlqTopic      string
	consumerGroup string
	kafkaUsername string
	kafkaPassword string
	workerCount   int
}

// loadConfig validates the processor settings needed before Kafka consumption.
func loadConfig() (config, error) {
	workerCount, err := workerCountFromEnv()
	if err != nil {
		return config{}, err
	}

	cfg := config{
		kafkaBrokers:  envList("KAFKA_BROKERS"),
		kafkaTopic:    envOrDefault("KAFKA_TOPIC", defaultKafkaTopic),
		dlqTopic:      envOrDefault("KAFKA_DLQ_TOPIC", defaultDLQTopic),
		consumerGroup: envOrDefault("KAFKA_CONSUMER_GROUP", defaultConsumerGroup),
		kafkaUsername: envOrDefault("KAFKA_USERNAME", ""),
		kafkaPassword: envOrDefault("KAFKA_PASSWORD", ""),
		workerCount:   workerCount,
	}
	if len(cfg.kafkaBrokers) == 0 {
		return config{}, fmt.Errorf("KAFKA_BROKERS is required")
	}
	if cfg.kafkaTopic == "" {
		return config{}, fmt.Errorf("KAFKA_TOPIC is required")
	}
	if cfg.dlqTopic == "" {
		return config{}, fmt.Errorf("KAFKA_DLQ_TOPIC is required")
	}
	if cfg.consumerGroup == "" {
		return config{}, fmt.Errorf("KAFKA_CONSUMER_GROUP is required")
	}
	return cfg, nil
}
