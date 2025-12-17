package stats

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetDeviceType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		uaStr    string
		expected string
	}

	testCases := []testCase{
		{
			name:     "desktop chrome windows",
			uaStr:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expected: "Desktop",
		},
		{
			name:     "desktop firefox macos",
			uaStr:    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Gecko/20100101 Firefox/89.0",
			expected: "Desktop",
		},
		{
			name:     "mobile android chrome",
			uaStr:    "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36",
			expected: "Mobile",
		},
		{
			name:     "mobile iphone safari",
			uaStr:    "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1",
			expected: "Mobile",
		},
		{
			name:     "tablet ipad safari",
			uaStr:    "Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1",
			expected: "Tablet",
		},
		{
			name:     "bot googlebot",
			uaStr:    "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expected: "Bot",
		},
		{
			name:     "bot bingbot",
			uaStr:    "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
			expected: "Bot",
		},
		{
			name:     "empty user agent",
			uaStr:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := getDeviceType(tt.uaStr)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestParseEvent(t *testing.T) {
	t.Parallel()

	testTimestamp := time.Date(2025, 12, 17, 10, 0, 0, 0, time.UTC)

	type testCase struct {
		name        string
		input       []byte
		expected    domain.RawStatsEvent
		expectError bool
	}

	testCases := []testCase{
		{
			name: "valid event with all fields",
			input: func() []byte {
				event := domain.RawStatsEvent{
					UrlToken:  "abc123",
					Timestamp: testTimestamp,
					IP:        "8.8.8.8",
					UserAgent: "Mozilla/5.0",
					Referrer:  "https://google.com",
				}
				data, _ := json.Marshal(event)
				return data
			}(),
			expected: domain.RawStatsEvent{
				UrlToken:  "abc123",
				Timestamp: testTimestamp,
				IP:        "8.8.8.8",
				UserAgent: "Mozilla/5.0",
				Referrer:  "https://google.com",
			},
			expectError: false,
		},
		{
			name: "valid event with empty optional fields",
			input: func() []byte {
				event := domain.RawStatsEvent{
					UrlToken:  "xyz789",
					Timestamp: testTimestamp,
					IP:        "1.1.1.1",
					UserAgent: "",
					Referrer:  "",
				}
				data, _ := json.Marshal(event)
				return data
			}(),
			expected: domain.RawStatsEvent{
				UrlToken:  "xyz789",
				Timestamp: testTimestamp,
				IP:        "1.1.1.1",
				UserAgent: "",
				Referrer:  "",
			},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			input:       []byte("invalid json"),
			expected:    domain.RawStatsEvent{},
			expectError: true,
		},
		{
			name:        "empty input",
			input:       []byte(""),
			expected:    domain.RawStatsEvent{},
			expectError: true,
		},
		{
			name:        "null JSON",
			input:       []byte("null"),
			expected:    domain.RawStatsEvent{},
			expectError: false,
		},
		{
			name:        "empty object JSON",
			input:       []byte("{}"),
			expected:    domain.RawStatsEvent{},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := parseEvent(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestConvertEvent(t *testing.T) {
	t.Parallel()

	testTimestamp := time.Date(2025, 12, 17, 10, 0, 0, 0, time.UTC)

	type testCase struct {
		name     string
		input    domain.RawStatsEvent
		expected domain.ProcessedStatsEvent

		statsStorageFn func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder
		loggerFn       func(t *testing.T, ctrl *gomock.Controller) domain.Logger
	}

	testCases := []testCase{
		{
			name: "event with desktop user agent",
			input: domain.RawStatsEvent{
				UrlToken:  "abc123",
				Timestamp: testTimestamp,
				IP:        "8.8.8.8",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				Referrer:  "https://google.com",
			},
			expected: domain.ProcessedStatsEvent{
				UrlToken:   "abc123",
				Timestamp:  testTimestamp,
				DeviceType: "Desktop",
				Referrer:   "https://google.com",
			},
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				return mocks.NewMockStatsEventAdder(ctrl)
			},
			loggerFn: func(t *testing.T, ctrl *gomock.Controller) domain.Logger {
				mock := mocks.NewMockLogger(ctrl)
				mock.EXPECT().Warn(gomock.Any()).AnyTimes()
				return mock
			},
		},
		{
			name: "event with mobile user agent",
			input: domain.RawStatsEvent{
				UrlToken:  "mobile123",
				Timestamp: testTimestamp,
				IP:        "192.168.1.1",
				UserAgent: "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36",
				Referrer:  "https://twitter.com",
			},
			expected: domain.ProcessedStatsEvent{
				UrlToken:   "mobile123",
				Timestamp:  testTimestamp,
				DeviceType: "Mobile",
				Referrer:   "https://twitter.com",
			},
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				return mocks.NewMockStatsEventAdder(ctrl)
			},
			loggerFn: func(t *testing.T, ctrl *gomock.Controller) domain.Logger {
				mock := mocks.NewMockLogger(ctrl)
				mock.EXPECT().Warn(gomock.Any()).AnyTimes()
				return mock
			},
		},
		{
			name: "event with tablet user agent",
			input: domain.RawStatsEvent{
				UrlToken:  "tablet456",
				Timestamp: testTimestamp,
				IP:        "10.0.0.1",
				UserAgent: "Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1",
				Referrer:  "",
			},
			expected: domain.ProcessedStatsEvent{
				UrlToken:   "tablet456",
				Timestamp:  testTimestamp,
				DeviceType: "Tablet",
				Referrer:   "",
			},
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				return mocks.NewMockStatsEventAdder(ctrl)
			},
			loggerFn: func(t *testing.T, ctrl *gomock.Controller) domain.Logger {
				mock := mocks.NewMockLogger(ctrl)
				mock.EXPECT().Warn(gomock.Any()).AnyTimes()
				return mock
			},
		},
		{
			name: "event with bot user agent",
			input: domain.RawStatsEvent{
				UrlToken:  "bot789",
				Timestamp: testTimestamp,
				IP:        "66.249.66.1",
				UserAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
				Referrer:  "",
			},
			expected: domain.ProcessedStatsEvent{
				UrlToken:   "bot789",
				Timestamp:  testTimestamp,
				DeviceType: "Bot",
				Referrer:   "",
			},
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				return mocks.NewMockStatsEventAdder(ctrl)
			},
			loggerFn: func(t *testing.T, ctrl *gomock.Controller) domain.Logger {
				mock := mocks.NewMockLogger(ctrl)
				mock.EXPECT().Warn(gomock.Any()).AnyTimes()
				return mock
			},
		},
		{
			name: "event with empty user agent",
			input: domain.RawStatsEvent{
				UrlToken:  "empty123",
				Timestamp: testTimestamp,
				IP:        "172.16.0.1",
				UserAgent: "",
				Referrer:  "https://facebook.com",
			},
			expected: domain.ProcessedStatsEvent{
				UrlToken:   "empty123",
				Timestamp:  testTimestamp,
				DeviceType: "",
				Referrer:   "https://facebook.com",
			},
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				return mocks.NewMockStatsEventAdder(ctrl)
			},
			loggerFn: func(t *testing.T, ctrl *gomock.Controller) domain.Logger {
				mock := mocks.NewMockLogger(ctrl)
				mock.EXPECT().Warn(gomock.Any()).AnyTimes()
				return mock
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			statsStorage := tt.statsStorageFn(t, ctrl)
			logger := tt.loggerFn(t, ctrl)

			processor := NewRedirectStatsProcessor(statsStorage, logger)
			res := processor.convertEvent(tt.input)

			assert.Equal(t, tt.expected.UrlToken, res.UrlToken)
			assert.Equal(t, tt.expected.Timestamp, res.Timestamp)
			assert.Equal(t, tt.expected.DeviceType, res.DeviceType)
			assert.Equal(t, tt.expected.Referrer, res.Referrer)
		})
	}
}

func TestProcessEvent(t *testing.T) {
	t.Parallel()

	testTimestamp := time.Date(2025, 12, 17, 10, 0, 0, 0, time.UTC)

	type testCase struct {
		name           string
		input          []byte
		expectingError bool

		statsStorageFn func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder
	}

	testCases := []testCase{
		{
			name: "successful event processing",
			input: func() []byte {
				event := domain.RawStatsEvent{
					UrlToken:  "abc123",
					Timestamp: testTimestamp,
					IP:        "8.8.8.8",
					UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
					Referrer:  "https://google.com",
				}
				data, _ := json.Marshal(event)
				return data
			}(),
			expectingError: false,
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				mock := mocks.NewMockStatsEventAdder(ctrl)
				mock.EXPECT().
					AddStatsEvent(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, event domain.ProcessedStatsEvent) error {
						assert.Equal(t, "abc123", event.UrlToken)
						assert.Equal(t, testTimestamp, event.Timestamp)
						assert.Equal(t, "https://google.com", event.Referrer)
						return nil
					}).
					Times(1)
				return mock
			},
		},
		{
			name:           "invalid JSON returns error",
			input:          []byte("invalid json"),
			expectingError: true,
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				mock := mocks.NewMockStatsEventAdder(ctrl)
				// AddStatsEvent should not be called for invalid JSON
				mock.EXPECT().AddStatsEvent(gomock.Any(), gomock.Any()).Times(0)
				return mock
			},
		},
		{
			name: "storage error returns error",
			input: func() []byte {
				event := domain.RawStatsEvent{
					UrlToken:  "def456",
					Timestamp: testTimestamp,
					IP:        "1.1.1.1",
					UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
					Referrer:  "",
				}
				data, _ := json.Marshal(event)
				return data
			}(),
			expectingError: true,
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				mock := mocks.NewMockStatsEventAdder(ctrl)
				mock.EXPECT().
					AddStatsEvent(gomock.Any(), gomock.Any()).
					Return(errors.New("storage error")).
					Times(1)
				return mock
			},
		},
		{
			name:           "empty event data returns error",
			input:          []byte(""),
			expectingError: true,
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				mock := mocks.NewMockStatsEventAdder(ctrl)
				mock.EXPECT().AddStatsEvent(gomock.Any(), gomock.Any()).Times(0)
				return mock
			},
		},
		{
			name: "event with mobile user agent",
			input: func() []byte {
				event := domain.RawStatsEvent{
					UrlToken:  "mobile123",
					Timestamp: testTimestamp,
					IP:        "192.168.1.1",
					UserAgent: "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36",
					Referrer:  "https://twitter.com",
				}
				data, _ := json.Marshal(event)
				return data
			}(),
			expectingError: false,
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				mock := mocks.NewMockStatsEventAdder(ctrl)
				mock.EXPECT().
					AddStatsEvent(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, event domain.ProcessedStatsEvent) error {
						assert.Equal(t, "mobile123", event.UrlToken)
						assert.Equal(t, "Mobile", event.DeviceType)
						assert.Equal(t, "https://twitter.com", event.Referrer)
						return nil
					}).
					Times(1)
				return mock
			},
		},
		{
			name: "event with bot user agent",
			input: func() []byte {
				event := domain.RawStatsEvent{
					UrlToken:  "bot123",
					Timestamp: testTimestamp,
					IP:        "66.249.66.1",
					UserAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
					Referrer:  "",
				}
				data, _ := json.Marshal(event)
				return data
			}(),
			expectingError: false,
			statsStorageFn: func(t *testing.T, ctrl *gomock.Controller) domain.StatsEventAdder {
				mock := mocks.NewMockStatsEventAdder(ctrl)
				mock.EXPECT().
					AddStatsEvent(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, event domain.ProcessedStatsEvent) error {
						assert.Equal(t, "bot123", event.UrlToken)
						assert.Equal(t, "Bot", event.DeviceType)
						return nil
					}).
					Times(1)
				return mock
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			statsStorage := tt.statsStorageFn(t, ctrl)
			logger := mocks.NewMockLogger(ctrl)
			logger.EXPECT().Warn(gomock.Any()).AnyTimes()

			processor := NewRedirectStatsProcessor(statsStorage, logger)
			err := processor.ProcessEvent(context.Background(), tt.input)

			if tt.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
