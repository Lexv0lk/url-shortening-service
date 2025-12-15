package statsbus

import (
	"context"
	"encoding/json"
	"time"
	"url-shortening-service/internal/domain"

	"github.com/segmentio/kafka-go"
)

type KafkaEventProducer struct {
	writer *kafka.Writer
}

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
