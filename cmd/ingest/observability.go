package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ingestEventsIngested = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gitstream_events_ingested_total",
			Help: "Total GitHub events accepted and published to Kafka by ingest.",
		},
	)
	ingestPollDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "gitstream_kafka_poll_duration_seconds",
			Help:    "Duration of one GitHub poll and publish cycle.",
			Buckets: prometheus.DefBuckets,
		},
	)
	ingestErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitstream_ingest_errors_total",
			Help: "Total ingest errors by operation.",
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(ingestEventsIngested, ingestPollDuration, ingestErrors)
	ingestErrors.WithLabelValues("github_poll").Add(0)
	ingestErrors.WithLabelValues("kafka_publish").Add(0)
}

func observePoll(start time.Time) {
	ingestPollDuration.Observe(time.Since(start).Seconds())
}
