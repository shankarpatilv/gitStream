package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func newRouter(checker healthChecker) http.Handler {
	router := chi.NewRouter()
	router.Get("/health", healthHandler(checker))
	router.Handle("/metrics", metricsHandler())
	return router
}
