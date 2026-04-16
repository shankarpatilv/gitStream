package main

import (
	"context"
	"fmt"
	"time"

	"github.com/vivekspatil/gitstream/internal/events"
	"github.com/vivekspatil/gitstream/internal/storage"
)

func seedClickHouse(ctx context.Context, store *storage.ClickHouseStore, cfg config) error {
	for start := 0; start < cfg.rows; start += cfg.batchSize {
		end := min(start+cfg.batchSize, cfg.rows)
		if err := store.InsertAnalyticsBatch(ctx, syntheticBatch(cfg, start, end)); err != nil {
			return fmt.Errorf("insert synthetic batch %d-%d: %w", start, end, err)
		}
	}
	return nil
}

func syntheticBatch(cfg config, start int, end int) []events.GitHubEvent {
	batch := make([]events.GitHubEvent, 0, end-start)
	for i := start; i < end; i++ {
		batch = append(batch, syntheticEvent(cfg.prefix, i))
	}
	return batch
}

func syntheticEvent(prefix string, i int) events.GitHubEvent {
	eventType := []string{"PushEvent", "PullRequestEvent", "IssuesEvent", "WatchEvent", "ForkEvent"}[i%5]
	repoRank := i % 100
	return events.GitHubEvent{
		ID:        fmt.Sprintf("%s-%06d", prefix, i),
		Type:      eventType,
		RepoName:  fmt.Sprintf("%s/repo-%03d", prefix, repoRank),
		ActorName: fmt.Sprintf("%s-actor-%03d", prefix, i%250),
		CreatedAt: time.Now().UTC().Add(-time.Duration(i%24) * time.Hour),
		Payload:   []byte(`{"synthetic":true}`),
	}
}
