package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const service = "api"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	loadDotEnv(".env")
	cfg := loadConfig()
	server := &http.Server{
		Addr: ":" + cfg.port,
		Handler: newRouter(apiDependencies{
			health:       newHealthChecker(cfg),
			trending:     clickHouseTrendingStore{cfg: cfg},
			recentEvents: postgresRecentEventsStore{cfg: cfg},
			breakdown:    clickHouseBreakdownStore{cfg: cfg},
		}),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("starting service", "service", service, "port", cfg.port)
	if err := runServer(ctx, server); err != nil {
		slog.Error("server stopped unexpectedly", "service", service, "error", err)
		os.Exit(1)
	}
	slog.Info("service stopped", "service", service)
}

func runServer(ctx context.Context, server *http.Server) error {
	serverErr := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		return shutdownServer(server)
	case err := <-serverErr:
		return err
	}
}

func shutdownServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	slog.Info("server stopped", "service", service)
	return nil
}
