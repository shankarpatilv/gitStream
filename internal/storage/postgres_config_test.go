package storage

import (
	"net/url"
	"testing"
)

func TestPostgresConfigConnectionString(t *testing.T) {
	config := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		Database: "gitstream",
		User:     "gitstream",
		Password: "test password",
	}

	parsed, err := url.Parse(config.ConnectionString())
	if err != nil {
		t.Fatalf("could not parse connection string: %v", err)
	}

	if parsed.Scheme != "postgres" || parsed.Host != "localhost:5432" {
		t.Fatalf("unexpected scheme/host: %q %q", parsed.Scheme, parsed.Host)
	}
	if parsed.Path != "/gitstream" || parsed.Query().Get("sslmode") != "disable" {
		t.Fatalf("unexpected path/query: %q %q", parsed.Path, parsed.RawQuery)
	}
	if password, _ := parsed.User.Password(); parsed.User.Username() != "gitstream" || password != "test password" {
		t.Fatalf("unexpected user info")
	}
}
