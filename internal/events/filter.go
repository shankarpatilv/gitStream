package events

var allowedEventTypes = map[string]struct{}{
	"PushEvent":        {},
	"PullRequestEvent": {},
	"IssuesEvent":      {},
	"WatchEvent":       {},
	"ForkEvent":        {},
}

func IsAllowedEventType(eventType string) bool {
	_, ok := allowedEventTypes[eventType]
	return ok
}

func FilterAllowedEvents(events []GitHubEvent) []GitHubEvent {
	filtered := make([]GitHubEvent, 0, len(events))
	for _, event := range events {
		if IsAllowedEventType(event.Type) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
