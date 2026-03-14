package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type closeProducer interface {
	Close() error
}

// runUntilShutdown keeps the server alive until an error or shutdown signal.
func runUntilShutdown(
	ctx context.Context,
	cancel context.CancelFunc,
	server *http.Server,
	producer closeProducer,
	pollerWG *sync.WaitGroup,
) {
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
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			slog.Error("server stopped unexpectedly", "error", err)
		}
	}

	cancel()
	shutdownServer(server)
	pollerWG.Wait()
	closeKafkaProducer(producer)
}

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		return
	}
	slog.Info("server stopped")
}

func closeKafkaProducer(producer closeProducer) {
	if err := producer.Close(); err != nil {
		slog.Error("kafka producer close failed", "error", err)
		return
	}
	slog.Info("kafka producer closed")
}
