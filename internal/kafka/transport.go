package kafka

import (
	"crypto/tls"
	"net"
	"strings"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// newTransport gives Kafka writers optional SASL/TLS when credentials exist.
func newTransport(brokers []string, username, password string) kafkago.RoundTripper {
	if username == "" || password == "" {
		return nil
	}

	return &kafkago.Transport{
		TLS: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: tlsServerName(brokers),
		},
		SASL: plain.Mechanism{
			Username: username,
			Password: password,
		},
	}
}

func tlsServerName(brokers []string) string {
	if len(brokers) == 0 {
		return ""
	}
	host := strings.TrimSpace(brokers[0])
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	return strings.Trim(host, "[]")
}
