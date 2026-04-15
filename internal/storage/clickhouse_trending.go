package storage

import (
	"context"
	"fmt"
	"time"
)

type TrendingRepo struct {
	RepoName string `json:"repo_name"`
	Count    uint64 `json:"count"`
}

// TrendingRepos returns repository activity counts from ClickHouse aggregates.
func (s *ClickHouseStore) TrendingRepos(
	ctx context.Context,
	hours int,
	limit int,
) ([]TrendingRepo, error) {
	// Use a Go-computed cutoff so the API owns the requested time window.
	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	rows, err := s.conn.Query(ctx, trendingReposSQL, cutoff, uint64(limit))
	if err != nil {
		return nil, fmt.Errorf("query trending repos: %w", err)
	}
	defer rows.Close()

	var repos []TrendingRepo
	for rows.Next() {
		var repo TrendingRepo
		if err := rows.Scan(&repo.RepoName, &repo.Count); err != nil {
			return nil, fmt.Errorf("scan trending repo: %w", err)
		}
		repos = append(repos, repo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read trending repos: %w", err)
	}
	return repos, nil
}

const trendingReposSQL = `
SELECT repo_name, sum(count) AS total
FROM events_hourly
WHERE hour >= ?
GROUP BY repo_name
ORDER BY total DESC, repo_name ASC
LIMIT ?`
