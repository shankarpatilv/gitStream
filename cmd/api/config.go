package main

import (
	"os"
	"strings"
)

const (
	defaultAPIPort        = "8090"
	defaultPostgresHost   = "localhost"
	defaultPostgresPort   = "5432"
	defaultPostgresDB     = "gitstream"
	defaultPostgresUser   = "gitstream"
	defaultClickHouseHost = "localhost"
	defaultClickHousePort = "9000"
	defaultClickHouseDB   = "gitstream"
	defaultClickHouseUser = "gitstream"
)

type config struct {
	port           string
	postgresHost   string
	postgresPort   string
	postgresDB     string
	postgresUser   string
	postgresPass   string
	clickHouseHost string
	clickHousePort string
	clickHouseDB   string
	clickHouseUser string
	clickHousePass string
}

func loadConfig() config {
	return config{
		port:           envOrDefault("API_PORT", defaultAPIPort),
		postgresHost:   envOrDefault("POSTGRES_HOST", defaultPostgresHost),
		postgresPort:   envOrDefault("POSTGRES_PORT", defaultPostgresPort),
		postgresDB:     envOrDefault("POSTGRES_DB", defaultPostgresDB),
		postgresUser:   envOrDefault("POSTGRES_USER", defaultPostgresUser),
		postgresPass:   envOrDefault("POSTGRES_PASSWORD", ""),
		clickHouseHost: envOrDefault("CLICKHOUSE_HOST", defaultClickHouseHost),
		clickHousePort: envOrDefault("CLICKHOUSE_NATIVE_PORT", defaultClickHousePort),
		clickHouseDB:   envOrDefault("CLICKHOUSE_DB", defaultClickHouseDB),
		clickHouseUser: envOrDefault("CLICKHOUSE_USER", defaultClickHouseUser),
		clickHousePass: envOrDefault("CLICKHOUSE_PASSWORD", ""),
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
