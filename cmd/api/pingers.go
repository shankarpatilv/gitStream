package main

import (
	"context"

	"github.com/vivekspatil/gitstream/internal/storage"
)

type databasePinger func(context.Context) error

func (p databasePinger) Ping(ctx context.Context) error {
	return p(ctx)
}

func newHealthChecker(cfg config) healthChecker {
	return healthChecker{
		postgres: databasePinger(func(ctx context.Context) error {
			store, err := storage.NewPostgresStore(ctx, postgresConfig(cfg))
			if err != nil {
				return err
			}
			store.Close()
			return nil
		}),
		clickHouse: databasePinger(func(ctx context.Context) error {
			store, err := storage.NewClickHouseStore(ctx, clickHouseConfig(cfg))
			if err != nil {
				return err
			}
			return store.Close()
		}),
	}
}

func postgresConfig(cfg config) storage.PostgresConfig {
	return storage.PostgresConfig{
		Host:     cfg.postgresHost,
		Port:     cfg.postgresPort,
		Database: cfg.postgresDB,
		User:     cfg.postgresUser,
		Password: cfg.postgresPass,
	}
}

func clickHouseConfig(cfg config) storage.ClickHouseConfig {
	return storage.ClickHouseConfig{
		Host:     cfg.clickHouseHost,
		Port:     cfg.clickHousePort,
		Database: cfg.clickHouseDB,
		User:     cfg.clickHouseUser,
		Password: cfg.clickHousePass,
	}
}
