package statsbus

import (
	"context"
	"errors"
	"testing"
	"time"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestKafkaEventProducer_SendEvent(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		inputEvent domain.RawStatsEvent
		wantErr    bool

		prepareMock func(t *testing.T, ctrl *gomock.Controller) domain.MessageWriter
	}

	testCases := []testCase{
		{
			name: "successful event send",
			inputEvent: domain.RawStatsEvent{
				UrlToken:  "abc123",
				Timestamp: time.Date(2025, 12, 18, 10, 0, 0, 0, time.UTC),
				IP:        "192.168.1.1",
				UserAgent: "Mozilla/5.0",
				Referrer:  "https://example.com",
			},
			wantErr: false,
			prepareMock: func(t *testing.T, ctrl *gomock.Controller) domain.MessageWriter {
				writer := mocks.NewMockMessageWriter(ctrl)
				writer.EXPECT().WriteMessages(gomock.Any(), gomock.Any()).Return(nil)
				return writer
			},
		},
		{
			name: "write messages error",
			inputEvent: domain.RawStatsEvent{
				UrlToken:  "abc123",
				Timestamp: time.Date(2025, 12, 18, 10, 0, 0, 0, time.UTC),
				IP:        "192.168.1.1",
				UserAgent: "Mozilla/5.0",
				Referrer:  "https://example.com",
			},
			wantErr: true,
			prepareMock: func(t *testing.T, ctrl *gomock.Controller) domain.MessageWriter {
				writer := mocks.NewMockMessageWriter(ctrl)
				writer.EXPECT().WriteMessages(gomock.Any(), gomock.Any()).Return(errors.New("kafka connection failed"))
				return writer
			},
		},
		{
			name: "event with empty fields",
			inputEvent: domain.RawStatsEvent{
				UrlToken:  "token",
				Timestamp: time.Time{},
				IP:        "",
				UserAgent: "",
				Referrer:  "",
			},
			wantErr: false,
			prepareMock: func(t *testing.T, ctrl *gomock.Controller) domain.MessageWriter {
				writer := mocks.NewMockMessageWriter(ctrl)
				writer.EXPECT().WriteMessages(gomock.Any(), gomock.Any()).Return(nil)
				return writer
			},
		},
		{
			name: "verifies message payload is correct JSON",
			inputEvent: domain.RawStatsEvent{
				UrlToken:  "test_token",
				Timestamp: time.Date(2025, 12, 18, 12, 30, 0, 0, time.UTC),
				IP:        "10.0.0.1",
				UserAgent: "TestAgent/1.0",
				Referrer:  "https://test.com",
			},
			wantErr: false,
			prepareMock: func(t *testing.T, ctrl *gomock.Controller) domain.MessageWriter {
				writer := mocks.NewMockMessageWriter(ctrl)
				writer.EXPECT().WriteMessages(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, msgs ...kafka.Message) error {
						assert.Len(t, msgs, 1)
						assert.Contains(t, string(msgs[0].Value), `"url_token":"test_token"`)
						assert.Contains(t, string(msgs[0].Value), `"ip":"10.0.0.1"`)
						assert.Contains(t, string(msgs[0].Value), `"user_agent":"TestAgent/1.0"`)
						assert.Contains(t, string(msgs[0].Value), `"referrer":"https://test.com"`)
						return nil
					},
				)
				return writer
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockWriter := tt.prepareMock(t, ctrl)
			producer := NewKafkaEventProducer(mockWriter)

			err := producer.SendEvent(context.Background(), tt.inputEvent)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
