package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vivekspatil/gitstream/internal/storage"
)

func TestTrendingUsesDefaults(t *testing.T) {
	store := &stubTrendingStore{repos: []storage.TrendingRepo{{RepoName: "a/b", Count: 3}}}
	recorder := httptest.NewRecorder()

	trendingHandler(store).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/trending", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if store.hours != defaultTrendingHours || store.limit != defaultTrendingLimit {
		t.Fatalf("got hours=%d limit=%d", store.hours, store.limit)
	}
	if !strings.Contains(recorder.Body.String(), `"repo_name":"a/b"`) {
		t.Fatalf("expected repo in response, got %s", recorder.Body.String())
	}
}

func TestTrendingUsesQueryParams(t *testing.T) {
	store := &stubTrendingStore{}
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/trending?hours=24&limit=5", nil)
	trendingHandler(store).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if store.hours != 24 || store.limit != 5 {
		t.Fatalf("got hours=%d limit=%d", store.hours, store.limit)
	}
}
