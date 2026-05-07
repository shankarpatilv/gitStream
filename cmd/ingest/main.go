package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
	"github.com/vivekspatil/gitstream/internal/kafka"
)

func main() {
	const service = "ingest"

	loadDotEnv(".env")
	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	ctx, cancel := context.WithCancel(signalCtx)
	defer cancel()

	port := env("INGEST_PORT", "8080")
	githubToken := env("GITHUB_TOKEN", "")
	addr := ":" + port
	if githubToken == "" {
		slog.Warn("github token not configured; polling without authentication")
	}

	producer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:  envList("KAFKA_BROKERS", "localhost:9092"),
		Topic:    env("KAFKA_TOPIC", "github-events"),
		Username: env("KAFKA_USERNAME", ""),
		Password: env("KAFKA_PASSWORD", ""),
	})
	if err != nil {
		slog.Error("could not create kafka producer", "error", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("GET /metrics", metricsHandler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	var pollerWG sync.WaitGroup
	pollerWG.Add(1)
	go func() {
		defer pollerWG.Done()
		runGitHubPoller(
			ctx,
			&http.Client{Timeout: 10 * time.Second},
			githubToken,
			events.NewDeduper(events.DefaultDeduperLimit),
			producer,
			30*time.Second,
		)
	}()

	slog.Info("starting service", "service", service, "port", port)
	runUntilShutdown(ctx, cancel, server, producer, &pollerWG)
	slog.Info("service stopped", "service", service)
}
