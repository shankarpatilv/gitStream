package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore opens a connection pool and verifies Postgres is reachable.
func NewPostgresStore(ctx context.Context, config PostgresConfig) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

// Close releases the Postgres connection pool.
func (s *PostgresStore) Close() {
	s.pool.Close()
}
