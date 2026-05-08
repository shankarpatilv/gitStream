package storage

import (
	"context"
	"fmt"
)

type TopContributor struct {
	ActorName string `json:"actor_name"`
	Count     uint64 `json:"count"`
}

// TopContributors returns the most active actors for one repository.
func (s *PostgresStore) TopContributors(
	ctx context.Context,
	repo string,
	limit int,
) ([]TopContributor, error) {
	rows, err := s.pool.Query(ctx, topContributorsSQL, repo, limit)
	if err != nil {
		return nil, fmt.Errorf("query top contributors: %w", err)
	}
	defer rows.Close()

	var contributors []TopContributor
	for rows.Next() {
		var contributor TopContributor
		if err := rows.Scan(&contributor.ActorName, &contributor.Count); err != nil {
			return nil, fmt.Errorf("scan top contributor: %w", err)
		}
		contributors = append(contributors, contributor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read top contributors: %w", err)
	}
	return contributors, nil
}

const topContributorsSQL = `
SELECT actor_name, COUNT(*)::bigint AS count
FROM events
WHERE repo_name = $1
GROUP BY actor_name
ORDER BY count DESC, actor_name ASC
LIMIT $2;`
