package redis

import (
	"context"
	"testing"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisIdGenerator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		expectedError error

		prepareMocks func(t *testing.T, ctrl *gomock.Controller) (domain.KeySetIncrementer, domain.MappingInfoLastIdGetter)
	}

	testCases := []testCase{
		{
			name: "successfully create RedisIdGenerator",

			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.KeySetIncrementer, domain.MappingInfoLastIdGetter) {
				t.Helper()
				clientMock := mocks.NewMockKeySetIncrementer(ctrl)
				lastIdGetterMock := mocks.NewMockMappingInfoLastIdGetter(ctrl)

				lastIdGetterMock.EXPECT().GetLastId(gomock.Any()).Return(int64(10), nil)
				clientMock.EXPECT().Set(gomock.Any(), counterId, int64(10), gomock.Any()).Return(redis.NewStatusCmd(context.Background()))

				return clientMock, lastIdGetterMock
			},
		},
		{
			name:          "error getting last id from storage",
			expectedError: assert.AnError,

			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.KeySetIncrementer, domain.MappingInfoLastIdGetter) {
				t.Helper()
				clientMock := mocks.NewMockKeySetIncrementer(ctrl)
				lastIdGetterMock := mocks.NewMockMappingInfoLastIdGetter(ctrl)

				lastIdGetterMock.EXPECT().GetLastId(gomock.Any()).Return(int64(0), assert.AnError)

				return clientMock, lastIdGetterMock
			},
		},
		{
			name:          "error setting initial counter in redis",
			expectedError: assert.AnError,

			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.KeySetIncrementer, domain.MappingInfoLastIdGetter) {
				t.Helper()
				clientMock := mocks.NewMockKeySetIncrementer(ctrl)
				lastIdGetterMock := mocks.NewMockMappingInfoLastIdGetter(ctrl)

				lastIdGetterMock.EXPECT().GetLastId(gomock.Any()).Return(int64(15), nil)
				clientMock.EXPECT().Set(gomock.Any(), counterId, int64(15), gomock.Any()).DoAndReturn(func(ctx context.Context, key string, value interface{}, expiration interface{}) *redis.StatusCmd {
					cmd := redis.NewStatusCmd(ctx)
					cmd.SetErr(assert.AnError)
					return cmd
				})

				return clientMock, lastIdGetterMock
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			clientMock, lastIdGetterMock := tt.prepareMocks(t, ctrl)
			_, err := NewRedisIdGenerator(context.Background(), clientMock, lastIdGetterMock)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedisIdGenerator_GetNextId(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		expectedRes   int64
		expectedError error

		prepareMocks func(t *testing.T, ctrl *gomock.Controller) (domain.KeySetIncrementer, domain.MappingInfoLastIdGetter)
	}

	testCases := []testCase{
		{
			name:        "successfully get next id",
			expectedRes: 6,

			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.KeySetIncrementer, domain.MappingInfoLastIdGetter) {
				t.Helper()
				clientMock := mocks.NewMockKeySetIncrementer(ctrl)
				lastIdGetterMock := mocks.NewMockMappingInfoLastIdGetter(ctrl)

				lastIdGetterMock.EXPECT().GetLastId(gomock.Any()).Return(int64(5), nil)
				clientMock.EXPECT().Set(gomock.Any(), counterId, int64(5), gomock.Any()).Return(redis.NewStatusCmd(context.Background()))
				clientMock.EXPECT().Incr(gomock.Any(), counterId).Return(redis.NewIntResult(6, nil))

				return clientMock, lastIdGetterMock
			},
		},
		{
			name:          "error incrementing id",
			expectedError: assert.AnError,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.KeySetIncrementer, domain.MappingInfoLastIdGetter) {
				t.Helper()
				clientMock := mocks.NewMockKeySetIncrementer(ctrl)
				lastIdGetterMock := mocks.NewMockMappingInfoLastIdGetter(ctrl)

				lastIdGetterMock.EXPECT().GetLastId(gomock.Any()).Return(int64(10), nil)
				clientMock.EXPECT().Set(gomock.Any(), counterId, int64(10), gomock.Any()).Return(redis.NewStatusCmd(context.Background()))
				clientMock.EXPECT().Incr(gomock.Any(), counterId).Return(redis.NewIntResult(0, assert.AnError))

				return clientMock, lastIdGetterMock
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			clientMock, lastIdGetterMock := tt.prepareMocks(t, ctrl)
			idGen, err := NewRedisIdGenerator(context.Background(), clientMock, lastIdGetterMock)
			require.NoError(t, err)

			res, err := idGen.GetNextId(context.Background())
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}
