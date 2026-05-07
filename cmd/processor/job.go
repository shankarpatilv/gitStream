package main

import (
	"github.com/vivekspatil/gitstream/internal/events"
	"github.com/vivekspatil/gitstream/internal/kafka"
)

type job struct {
	message kafka.Message
	event   events.GitHubEvent
}
