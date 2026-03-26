package main

import "github.com/prometheus/client_golang/prometheus"

var processorFailedEvents = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "gitstream_events_failed_total",
		Help: "Total processor events published to the DLQ after retries failed.",
	},
)

func init() {
	prometheus.MustRegister(processorFailedEvents)
}
