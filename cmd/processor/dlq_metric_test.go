package main

import (
	"context"
	"errors"
	"testing"

	dto "github.com/prometheus/client_model/go"
)

func TestPublishFailedJobUpdatesMetricAfterDLQSuccess(t *testing.T) {
	before := failedEventsValue(t)

	publishFailedJob(context.Background(), 1, retryTestJob("event-metric"), &fakeDLQPublisher{})

	after := failedEventsValue(t)
	if after != before+1 {
		t.Fatalf("failed events metric = %v, want %v", after, before+1)
	}
}

func TestPublishFailedJobDoesNotUpdateMetricAfterDLQFailure(t *testing.T) {
	before := failedEventsValue(t)
	dlq := &fakeDLQPublisher{failErr: errors.New("kafka unavailable")}

	publishFailedJob(context.Background(), 1, retryTestJob("event-metric"), dlq)

	after := failedEventsValue(t)
	if after != before {
		t.Fatalf("failed events metric = %v, want %v", after, before)
	}
}

func failedEventsValue(t *testing.T) float64 {
	t.Helper()

	metric := &dto.Metric{}
	if err := processorFailedEvents.Write(metric); err != nil {
		t.Fatalf("could not read failed events metric: %v", err)
	}
	return metric.GetCounter().GetValue()
}
