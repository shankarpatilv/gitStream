package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// fetchGitHubEvents returns each event as raw JSON so the shared parser owns normalization.
func fetchGitHubEvents(
	ctx context.Context,
	client *http.Client,
	token string,
) ([]json.RawMessage, error) {
	request, err := buildGitHubEventsRequest(ctx, token)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send github events request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		return nil, fmt.Errorf("github events status %d: %s", response.StatusCode, body)
	}

	var rawEvents []json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&rawEvents); err != nil {
		return nil, fmt.Errorf("decode github events response: %w", err)
	}

	return rawEvents, nil
}

// buildGitHubEventsRequest prepares the GitHub API request before it is sent.
func buildGitHubEventsRequest(ctx context.Context, token string) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, githubEventsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create github events request: %w", err)
	}

	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "gitstream-ingest")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	return request, nil
}
