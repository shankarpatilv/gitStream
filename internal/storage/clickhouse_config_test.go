package storage

import "testing"

func TestClickHouseConfigAddress(t *testing.T) {
	config := ClickHouseConfig{
		Host: "localhost",
		Port: "9000",
	}

	if got := config.Address(); got != "localhost:9000" {
		t.Fatalf("Address() = %q, want localhost:9000", got)
	}
}
