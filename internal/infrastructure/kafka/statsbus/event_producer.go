package statsbus

import (
	"context"
	"encoding/json"
	"time"
	"url-shortening-service/internal/domain"

	"github.com/segmentio/kafka-go"
)

// KafkaEventProducer publishes statistics events to Kafka.
// It serializes events to JSON and sends them asynchronously.
type KafkaEventProducer struct {
	writer *kafka.Writer
}

// NewKafkaEventProducer creates a new KafkaEventProducer instance.
func NewKafkaEventProducer(topic string, batchTimeout time.Duration, kafkaAddresses ...string) *KafkaEventProducer {
	return &KafkaEventProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(kafkaAddresses...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: batchTimeout,
			Async:        true,
		},
	}
}

// SendEvent publishes a raw statistics event to Kafka.
// The event is serialized to JSON before sending.
//
// Returns an error if:
//   - JSON serialization fails
//   - Kafka write operation fails
func (kep *KafkaEventProducer) SendEvent(ctx context.Context, rawEvent domain.RawStatsEvent) error {
	payload, err := json.Marshal(rawEvent)
	if err != nil {
		return err
	}

	err = kep.writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
	})

	if err != nil {
		return err
	}

	return nil
}
