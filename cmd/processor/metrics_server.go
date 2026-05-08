package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startMetricsServer(ctx context.Context, port string) <-chan struct{} {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
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
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				slog.Error("processor metrics server shutdown failed", "error", err)
			}
		case err := <-serverErr:
			if err != nil {
				slog.Error("processor metrics server stopped unexpectedly", "error", err)
			}
		}
	}()
	slog.Info("processor metrics server started", "port", port)
	return done
}

func waitForMetricsServer(done <-chan struct{}) {
	<-done
}
