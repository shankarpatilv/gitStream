package events

import "time"

type GitHubEvent struct {
	ID        string
	Type      string
	RepoName  string
	ActorName string
	CreatedAt time.Time
	Payload   []byte
}
