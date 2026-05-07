package kafka

import (
	"crypto/tls"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// newDialer supports local plaintext Kafka and future SASL/TLS cloud Kafka.
func newDialer(username, password string) *kafkago.Dialer {
	dialer := &kafkago.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	if username == "" || password == "" {
		return dialer
	}

	dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	dialer.SASLMechanism = plain.Mechanism{
		Username: username,
		Password: password,
	}
	return dialer
}
