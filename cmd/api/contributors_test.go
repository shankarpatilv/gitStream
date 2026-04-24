package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vivekspatil/gitstream/internal/storage"
)

func TestContributorsUsesRepoAndDefaultLimit(t *testing.T) {
	store := &stubContributorsStore{
		contributors: []storage.TopContributor{{ActorName: "codex", Count: 2}},
	}
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/contributors/top?repo=a/b", nil)
	contributorsHandler(store).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if store.repo != "a/b" || store.limit != defaultContributorsLimit {
		t.Fatalf("got repo=%q limit=%d", store.repo, store.limit)
	}
	if !strings.Contains(recorder.Body.String(), `"actor_name":"codex"`) {
		t.Fatalf("expected contributor in response, got %s", recorder.Body.String())
	}
}

func TestContributorsRejectsMissingRepo(t *testing.T) {
	recorder := httptest.NewRecorder()

	contributorsHandler(&stubContributorsStore{}).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/contributors/top", nil),
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestContributorsRejectsBadLimit(t *testing.T) {
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/contributors/top?repo=a/b&limit=0", nil)
	contributorsHandler(&stubContributorsStore{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestContributorsRejectsLimitAboveMax(t *testing.T) {
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/contributors/top?repo=a/b&limit=101", nil)
	contributorsHandler(&stubContributorsStore{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestContributorsReturnsUnavailableOnStoreError(t *testing.T) {
	store := &stubContributorsStore{err: errors.New("down")}
	recorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, "/api/contributors/top?repo=a/b", nil)
	contributorsHandler(store).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
}
