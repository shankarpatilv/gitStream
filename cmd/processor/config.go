package main

import "fmt"

// loadConfig validates the processor settings needed before Kafka consumption.
func loadConfig() (config, error) {
	workerCount, err := workerCountFromEnv()
	if err != nil {
		return config{}, err
	}

	cfg := config{
		kafkaBrokers:   envList("KAFKA_BROKERS"),
		kafkaTopic:     envOrDefault("KAFKA_TOPIC", defaultKafkaTopic),
		dlqTopic:       envOrDefault("KAFKA_DLQ_TOPIC", defaultDLQTopic),
		consumerGroup:  envOrDefault("KAFKA_CONSUMER_GROUP", defaultConsumerGroup),
		kafkaUsername:  envOrDefault("KAFKA_USERNAME", ""),
		kafkaPassword:  envOrDefault("KAFKA_PASSWORD", ""),
		workerCount:    workerCount,
		postgresHost:   envOrDefault("POSTGRES_HOST", defaultPostgresHost),
		postgresPort:   envOrDefault("POSTGRES_PORT", defaultPostgresPort),
		postgresDB:     envOrDefault("POSTGRES_DB", defaultPostgresDB),
		postgresUser:   envOrDefault("POSTGRES_USER", defaultPostgresUser),
		postgresPass:   envOrDefault("POSTGRES_PASSWORD", ""),
		clickHouseHost: envOrDefault("CLICKHOUSE_HOST", defaultClickHouseHost),
		clickHousePort: envOrDefault("CLICKHOUSE_NATIVE_PORT", defaultClickHousePort),
		clickHouseDB:   envOrDefault("CLICKHOUSE_DB", defaultClickHouseDB),
		clickHouseUser: envOrDefault("CLICKHOUSE_USER", defaultClickHouseUser),
		clickHousePass: envOrDefault("CLICKHOUSE_PASSWORD", ""),
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
		{name: "CLICKHOUSE_HOST", value: cfg.clickHouseHost},
		{name: "CLICKHOUSE_NATIVE_PORT", value: cfg.clickHousePort},
		{name: "CLICKHOUSE_DB", value: cfg.clickHouseDB},
		{name: "CLICKHOUSE_USER", value: cfg.clickHouseUser},
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
