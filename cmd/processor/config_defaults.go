package main

const (
	defaultKafkaTopic     = "github-events"
	defaultDLQTopic       = "github-events-dlq"
	defaultConsumerGroup  = "gitstream-processors"
	defaultMetricsPort    = "8091"
	defaultWorkerCount    = 10
	defaultPostgresHost   = "localhost"
	defaultPostgresPort   = "5432"
	defaultPostgresDB     = "gitstream"
	defaultPostgresUser   = "gitstream"
	defaultClickHouseHost = "localhost"
	defaultClickHousePort = "9000"
	defaultClickHouseDB   = "gitstream"
	defaultClickHouseUser = "gitstream"
)
