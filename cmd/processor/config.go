package main

import (
	"fmt"
)

const (
	defaultKafkaTopic    = "github-events"
	defaultDLQTopic      = "github-events-dlq"
	defaultConsumerGroup = "gitstream-processors"
	defaultWorkerCount   = 10
	defaultPostgresHost  = "localhost"
	defaultPostgresPort  = "5432"
	defaultPostgresDB    = "gitstream"
	defaultPostgresUser  = "gitstream"
)

type config struct {
	kafkaBrokers  []string
	kafkaTopic    string
	dlqTopic      string
	consumerGroup string
	kafkaUsername string
	kafkaPassword string
	workerCount   int
	postgresHost  string
	postgresPort  string
	postgresDB    string
	postgresUser  string
	postgresPass  string
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
		postgresHost:  envOrDefault("POSTGRES_HOST", defaultPostgresHost),
		postgresPort:  envOrDefault("POSTGRES_PORT", defaultPostgresPort),
		postgresDB:    envOrDefault("POSTGRES_DB", defaultPostgresDB),
		postgresUser:  envOrDefault("POSTGRES_USER", defaultPostgresUser),
		postgresPass:  envOrDefault("POSTGRES_PASSWORD", ""),
	}
	if len(cfg.kafkaBrokers) == 0 {
		return config{}, fmt.Errorf("KAFKA_BROKERS is required")
	}

	required := []requiredConfig{
		{name: "KAFKA_TOPIC", value: cfg.kafkaTopic},
		{name: "KAFKA_DLQ_TOPIC", value: cfg.dlqTopic},
		{name: "KAFKA_CONSUMER_GROUP", value: cfg.consumerGroup},
		{name: "POSTGRES_HOST", value: cfg.postgresHost},
		{name: "POSTGRES_PORT", value: cfg.postgresPort},
		{name: "POSTGRES_DB", value: cfg.postgresDB},
		{name: "POSTGRES_USER", value: cfg.postgresUser},
	}
	if err := validateRequiredConfig(required); err != nil {
		return config{}, err
	}
	return cfg, nil
}

type requiredConfig struct {
	name  string
	value string
}

func validateRequiredConfig(values []requiredConfig) error {
	for _, next := range values {
		if next.value == "" {
			return fmt.Errorf("%s is required", next.name)
		}
	}
	return nil
}
