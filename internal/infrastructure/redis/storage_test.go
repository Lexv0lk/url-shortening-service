package redis

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisStorage_GetOriginalUrl(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		shortUrl   string
		wantUrl    string
		wantExists bool

		setupMock func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger)
	}

	testCases := []testCase{
		{
			name:       "URL exists in Redis",
			shortUrl:   "short123",
			wantUrl:    "http://example.com/original",
			wantExists: true,
			setupMock: func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger) {
				mockClient := mocks.NewMockKeyStorage(ctrl)
				mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

				mockClient.EXPECT().
					Get(gomock.Any(), "short123").
					DoAndReturn(func(ctx context.Context, key string) *redis.StringCmd {
						strCmd := redis.NewStringCmd(ctx)
						strCmd.SetVal("http://example.com/original")
						return strCmd
					}).
					Times(1)

				return mockClient, mockLogger
			},
		},
		{
			name:       "URL does not exist in Redis",
			shortUrl:   "nonexistent",
			wantUrl:    "",
			wantExists: false,
			setupMock: func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger) {
				mockClient := mocks.NewMockKeyStorage(ctrl)
				mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

				mockClient.EXPECT().
					Get(gomock.Any(), "nonexistent").
					DoAndReturn(func(ctx context.Context, key string) *redis.StringCmd {
						strCmd := redis.NewStringCmd(ctx)
						strCmd.SetErr(redis.Nil)
						return strCmd
					}).
					Times(1)

				return mockClient, mockLogger
			},
		},
		{
			name:       "Redis GET error",
			shortUrl:   "errorcase",
			wantUrl:    "",
			wantExists: false,
			setupMock: func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger) {
				mockClient := mocks.NewMockKeyStorage(ctrl)
				mockLogger := mocks.NewMockLogger(ctrl)

				mockClient.EXPECT().
					Get(gomock.Any(), "errorcase").
					DoAndReturn(func(ctx context.Context, key string) *redis.StringCmd {
						strCmd := redis.NewStringCmd(ctx)
						strCmd.SetErr(assert.AnError)
						return strCmd
					}).
					Times(1)

				mockLogger.EXPECT().
					Error(gomock.Any()).
					Times(1)

				return mockClient, mockLogger
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockClient, mockLogger := tt.setupMock(t, ctrl)
			storage := NewRedisStorage(mockClient, mockLogger)

			gotUrl, gotExists := storage.GetOriginalUrl(nil, tt.shortUrl)
			assert.Equal(t, tt.wantUrl, gotUrl)
			assert.Equal(t, tt.wantExists, gotExists)
		})
	}
}

func TestRedisStorage_SetMapping(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		originalUrl string
		urlToken    string
		wantErr     bool

		setupMock func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger)
	}

	testCases := []testCase{
		{
			name:        "Successfully set mapping",
			originalUrl: "http://example.com/original",
			urlToken:    "short123",
			wantErr:     false,
			setupMock: func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger) {
				mockClient := mocks.NewMockKeyStorage(ctrl)
				mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

				mockClient.EXPECT().
					Set(gomock.Any(), "short123", "http://example.com/original", gomock.Any()).
					DoAndReturn(func(ctx context.Context, key string, value interface{}, expiration interface{}) *redis.StatusCmd {
						statusCmd := redis.NewStatusCmd(ctx)
						statusCmd.SetVal("OK")
						return statusCmd
					}).
					Times(1)

				return mockClient, mockLogger
			},
		},
		{
			name:        "Redis SET error",
			originalUrl: "http://example.com/error",
			urlToken:    "errortoken",
			wantErr:     true,
			setupMock: func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger) {
				mockClient := mocks.NewMockKeyStorage(ctrl)
				mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

				mockClient.EXPECT().
					Set(gomock.Any(), "errortoken", "http://example.com/error", gomock.Any()).
					DoAndReturn(func(ctx context.Context, key string, value interface{}, expiration interface{}) *redis.StatusCmd {
						statusCmd := redis.NewStatusCmd(ctx)
						statusCmd.SetErr(assert.AnError)
						return statusCmd
					}).
					Times(1)

				return mockClient, mockLogger
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockClient, mockLogger := tt.setupMock(t, ctrl)
			storage := NewRedisStorage(mockClient, mockLogger)

			err := storage.SetMapping(context.Background(), tt.originalUrl, tt.urlToken)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedisStorage_DeleteMapping(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		urlToken string
		wantErr  error

		setupMock func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger)
	}

	testCases := []testCase{
		{
			name:     "Successfully delete mapping",
			urlToken: "short123",
			wantErr:  nil,
			setupMock: func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger) {
				mockClient := mocks.NewMockKeyStorage(ctrl)
				mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

				mockClient.EXPECT().
					Del(gomock.Any(), "short123").
					DoAndReturn(func(ctx context.Context, keys ...string) *redis.IntCmd {
						intCmd := redis.NewIntCmd(ctx)
						intCmd.SetVal(1)
						return intCmd
					}).
					Times(1)

				return mockClient, mockLogger
			},
		},
		{
			name:     "Token does not exist",
			urlToken: "nonexistent",
			wantErr:  &domain.TokenNonExistingError{},
			setupMock: func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger) {
				mockClient := mocks.NewMockKeyStorage(ctrl)
				mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

				mockClient.EXPECT().
					Del(gomock.Any(), "nonexistent").
					DoAndReturn(func(ctx context.Context, keys ...string) *redis.IntCmd {
						intCmd := redis.NewIntCmd(ctx)
						intCmd.SetVal(0)
						return intCmd
					}).
					Times(1)

				return mockClient, mockLogger
			},
		},
		{
			name:     "Redis DEL error",
			urlToken: "errortoken",
			wantErr:  assert.AnError,
			setupMock: func(t *testing.T, ctrl *gomock.Controller) (domain.KeyStorage, domain.Logger) {
				mockClient := mocks.NewMockKeyStorage(ctrl)
				mockLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

				mockClient.EXPECT().
					Del(gomock.Any(), "errortoken").
					DoAndReturn(func(ctx context.Context, keys ...string) *redis.IntCmd {
						intCmd := redis.NewIntCmd(ctx)
						intCmd.SetErr(assert.AnError)
						return intCmd
					}).
					Times(1)

				return mockClient, mockLogger
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockClient, mockLogger := tt.setupMock(t, ctrl)
			storage := NewRedisStorage(mockClient, mockLogger)

			err := storage.DeleteMapping(context.Background(), tt.urlToken)
			if tt.wantErr != nil {
				assert.Error(t, err)
				if _, ok := tt.wantErr.(*domain.TokenNonExistingError); ok {
					assert.IsType(t, &domain.TokenNonExistingError{}, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
