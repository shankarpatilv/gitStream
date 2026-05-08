package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testDependencies() apiDependencies {
	return apiDependencies{
		health: healthChecker{
			postgres:   stubPinger{},
			clickHouse: stubPinger{},
		},
		trending: &stubTrendingStore{},
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
