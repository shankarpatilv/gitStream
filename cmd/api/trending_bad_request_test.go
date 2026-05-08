package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrendingRejectsBadHours(t *testing.T) {
	store := &stubTrendingStore{}
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/trending?hours=bad", nil)
	trendingHandler(store).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestTrendingRejectsBadLimit(t *testing.T) {
	store := &stubTrendingStore{}
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/trending?limit=0", nil)
	trendingHandler(store).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestTrendingReturnsUnavailableOnStoreError(t *testing.T) {
	store := &stubTrendingStore{err: errors.New("down")}
	recorder := httptest.NewRecorder()

	trendingHandler(store).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/trending", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
}
