package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type apiDependencies struct {
	health       healthChecker
	trending     trendingStore
	recentEvents recentEventsStore
	breakdown    breakdownStore
	contributors contributorsStore
	pipeline     pipelineStatsReader
}

func newRouter(deps apiDependencies) http.Handler {
	router := chi.NewRouter()
	router.Use(observeRequests)
	router.Get("/", dashboardHandler())
	router.Get("/dashboard", dashboardHandler())
	router.Get("/health", healthHandler(deps.health))
	router.Handle("/metrics", metricsHandler())
	router.Get("/api/trending", trendingHandler(deps.trending))
	router.Get("/api/events/recent", recentEventsHandler(deps.recentEvents))
	router.Get("/api/stats/breakdown", breakdownHandler(deps.breakdown))
	router.Get("/api/contributors/top", contributorsHandler(deps.contributors))
	router.Get("/api/stats/pipeline", pipelineStatsHandler(deps.pipeline))
	return router
}
