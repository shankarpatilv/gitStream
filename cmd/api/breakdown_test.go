package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vivekspatil/gitstream/internal/storage"
)

func TestBreakdownUsesDefaultHours(t *testing.T) {
	store := &stubBreakdownStore{
		breakdown: []storage.EventBreakdown{{EventType: "PushEvent", Count: 3}},
	}
	recorder := httptest.NewRecorder()

	breakdownHandler(store).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/stats/breakdown", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if store.hours != defaultBreakdownHours {
		t.Fatalf("hours = %d, want %d", store.hours, defaultBreakdownHours)
	}
	if !strings.Contains(recorder.Body.String(), `"event_type":"PushEvent"`) {
		t.Fatalf("expected event type in response, got %s", recorder.Body.String())
	}
}

func TestBreakdownUsesQueryHours(t *testing.T) {
	store := &stubBreakdownStore{}
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/stats/breakdown?hours=6", nil)
	breakdownHandler(store).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if store.hours != 6 {
		t.Fatalf("hours = %d, want 6", store.hours)
	}
}

func TestBreakdownRejectsBadHours(t *testing.T) {
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/stats/breakdown?hours=0", nil)
	breakdownHandler(&stubBreakdownStore{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestBreakdownRejectsHoursAboveMax(t *testing.T) {
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/stats/breakdown?hours=169", nil)
	breakdownHandler(&stubBreakdownStore{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestBreakdownAllowsEmptyResults(t *testing.T) {
	recorder := httptest.NewRecorder()

	breakdownHandler(&stubBreakdownStore{}).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/stats/breakdown", nil),
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestBreakdownReturnsUnavailableOnStoreError(t *testing.T) {
	store := &stubBreakdownStore{err: errors.New("down")}
	recorder := httptest.NewRecorder()

	breakdownHandler(store).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/stats/breakdown", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
}
