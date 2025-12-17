package urlcases

import (
	"context"
	"testing"
	"time"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUrlUpdater_UpdateUrlMapping(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2025, 12, 17, 10, 0, 0, 0, time.UTC)

	type testCase struct {
		name           string
		urlToken       string
		newOriginalUrl string
		expectedInfo   domain.MappingInfo
		expectedError  error

		setupMocks func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenSetter, domain.MappingInfoUpdater, domain.Logger)
	}

	testCases := []testCase{
		{
			name:           "successful url update",
			urlToken:       "abc123",
			newOriginalUrl: "https://example.com/new-url",
			expectedInfo: domain.MappingInfo{
				Id:          1,
				OriginalURL: "https://example.com/new-url",
				Token:       "abc123",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
			expectedError: nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenSetter, domain.MappingInfoUpdater, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenSetter(ctrl)
				storageMock := mocks.NewMockMappingInfoUpdater(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().UpdateOriginalUrl(gomock.Any(), "abc123", "https://example.com/new-url").Return(domain.MappingInfo{
					Id:          1,
					OriginalURL: "https://example.com/new-url",
					Token:       "abc123",
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime,
				}, nil)
				loggerMock.EXPECT().Info(gomock.Any()).AnyTimes()

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:           "successful url update with http scheme",
			urlToken:       "xyz789",
			newOriginalUrl: "http://example.com/http-url",
			expectedInfo: domain.MappingInfo{
				Id:          2,
				OriginalURL: "http://example.com/http-url",
				Token:       "xyz789",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
			expectedError: nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenSetter, domain.MappingInfoUpdater, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenSetter(ctrl)
				storageMock := mocks.NewMockMappingInfoUpdater(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().UpdateOriginalUrl(gomock.Any(), "xyz789", "http://example.com/http-url").Return(domain.MappingInfo{
					Id:          2,
					OriginalURL: "http://example.com/http-url",
					Token:       "xyz789",
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime,
				}, nil)
				loggerMock.EXPECT().Info(gomock.Any()).AnyTimes()

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:           "invalid new url returns error",
			urlToken:       "abc123",
			newOriginalUrl: "not-a-valid-url",
			expectedInfo:   domain.MappingInfo{},
			expectedError:  &domain.InvalidUrlError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenSetter, domain.MappingInfoUpdater, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenSetter(ctrl)
				storageMock := mocks.NewMockMappingInfoUpdater(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:           "empty new url returns error",
			urlToken:       "abc123",
			newOriginalUrl: "",
			expectedInfo:   domain.MappingInfo{},
			expectedError:  &domain.InvalidUrlError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenSetter, domain.MappingInfoUpdater, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenSetter(ctrl)
				storageMock := mocks.NewMockMappingInfoUpdater(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:           "unsupported scheme returns error",
			urlToken:       "abc123",
			newOriginalUrl: "ftp://example.com/file",
			expectedInfo:   domain.MappingInfo{},
			expectedError:  &domain.InvalidUrlError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenSetter, domain.MappingInfoUpdater, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenSetter(ctrl)
				storageMock := mocks.NewMockMappingInfoUpdater(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:           "token not found returns error",
			urlToken:       "nonexistent",
			newOriginalUrl: "https://example.com/valid-url",
			expectedInfo:   domain.MappingInfo{},
			expectedError:  &domain.TokenNonExistingError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenSetter, domain.MappingInfoUpdater, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenSetter(ctrl)
				storageMock := mocks.NewMockMappingInfoUpdater(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().UpdateOriginalUrl(gomock.Any(), "nonexistent", "https://example.com/valid-url").Return(domain.MappingInfo{}, &domain.TokenNonExistingError{Msg: "token not found"})

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:           "storage error returns error",
			urlToken:       "abc123",
			newOriginalUrl: "https://example.com/valid-url",
			expectedInfo:   domain.MappingInfo{},
			expectedError:  assert.AnError,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenSetter, domain.MappingInfoUpdater, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenSetter(ctrl)
				storageMock := mocks.NewMockMappingInfoUpdater(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().UpdateOriginalUrl(gomock.Any(), "abc123", "https://example.com/valid-url").Return(domain.MappingInfo{}, assert.AnError)

				return cacheMock, storageMock, loggerMock
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			cacheMock, storageMock, loggerMock := tt.setupMocks(t, ctrl)
			urlUpdater := NewUrlUpdater(cacheMock, storageMock, loggerMock)

			actualInfo, actualError := urlUpdater.UpdateUrlMapping(
				context.Background(),
				tt.urlToken,
				tt.newOriginalUrl,
			)

			if tt.expectedError != nil {
				assert.ErrorIs(t, actualError, tt.expectedError)
			} else {
				assert.NoError(t, actualError)
				assert.Equal(t, tt.expectedInfo, actualInfo)
			}
		})
	}
}
