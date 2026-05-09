package kafka

import (
	"context"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/vivekspatil/gitstream/internal/events"
)

type Producer struct {
	writer *kafkago.Writer
}

type ProducerConfig struct {
	Brokers  []string
	Topic    string
	Username string
	Password string
}

// NewProducer creates a Kafka writer for accepted GitHub events.
func NewProducer(config ProducerConfig) (*Producer, error) {
	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}
	if config.Topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}

	writer := &kafkago.Writer{
		Addr:                   kafkago.TCP(config.Brokers...),
		Topic:                  config.Topic,
		Balancer:               &kafkago.Hash{},
		RequiredAcks:           kafkago.RequireOne,
		BatchTimeout:           100 * time.Millisecond,
		AllowAutoTopicCreation: true,
		Transport:              newTransport(config.Username, config.Password),
	}

	return &Producer{writer: writer}, nil
}

// Publish writes the raw GitHub event payload using repo name as the Kafka key.
func (p *Producer) Publish(ctx context.Context, event events.GitHubEvent) error {
	message := kafkago.Message{
		Key:   []byte(event.RepoName),
		Value: event.Payload,
	}
	return p.PublishRaw(ctx, message.Key, message.Value)
}

// PublishRaw writes a message while preserving the caller-provided key/value.
func (p *Producer) PublishRaw(ctx context.Context, key, value []byte) error {
	return p.PublishRawBatch(ctx, []RawMessage{{Key: key, Value: value}})
}

// PublishRawBatch writes messages while preserving caller-provided keys/values.
func (p *Producer) PublishRawBatch(ctx context.Context, raw []RawMessage) error {
	messages := make([]kafkago.Message, 0, len(raw))
	for _, next := range raw {
		messages = append(messages, kafkago.Message{
			Key:   append([]byte(nil), next.Key...),
			Value: append([]byte(nil), next.Value...),
		})
	}
	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
