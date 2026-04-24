package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardRouteServesHTML(t *testing.T) {
	recorder := httptest.NewRecorder()

	newRouter(testDependencies()).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/dashboard", nil),
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected text/html content type, got %q", recorder.Header().Get("Content-Type"))
	}
	if !strings.Contains(recorder.Body.String(), "GitStream") {
		t.Fatalf("expected dashboard HTML, got %s", recorder.Body.String())
	}
}

func TestRootRedirectsToDashboard(t *testing.T) {
	recorder := httptest.NewRecorder()

	newRouter(testDependencies()).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected status 302, got %d", recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %q", location)
	}
}
