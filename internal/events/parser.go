package events

import (
	"encoding/json"
	"fmt"
	"time"
)

type githubAPIEvent struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Repo      named  `json:"repo"`
	Actor     login  `json:"actor"`
	CreatedAt string `json:"created_at"`
}

type named struct {
	Name string `json:"name"`
}

type login struct {
	Login string `json:"login"`
}

func ParseGitHubEvent(data []byte) (GitHubEvent, error) {
	var raw githubAPIEvent
	if err := json.Unmarshal(data, &raw); err != nil {
		return GitHubEvent{}, fmt.Errorf("parse github event json: %w", err)
	}

	if err := validateRawEvent(raw); err != nil {
		return GitHubEvent{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		return GitHubEvent{}, fmt.Errorf("parse created_at: %w", err)
	}

	return GitHubEvent{
		ID:        raw.ID,
		Type:      raw.Type,
		RepoName:  raw.Repo.Name,
		ActorName: raw.Actor.Login,
		CreatedAt: createdAt,
		Payload:   append([]byte(nil), data...),
	}, nil
}

func validateRawEvent(raw githubAPIEvent) error {
	switch {
	case raw.ID == "":
		return fmt.Errorf("missing required field id")
	case raw.Type == "":
		return fmt.Errorf("missing required field type")
	case raw.Repo.Name == "":
		return fmt.Errorf("missing required field repo.name")
	case raw.Actor.Login == "":
		return fmt.Errorf("missing required field actor.login")
	case raw.CreatedAt == "":
		return fmt.Errorf("missing required field created_at")
	default:
		return nil
	}
}
