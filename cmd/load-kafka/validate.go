package main

import "fmt"

func validateConfig(cfg config) error {
	if cfg.events < 1 {
		return fmt.Errorf("events must be at least 1")
	}
	if cfg.batchSize < 1 {
		return fmt.Errorf("batch-size must be at least 1")
	}
	if cfg.prefix == "" {
		return fmt.Errorf("prefix is required")
	}
	if len(cfg.brokers) == 0 {
		return fmt.Errorf("KAFKA_BROKERS is required")
	}
	if cfg.topic == "" {
		return fmt.Errorf("KAFKA_TOPIC is required")
	}
	return nil
}
