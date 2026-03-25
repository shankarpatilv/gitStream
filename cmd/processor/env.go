package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// envOrDefault reads an env var while keeping local defaults explicit.
func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

// envList converts comma-separated broker config into a clean string slice.
func envList(key string) []string {
	parts := strings.Split(os.Getenv(key), ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

// workerCountFromEnv validates worker configuration before workers exist.
func workerCountFromEnv() (int, error) {
	value := strings.TrimSpace(os.Getenv("WORKER_COUNT"))
	if value == "" {
		return defaultWorkerCount, nil
	}
	workerCount, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("WORKER_COUNT must be an integer: %w", err)
	}
	if workerCount < 1 {
		return 0, fmt.Errorf("WORKER_COUNT must be at least 1")
	}
	return workerCount, nil
}
