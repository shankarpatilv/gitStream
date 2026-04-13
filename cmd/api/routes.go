package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func newRouter() http.Handler {
	router := chi.NewRouter()
	return router
}
