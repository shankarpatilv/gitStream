package kafka

import (
	"context"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafkago.Reader
}

type ConsumerConfig struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	Username      string
	Password      string
}

// Message keeps processor code independent of kafka-go's concrete message type.
type Message struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
}

// NewConsumer creates a group reader for manual offset commit flow.
func NewConsumer(config ConsumerConfig) (*Consumer, error) {
	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}
	if config.Topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}
	if config.ConsumerGroup == "" {
		return nil, fmt.Errorf("kafka consumer group is required")
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     config.Brokers,
		Topic:       config.Topic,
		GroupID:     config.ConsumerGroup,
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     1 * time.Second,
		Dialer:      newDialer(config.Brokers, config.Username, config.Password),
		StartOffset: kafkago.FirstOffset,
	})
	return &Consumer{reader: reader}, nil
}

// FetchMessage reads one record without telling Kafka it is finished.
func (c *Consumer) FetchMessage(ctx context.Context) (Message, error) {
	message, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return Message{}, fmt.Errorf("fetch kafka message: %w", err)
	}
	return Message{
		Topic:     message.Topic,
		Partition: message.Partition,
		Offset:    message.Offset,
		Key:       append([]byte(nil), message.Key...),
		Value:     append([]byte(nil), message.Value...),
	}, nil
}

// CommitMessage tells Kafka the message and earlier partition offsets are done.
func (c *Consumer) CommitMessage(ctx context.Context, message Message) error {
	err := c.reader.CommitMessages(ctx, kafkago.Message{
		Topic:     message.Topic,
		Partition: message.Partition,
		Offset:    message.Offset,
	})
	if err != nil {
		return fmt.Errorf("commit kafka message: %w", err)
	}
	return nil
}

// Close closes the underlying Kafka reader.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
