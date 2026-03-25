package kafka

import (
	"crypto/tls"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// newTransport gives Kafka writers optional SASL/TLS when credentials exist.
func newTransport(username, password string) kafkago.RoundTripper {
	if username == "" || password == "" {
		return nil
	}

	return &kafkago.Transport{
		TLS: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		SASL: plain.Mechanism{
			Username: username,
			Password: password,
		},
	}
}
