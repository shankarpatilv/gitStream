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
