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

func TestUrlGetter_GetOriginalUrl(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name                string
		urlToken            string
		expectedOriginalUrl string
		expectedError       error

		setupMocks func(t *testing.T, ctrl *gomock.Controller) (domain.MappedGetSetter, domain.MappingInfoGetter, domain.Logger)
	}

	testCases := []testCase{
		{
			name:                "cache hit returns original url",
			urlToken:            "abc123",
			expectedOriginalUrl: "https://example.com/long-url",
			expectedError:       nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.MappedGetSetter, domain.MappingInfoGetter, domain.Logger) {
				cacheMock := mocks.NewMockMappedGetSetter(ctrl)
				storeMock := mocks.NewMockMappingInfoGetter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				cacheMock.EXPECT().GetOriginalUrl(gomock.Any(), "abc123").Return("https://example.com/long-url", true)

				return cacheMock, storeMock, loggerMock
			},
		},
		{
			name:                "cache miss then storage hit and cache populated",
			urlToken:            "xyz789",
			expectedOriginalUrl: "https://example.com/another-url",
			expectedError:       nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.MappedGetSetter, domain.MappingInfoGetter, domain.Logger) {
				cacheMock := mocks.NewMockMappedGetSetter(ctrl)
				storeMock := mocks.NewMockMappingInfoGetter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				cacheMock.EXPECT().GetOriginalUrl(gomock.Any(), "xyz789").Return("", false)
				storeMock.EXPECT().GetMappingByToken(gomock.Any(), "xyz789").Return(domain.MappingInfo{
					Id:          1,
					OriginalURL: "https://example.com/another-url",
					Token:       "xyz789",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, true)
				cacheMock.EXPECT().SetMapping(gomock.Any(), "https://example.com/another-url", "xyz789").Return(nil)

				return cacheMock, storeMock, loggerMock
			},
		},
		{
			name:                "cache miss and storage miss returns error",
			urlToken:            "nonexistent",
			expectedOriginalUrl: "",
			expectedError:       &domain.UrlNonExistingError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.MappedGetSetter, domain.MappingInfoGetter, domain.Logger) {
				cacheMock := mocks.NewMockMappedGetSetter(ctrl)
				storeMock := mocks.NewMockMappingInfoGetter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				cacheMock.EXPECT().GetOriginalUrl(gomock.Any(), "nonexistent").Return("", false)
				storeMock.EXPECT().GetMappingByToken(gomock.Any(), "nonexistent").Return(domain.MappingInfo{}, false)

				return cacheMock, storeMock, loggerMock
			},
		},
		{
			name:                "cache miss storage hit but cache set fails logs warning and returns url",
			urlToken:            "def456",
			expectedOriginalUrl: "https://example.com/cached-fail-url",
			expectedError:       nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.MappedGetSetter, domain.MappingInfoGetter, domain.Logger) {
				cacheMock := mocks.NewMockMappedGetSetter(ctrl)
				storeMock := mocks.NewMockMappingInfoGetter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				cacheMock.EXPECT().GetOriginalUrl(gomock.Any(), "def456").Return("", false)
				storeMock.EXPECT().GetMappingByToken(gomock.Any(), "def456").Return(domain.MappingInfo{
					Id:          2,
					OriginalURL: "https://example.com/cached-fail-url",
					Token:       "def456",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, true)
				cacheMock.EXPECT().SetMapping(gomock.Any(), "https://example.com/cached-fail-url", "def456").Return(assert.AnError)
				loggerMock.EXPECT().Warn(gomock.Any()).AnyTimes()

				return cacheMock, storeMock, loggerMock
			},
		},
		{
			name:                "empty token cache miss and storage miss",
			urlToken:            "",
			expectedOriginalUrl: "",
			expectedError:       &domain.UrlNonExistingError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.MappedGetSetter, domain.MappingInfoGetter, domain.Logger) {
				cacheMock := mocks.NewMockMappedGetSetter(ctrl)
				storeMock := mocks.NewMockMappingInfoGetter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				cacheMock.EXPECT().GetOriginalUrl(gomock.Any(), "").Return("", false)
				storeMock.EXPECT().GetMappingByToken(gomock.Any(), "").Return(domain.MappingInfo{}, false)

				return cacheMock, storeMock, loggerMock
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			cacheMock, storeMock, loggerMock := tt.setupMocks(t, ctrl)
			urlGetter := NewUrlGetter(cacheMock, storeMock, loggerMock)

			originalUrl, err := urlGetter.GetOriginalUrl(context.Background(), tt.urlToken)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOriginalUrl, originalUrl)
			}
		})
	}
}
