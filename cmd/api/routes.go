package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type apiDependencies struct {
	health   healthChecker
	trending trendingStore
}

func newRouter(deps apiDependencies) http.Handler {
	router := chi.NewRouter()
	router.Get("/health", healthHandler(deps.health))
	router.Handle("/metrics", metricsHandler())
	router.Get("/api/trending", trendingHandler(deps.trending))
	return router
}
