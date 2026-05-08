package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vivekspatil/gitstream/internal/storage"
)

func TestRecentEventsUsesRepoAndDefaultLimit(t *testing.T) {
	store := &stubRecentEventsStore{events: []storage.RecentEvent{{ID: "1", RepoName: "a/b"}}}
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/events/recent?repo=a/b", nil)
	recentEventsHandler(store).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if store.repo != "a/b" || store.limit != defaultRecentLimit {
		t.Fatalf("got repo=%q limit=%d", store.repo, store.limit)
	}
	if !strings.Contains(recorder.Body.String(), `"repo":"a/b"`) {
		t.Fatalf("expected repo in response, got %s", recorder.Body.String())
	}
}

func TestRecentEventsRejectsMissingRepo(t *testing.T) {
	recorder := httptest.NewRecorder()

	recentEventsHandler(&stubRecentEventsStore{}).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/events/recent", nil),
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestRecentEventsRejectsBadLimit(t *testing.T) {
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/events/recent?repo=a/b&limit=0", nil)
	recentEventsHandler(&stubRecentEventsStore{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestRecentEventsRejectsLimitAboveMax(t *testing.T) {
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/events/recent?repo=a/b&limit=101", nil)
	recentEventsHandler(&stubRecentEventsStore{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestRecentEventsReturnsUnavailableOnStoreError(t *testing.T) {
	store := &stubRecentEventsStore{err: errors.New("down")}
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/events/recent?repo=a/b", nil)
	recentEventsHandler(store).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
}
