package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
)

func main() {
	const service = "ingest"

	loadDotEnv(".env")

	port := env("INGEST_PORT", "8080")
	githubToken := env("GITHUB_TOKEN", "")
	addr := ":" + port
	if githubToken == "" {
		slog.Warn("github token not configured; polling without authentication")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("GET /metrics", metricsHandler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go runGitHubPoller(
		context.Background(),
		&http.Client{Timeout: 10 * time.Second},
		githubToken,
		events.NewDeduper(events.DefaultDeduperLimit),
		30*time.Second,
	)

	slog.Info("starting service", "service", service, "port", port)

	if err := server.ListenAndServe(); err != nil {
		slog.Error("server stopped", "service", service, "error", err)
	}
}
