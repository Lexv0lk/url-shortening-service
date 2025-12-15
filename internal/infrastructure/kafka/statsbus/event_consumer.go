package statsbus

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"

	"github.com/segmentio/kafka-go"
)

type KafkaEventConsumer struct {
	reader         *kafka.Reader
	statsProcessor domain.StatisticsProcessor
	logger         domain.Logger
}

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
