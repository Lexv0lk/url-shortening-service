package statsbus

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"

	"github.com/segmentio/kafka-go"
)

// KafkaEventConsumer consumes statistics events from Kafka and processes them.
// It reads messages from a Kafka topic and delegates processing to a StatisticsProcessor.
type KafkaEventConsumer struct {
	reader         *kafka.Reader
	statsProcessor domain.StatisticsProcessor
	logger         domain.Logger
}

// NewKafkaEventConsumer creates a new KafkaEventConsumer instance.
func NewKafkaEventConsumer(topic, groupId string, statsProcessor domain.StatisticsProcessor, logger domain.Logger, brokers ...string) *KafkaEventConsumer {
	return &KafkaEventConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupId,
		}),
		statsProcessor: statsProcessor,
		logger:         logger,
	}
}

// StartConsuming starts consuming messages from Kafka in a blocking loop.
// It continuously fetches messages, processes them, and commits offsets.
// The loop terminates when the context is cancelled.
// Errors during message fetch, processing, or commit are logged but don't stop the consumer.
func (kec *KafkaEventConsumer) StartConsuming(ctx context.Context) {
	defer func(reader *kafka.Reader) {
		err := reader.Close()
		if err != nil {
			kec.logger.Error(fmt.Sprintf("Failed to close kafka reader: %v", err))
		}
	}(kec.reader)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := kec.reader.FetchMessage(ctx)
			if err != nil {
				kec.logger.Error(fmt.Sprintf("Failed to fetch message from kafka: %v", err))
				continue
			}

			err = kec.statsProcessor.ProcessEvent(ctx, msg.Value)
			if err != nil {
				kec.logger.Error(fmt.Sprintf("Failed to process stats event: %v", err))
				continue
			}

			err = kec.reader.CommitMessages(ctx, msg)
			if err != nil {
				kec.logger.Error(fmt.Sprintf("Failed to commit message: %v", err))
			}
		}
	}
}
