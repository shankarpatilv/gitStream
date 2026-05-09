package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type githubPayload struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Repo      githubRepo  `json:"repo"`
	Actor     githubActor `json:"actor"`
	CreatedAt string      `json:"created_at"`
	Synthetic bool        `json:"synthetic"`
	Index     int         `json:"index"`
}

type githubRepo struct {
	Name string `json:"name"`
}

type githubActor struct {
	Login string `json:"login"`
}

func syntheticKafkaEvent(prefix string, i int, now time.Time) ([]byte, []byte, error) {
	repo := fmt.Sprintf("%s/repo-%03d", prefix, i%100)
	payload := githubPayload{
		ID:        fmt.Sprintf("%s-%06d", prefix, i),
		Type:      eventType(i),
		Repo:      githubRepo{Name: repo},
		Actor:     githubActor{Login: fmt.Sprintf("%s-actor-%03d", prefix, i%250)},
		CreatedAt: now.Add(-time.Duration(i%24) * time.Hour).Format(time.RFC3339),
		Synthetic: true,
		Index:     i,
	}
	value, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal synthetic event: %w", err)
	}
	return []byte(repo), value, nil
}

func eventType(i int) string {
	types := []string{"PushEvent", "PullRequestEvent", "IssuesEvent", "WatchEvent", "ForkEvent"}
	return types[i%len(types)]
}
