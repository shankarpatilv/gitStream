package storage

import "fmt"

// ClickHouseConfig holds the native TCP connection settings for ClickHouse.
type ClickHouseConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

// Address returns the host:port string expected by clickhouse-go.
func (c ClickHouseConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
