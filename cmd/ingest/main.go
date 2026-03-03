package main

import (
	"log/slog"
	"net/http"
)

func main() {
	const service = "ingest"

	loadDotEnv(".env")

	port := env("INGEST_PORT", "8080")
	addr := ":" + port

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	slog.Info("starting service", "service", service, "port", port)

	if err := server.ListenAndServe(); err != nil {
		slog.Error("server stopped", "service", service, "error", err)
	}
}
