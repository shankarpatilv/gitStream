package storage

import (
	"fmt"
	"net/url"
)

type PostgresConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

// ConnectionString builds the pgx DSN used by the local processor service.
func (c PostgresConfig) ConnectionString() string {
	values := url.Values{}
	values.Set("sslmode", "disable")

	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:     c.Database,
		RawQuery: values.Encode(),
	}
	return dsn.String()
}
