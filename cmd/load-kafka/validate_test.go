package main

import "testing"

func TestValidateConfigRequiresBroker(t *testing.T) {
	cfg := config{events: 1, batchSize: 1, prefix: "load", topic: "github-events"}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig returned nil, want broker error")
	}
}

func TestValidateConfigAcceptsValidConfig(t *testing.T) {
	cfg := config{
		events:    1,
		batchSize: 1,
		prefix:    "load",
		topic:     "github-events",
		brokers:   []string{"localhost:9092"},
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig returned error: %v", err)
	}
}
