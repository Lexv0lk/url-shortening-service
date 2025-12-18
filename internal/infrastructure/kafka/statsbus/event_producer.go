package statsbus

import (
	"context"
	"encoding/json"
	"url-shortening-service/internal/domain"

	"github.com/segmentio/kafka-go"
)

// KafkaEventProducer publishes statistics events to Kafka.
// It serializes events to JSON and sends them asynchronously.
type KafkaEventProducer struct {
	writer domain.MessageWriter
}

// NewKafkaEventProducer creates a new KafkaEventProducer instance.
func NewKafkaEventProducer(messageWriter domain.MessageWriter) *KafkaEventProducer {
	return &KafkaEventProducer{
		writer: messageWriter,
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
