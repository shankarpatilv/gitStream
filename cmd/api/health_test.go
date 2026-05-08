package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubPinger struct {
	err error
}

func (s stubPinger) Ping(context.Context) error {
	return s.err
}

func TestHealthReturnsOKWhenDatabasesRespond(t *testing.T) {
	checker := healthChecker{postgres: stubPinger{}, clickHouse: stubPinger{}}
	recorder := httptest.NewRecorder()

	healthHandler(checker).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected ok response, got %s", recorder.Body.String())
	}
}

func TestHealthReturnsUnavailableWhenDatabaseFails(t *testing.T) {
	checker := healthChecker{
		postgres:   stubPinger{},
		clickHouse: stubPinger{err: errors.New("down")},
	}
	recorder := httptest.NewRecorder()

	healthHandler(checker).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"clickhouse":"unavailable"`) {
		t.Fatalf("expected clickhouse unavailable, got %s", recorder.Body.String())
	}
}
