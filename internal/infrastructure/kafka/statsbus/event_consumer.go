package statsbus

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"
)

// StatsEventConsumer consumes statistics events from Kafka and processes them.
// It reads messages from a Kafka topic and delegates processing to a StatisticsProcessor.
type StatsEventConsumer struct {
	messageFetcher domain.MessageFetcher
	statsProcessor domain.StatisticsProcessor
	logger         domain.Logger
}

// NewStatsEventConsumer creates a new StatsEventConsumer instance.
func NewStatsEventConsumer(messageFetcher domain.MessageFetcher, statsProcessor domain.StatisticsProcessor, logger domain.Logger) *StatsEventConsumer {
	return &StatsEventConsumer{
		messageFetcher: messageFetcher,
		statsProcessor: statsProcessor,
		logger:         logger,
	}
}

// StartConsuming starts consuming messages in a blocking loop.
// It continuously fetches messages, processes them, and commits offsets.
// The loop terminates when the context is cancelled.
// Errors during message fetch, processing, or commit are logged but don't stop the consumer.
func (kec *StatsEventConsumer) StartConsuming(ctx context.Context) {
	defer func(reader domain.MessageFetcher) {
		err := reader.Close()
		if err != nil {
			kec.logger.Error(fmt.Sprintf("Failed to close messageFetcher: %v", err))
		}
	}(kec.messageFetcher)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := kec.messageFetcher.FetchMessage(ctx)
			if err != nil {
				kec.logger.Error(fmt.Sprintf("Failed to fetch message: %v", err))
				continue
			}

			err = kec.statsProcessor.ProcessEvent(ctx, msg.Value)
			if err != nil {
				kec.logger.Error(fmt.Sprintf("Failed to process stats event: %v", err))
				continue
			}

			err = kec.messageFetcher.CommitMessages(ctx, msg)
			if err != nil {
				kec.logger.Error(fmt.Sprintf("Failed to commit message: %v", err))
			}
		}
	}
}
