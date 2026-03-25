package main

import (
	"fmt"

	"github.com/vivekspatil/gitstream/internal/events"
	"github.com/vivekspatil/gitstream/internal/kafka"
)

// decodeMessageEvent turns the Kafka payload back into the shared event model.
func decodeMessageEvent(message kafka.Message) (events.GitHubEvent, error) {
	event, err := events.ParseGitHubEvent(message.Value)
	if err != nil {
		return events.GitHubEvent{}, fmt.Errorf("decode kafka event value: %w", err)
	}
	return event, nil
}
