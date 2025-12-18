package statsbus

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
)

func discardLogger() domain.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestStatsEventConsumer_StartConsuming(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string

		prepareMocks func(t *testing.T, ctrl *gomock.Controller, wg *sync.WaitGroup) (domain.MessageFetcher, domain.StatisticsProcessor, domain.Logger)
	}

	testCases := []testCase{
		{
			name: "successful message processing",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller, wg *sync.WaitGroup) (domain.MessageFetcher, domain.StatisticsProcessor, domain.Logger) {
				fetcher := mocks.NewMockMessageFetcher(ctrl)
				processor := mocks.NewMockStatisticsProcessor(ctrl)
				logger := discardLogger()

				msg := kafka.Message{Value: []byte(`{"url_token":"abc123"}`)}

				wg.Add(1)
				gomock.InOrder(
					fetcher.EXPECT().FetchMessage(gomock.Any()).Return(msg, nil),
					processor.EXPECT().ProcessEvent(gomock.Any(), msg.Value).Return(nil),
					fetcher.EXPECT().CommitMessages(gomock.Any(), msg).Return(nil),
					fetcher.EXPECT().FetchMessage(gomock.Any()).DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
						wg.Done()
						<-ctx.Done()
						return kafka.Message{}, ctx.Err()
					}),
					fetcher.EXPECT().Close().Return(nil),
				)

				return fetcher, processor, logger
			},
		},
		{
			name: "fetch message error logs and continues",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller, wg *sync.WaitGroup) (domain.MessageFetcher, domain.StatisticsProcessor, domain.Logger) {
				fetcher := mocks.NewMockMessageFetcher(ctrl)
				processor := mocks.NewMockStatisticsProcessor(ctrl)
				logger := mocks.NewMockLogger(ctrl)

				fetchErr := errors.New("fetch error")

				wg.Add(1)
				gomock.InOrder(
					fetcher.EXPECT().FetchMessage(gomock.Any()).Return(kafka.Message{}, fetchErr),
					logger.EXPECT().Error(gomock.Any()).Do(func(msg string, args ...any) {
						wg.Done()
					}),
					fetcher.EXPECT().FetchMessage(gomock.Any()).DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
						<-ctx.Done()
						return kafka.Message{}, ctx.Err()
					}),
					logger.EXPECT().Error(gomock.Any()),
					fetcher.EXPECT().Close().Return(nil),
				)

				return fetcher, processor, logger
			},
		},
		{
			name: "process event error logs and continues without commit",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller, wg *sync.WaitGroup) (domain.MessageFetcher, domain.StatisticsProcessor, domain.Logger) {
				fetcher := mocks.NewMockMessageFetcher(ctrl)
				processor := mocks.NewMockStatisticsProcessor(ctrl)
				logger := mocks.NewMockLogger(ctrl)

				msg := kafka.Message{Value: []byte(`{"url_token":"abc123"}`)}
				processErr := errors.New("process error")

				wg.Add(1)
				gomock.InOrder(
					fetcher.EXPECT().FetchMessage(gomock.Any()).Return(msg, nil),
					processor.EXPECT().ProcessEvent(gomock.Any(), msg.Value).Return(processErr),
					logger.EXPECT().Error(gomock.Any()).Do(func(msg string, args ...any) {
						wg.Done()
					}),
					fetcher.EXPECT().FetchMessage(gomock.Any()).DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
						<-ctx.Done()
						return kafka.Message{}, ctx.Err()
					}),
					logger.EXPECT().Error(gomock.Any()),
					fetcher.EXPECT().Close().Return(nil),
				)

				return fetcher, processor, logger
			},
		},
		{
			name: "commit message error logs",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller, wg *sync.WaitGroup) (domain.MessageFetcher, domain.StatisticsProcessor, domain.Logger) {
				fetcher := mocks.NewMockMessageFetcher(ctrl)
				processor := mocks.NewMockStatisticsProcessor(ctrl)
				logger := mocks.NewMockLogger(ctrl)

				msg := kafka.Message{Value: []byte(`{"url_token":"abc123"}`)}
				commitErr := errors.New("commit error")

				wg.Add(1)
				gomock.InOrder(
					fetcher.EXPECT().FetchMessage(gomock.Any()).Return(msg, nil),
					processor.EXPECT().ProcessEvent(gomock.Any(), msg.Value).Return(nil),
					fetcher.EXPECT().CommitMessages(gomock.Any(), msg).Return(commitErr),
					logger.EXPECT().Error(gomock.Any()).Do(func(msg string, args ...any) {
						wg.Done()
					}),
					fetcher.EXPECT().FetchMessage(gomock.Any()).DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
						<-ctx.Done()
						return kafka.Message{}, ctx.Err()
					}),
					logger.EXPECT().Error(gomock.Any()),
					fetcher.EXPECT().Close().Return(nil),
				)

				return fetcher, processor, logger
			},
		},
		{
			name: "close error logs",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller, wg *sync.WaitGroup) (domain.MessageFetcher, domain.StatisticsProcessor, domain.Logger) {
				fetcher := mocks.NewMockMessageFetcher(ctrl)
				processor := mocks.NewMockStatisticsProcessor(ctrl)
				logger := mocks.NewMockLogger(ctrl)

				closeErr := errors.New("close error")

				wg.Add(1)
				gomock.InOrder(
					fetcher.EXPECT().FetchMessage(gomock.Any()).DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
						wg.Done()
						<-ctx.Done()
						return kafka.Message{}, ctx.Err()
					}),
					logger.EXPECT().Error(gomock.Any()),
					fetcher.EXPECT().Close().Return(closeErr),
					logger.EXPECT().Error(gomock.Any()),
				)

				return fetcher, processor, logger
			},
		},
		{
			name: "multiple messages processed successfully",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller, wg *sync.WaitGroup) (domain.MessageFetcher, domain.StatisticsProcessor, domain.Logger) {
				fetcher := mocks.NewMockMessageFetcher(ctrl)
				processor := mocks.NewMockStatisticsProcessor(ctrl)
				logger := discardLogger()

				msg1 := kafka.Message{Value: []byte(`{"url_token":"token1"}`)}
				msg2 := kafka.Message{Value: []byte(`{"url_token":"token2"}`)}

				wg.Add(1)
				gomock.InOrder(
					fetcher.EXPECT().FetchMessage(gomock.Any()).Return(msg1, nil),
					processor.EXPECT().ProcessEvent(gomock.Any(), msg1.Value).Return(nil),
					fetcher.EXPECT().CommitMessages(gomock.Any(), msg1).Return(nil),
					fetcher.EXPECT().FetchMessage(gomock.Any()).Return(msg2, nil),
					processor.EXPECT().ProcessEvent(gomock.Any(), msg2.Value).Return(nil),
					fetcher.EXPECT().CommitMessages(gomock.Any(), msg2).Return(nil),
					fetcher.EXPECT().FetchMessage(gomock.Any()).DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
						wg.Done()
						<-ctx.Done()
						return kafka.Message{}, ctx.Err()
					}),
					fetcher.EXPECT().Close().Return(nil),
				)

				return fetcher, processor, logger
			},
		},
		{
			name: "context cancelled immediately",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller, wg *sync.WaitGroup) (domain.MessageFetcher, domain.StatisticsProcessor, domain.Logger) {
				fetcher := mocks.NewMockMessageFetcher(ctrl)
				processor := mocks.NewMockStatisticsProcessor(ctrl)
				logger := discardLogger()

				wg.Add(1)
				gomock.InOrder(
					fetcher.EXPECT().FetchMessage(gomock.Any()).DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
						wg.Done()
						<-ctx.Done()
						return kafka.Message{}, ctx.Err()
					}),
					fetcher.EXPECT().Close().Return(nil),
				)

				return fetcher, processor, logger
			},
		},
		{
			name: "commit error does not prevent next message processing",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller, wg *sync.WaitGroup) (domain.MessageFetcher, domain.StatisticsProcessor, domain.Logger) {
				fetcher := mocks.NewMockMessageFetcher(ctrl)
				processor := mocks.NewMockStatisticsProcessor(ctrl)
				logger := discardLogger()

				msg1 := kafka.Message{Value: []byte(`{"url_token":"token1"}`)}
				msg2 := kafka.Message{Value: []byte(`{"url_token":"token2"}`)}
				commitErr := errors.New("commit error")

				wg.Add(1)
				gomock.InOrder(
					fetcher.EXPECT().FetchMessage(gomock.Any()).Return(msg1, nil),
					processor.EXPECT().ProcessEvent(gomock.Any(), msg1.Value).Return(nil),
					fetcher.EXPECT().CommitMessages(gomock.Any(), msg1).Return(commitErr),
					fetcher.EXPECT().FetchMessage(gomock.Any()).Return(msg2, nil),
					processor.EXPECT().ProcessEvent(gomock.Any(), msg2.Value).Return(nil),
					fetcher.EXPECT().CommitMessages(gomock.Any(), msg2).Return(nil),
					fetcher.EXPECT().FetchMessage(gomock.Any()).DoAndReturn(func(ctx context.Context) (kafka.Message, error) {
						wg.Done()
						<-ctx.Done()
						return kafka.Message{}, ctx.Err()
					}),
					fetcher.EXPECT().Close().Return(nil),
				)

				return fetcher, processor, logger
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			var wg sync.WaitGroup
			fetcher, processor, logger := tt.prepareMocks(t, ctrl, &wg)
			consumer := NewStatsEventConsumer(fetcher, processor, logger)
			ctx, cancelFunc := context.WithCancel(context.Background())

			var done sync.WaitGroup
			done.Add(1)
			go func() {
				defer done.Done()
				consumer.StartConsuming(ctx)
			}()

			wg.Wait()
			cancelFunc()
			done.Wait()
		})
	}
}
