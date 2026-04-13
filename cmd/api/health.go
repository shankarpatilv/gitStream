package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type healthChecker struct {
	postgres   pinger
	clickHouse pinger
}

type pinger interface {
	Ping(context.Context) error
}

func healthHandler(checker healthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		postgresOK := checker.postgres.Ping(ctx) == nil
		clickHouseOK := checker.clickHouse.Ping(ctx) == nil
		status := "ok"
		code := http.StatusOK
		if !postgresOK || !clickHouseOK {
			status = "unavailable"
			code = http.StatusServiceUnavailable
		}

		writeJSON(w, code, healthResponse{
			Status:     status,
			Postgres:   componentStatus(postgresOK),
			ClickHouse: componentStatus(clickHouseOK),
		})
	}
}

type healthResponse struct {
	Status     string `json:"status"`
	Postgres   string `json:"postgres"`
	ClickHouse string `json:"clickhouse"`
}

func componentStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "unavailable"
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(value)
}
