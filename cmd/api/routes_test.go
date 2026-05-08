package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testDependencies() apiDependencies {
	return apiDependencies{
		health: healthChecker{
			postgres:   stubPinger{},
			clickHouse: stubPinger{},
		},
		trending:     &stubTrendingStore{},
		recentEvents: &stubRecentEventsStore{},
		breakdown:    &stubBreakdownStore{},
		contributors: &stubContributorsStore{},
		pipeline:     newPipelineStats(time.Now()),
	}
}

func TestMetricsRouteReturnsPrometheusOutput(t *testing.T) {
	server := httptest.NewServer(newRouter(testDependencies()))
	defer server.Close()

	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("request metrics: %v", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(contentType, "text/plain") {
		t.Fatalf("expected prometheus content type, got %q", contentType)
	}
}

func TestPipelineStatsRouteExists(t *testing.T) {
	recorder := httptest.NewRecorder()

	newRouter(testDependencies()).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/stats/pipeline", nil),
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"events_processed_total":`) {
		t.Fatalf("expected pipeline response, got %s", recorder.Body.String())
	}
}

func TestLockedAPIRoutesRespond(t *testing.T) {
	router := newRouter(testDependencies())
	tests := []struct {
		name string
		path string
		code int
	}{
		{name: "health", path: "/health", code: http.StatusOK},
		{name: "metrics", path: "/metrics", code: http.StatusOK},
		{name: "trending", path: "/api/trending?hours=24&limit=5", code: http.StatusOK},
		{name: "recent events", path: "/api/events/recent?repo=a/b&limit=5", code: http.StatusOK},
		{name: "breakdown", path: "/api/stats/breakdown?hours=24", code: http.StatusOK},
		{name: "contributors", path: "/api/contributors/top?repo=a/b&limit=5", code: http.StatusOK},
		{name: "pipeline", path: "/api/stats/pipeline", code: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if recorder.Code != tt.code {
				t.Fatalf("expected status %d, got %d", tt.code, recorder.Code)
			}
		})
	}
}
