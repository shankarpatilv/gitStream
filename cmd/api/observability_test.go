package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIMetricsIncludeRequestCounters(t *testing.T) {
	server := httptest.NewServer(newRouter(testDependencies()))
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("request health: %v", err)
	}
	resp.Body.Close()

	metricsResp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("request metrics: %v", err)
	}
	defer metricsResp.Body.Close()

	body := readResponseBody(t, metricsResp)
	if !strings.Contains(body, "gitstream_api_requests_total") {
		t.Fatalf("expected API request counter in metrics, got %s", body)
	}
	if !strings.Contains(body, "gitstream_api_request_duration_seconds") {
		t.Fatalf("expected API duration histogram in metrics, got %s", body)
	}
}

func readResponseBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return string(body)
}
