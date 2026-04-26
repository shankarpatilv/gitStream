package main

import "github.com/prometheus/client_golang/prometheus"

var (
	processorProcessedEvents = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gitstream_events_processed_total",
			Help: "Total events durably processed by processor workers.",
		},
	)
	processorFailedEvents = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gitstream_events_failed_total",
			Help: "Total processor events published to the DLQ after retries failed.",
		},
	)
	processorConsumerLag = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gitstream_consumer_lag",
			Help: "Fetched Kafka messages not yet durably completed by this processor.",
		},
	)
	processorProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "gitstream_processing_duration_seconds",
			Help:    "Duration of processor job handling including retry and DLQ routing.",
			Buckets: prometheus.DefBuckets,
		},
	)
	processorPostgresWriteDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "gitstream_postgres_write_duration_seconds",
			Help:    "Duration of Postgres raw event writes.",
			Buckets: prometheus.DefBuckets,
		},
	)
	processorClickHouseWriteDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "gitstream_clickhouse_write_duration_seconds",
			Help:    "Duration of ClickHouse analytics writes.",
			Buckets: prometheus.DefBuckets,
		},
	)
	processorDLQDepth = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gitstream_dlq_depth",
			Help: "DLQ messages published by this processor process.",
		},
	)
	processorActiveWorkers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gitstream_active_workers",
			Help: "Processor worker goroutines currently running.",
		},
	)
	processorRetries = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gitstream_processor_retries_total",
			Help: "Total processor retry attempts after failed job attempts.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		processorProcessedEvents,
		processorFailedEvents,
		processorConsumerLag,
		processorProcessingDuration,
		processorPostgresWriteDuration,
		processorClickHouseWriteDuration,
		processorDLQDepth,
		processorActiveWorkers,
		processorRetries,
	)
}
