package main

import "fmt"

func validateConfig(cfg config) error {
	if cfg.rows < 1 {
		return fmt.Errorf("rows must be at least 1")
	}
	if cfg.batchSize < 1 {
		return fmt.Errorf("batch-size must be at least 1")
	}
	if cfg.prefix == "" {
		return fmt.Errorf("prefix is required")
	}
	if cfg.clickhouse.Host == "" || cfg.clickhouse.Port == "" {
		return fmt.Errorf("clickhouse host and port are required")
	}
	return nil
}
