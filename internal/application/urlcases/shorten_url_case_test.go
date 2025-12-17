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

func TestUrlShortener_ShortenUrl(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2025, 12, 17, 10, 0, 0, 0, time.UTC)

	type testCase struct {
		name                string
		originalUrl         string
		expectedMappingInfo domain.MappingInfo
		expectedError       error

		setupMocks func(t *testing.T, ctrl *gomock.Controller) (domain.IdGenerator, domain.MappingInfoAdder)
	}

	testCases := []testCase{
		{
			name:        "successful url shortening",
			originalUrl: "https://example.com/very-long-url",
			expectedMappingInfo: domain.MappingInfo{
				Id:          1,
				OriginalURL: "https://example.com/very-long-url",
				Token:       "b",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
			expectedError: nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.IdGenerator, domain.MappingInfoAdder) {
				idGenMock := mocks.NewMockIdGenerator(ctrl)
				storeMock := mocks.NewMockMappingInfoAdder(ctrl)

				idGenMock.EXPECT().GetNextId(gomock.Any()).Return(int64(1), nil)
				storeMock.EXPECT().AddNewMapping(gomock.Any(), int64(1), "https://example.com/very-long-url", "b").Return(domain.MappingInfo{
					Id:          1,
					OriginalURL: "https://example.com/very-long-url",
					Token:       "b",
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime,
				}, nil)

				return idGenMock, storeMock
			},
		},
		{
			name:        "successful url shortening with http scheme",
			originalUrl: "http://example.com/path",
			expectedMappingInfo: domain.MappingInfo{
				Id:          100,
				OriginalURL: "http://example.com/path",
				Token:       "bM",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
			expectedError: nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.IdGenerator, domain.MappingInfoAdder) {
				idGenMock := mocks.NewMockIdGenerator(ctrl)
				storeMock := mocks.NewMockMappingInfoAdder(ctrl)

				idGenMock.EXPECT().GetNextId(gomock.Any()).Return(int64(100), nil)
				storeMock.EXPECT().AddNewMapping(gomock.Any(), int64(100), "http://example.com/path", "bM").Return(domain.MappingInfo{
					Id:          100,
					OriginalURL: "http://example.com/path",
					Token:       "bM",
					CreatedAt:   fixedTime,
					UpdatedAt:   fixedTime,
				}, nil)

				return idGenMock, storeMock
			},
		},
		{
			name:                "invalid url returns error",
			originalUrl:         "not-a-valid-url",
			expectedMappingInfo: domain.MappingInfo{},
			expectedError:       &domain.InvalidUrlError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.IdGenerator, domain.MappingInfoAdder) {
				idGenMock := mocks.NewMockIdGenerator(ctrl)
				storeMock := mocks.NewMockMappingInfoAdder(ctrl)

				return idGenMock, storeMock
			},
		},
		{
			name:                "empty url returns error",
			originalUrl:         "",
			expectedMappingInfo: domain.MappingInfo{},
			expectedError:       &domain.InvalidUrlError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.IdGenerator, domain.MappingInfoAdder) {
				idGenMock := mocks.NewMockIdGenerator(ctrl)
				storeMock := mocks.NewMockMappingInfoAdder(ctrl)

				return idGenMock, storeMock
			},
		},
		{
			name:                "unsupported scheme returns error",
			originalUrl:         "ftp://example.com/file",
			expectedMappingInfo: domain.MappingInfo{},
			expectedError:       &domain.InvalidUrlError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.IdGenerator, domain.MappingInfoAdder) {
				idGenMock := mocks.NewMockIdGenerator(ctrl)
				storeMock := mocks.NewMockMappingInfoAdder(ctrl)

				return idGenMock, storeMock
			},
		},
		{
			name:                "id generator error returns error",
			originalUrl:         "https://example.com/valid-url",
			expectedMappingInfo: domain.MappingInfo{},
			expectedError:       assert.AnError,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.IdGenerator, domain.MappingInfoAdder) {
				idGenMock := mocks.NewMockIdGenerator(ctrl)
				storeMock := mocks.NewMockMappingInfoAdder(ctrl)

				idGenMock.EXPECT().GetNextId(gomock.Any()).Return(int64(0), assert.AnError)

				return idGenMock, storeMock
			},
		},
		{
			name:                "storage error returns error",
			originalUrl:         "https://example.com/another-url",
			expectedMappingInfo: domain.MappingInfo{},
			expectedError:       assert.AnError,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.IdGenerator, domain.MappingInfoAdder) {
				idGenMock := mocks.NewMockIdGenerator(ctrl)
				storeMock := mocks.NewMockMappingInfoAdder(ctrl)

				idGenMock.EXPECT().GetNextId(gomock.Any()).Return(int64(5), nil)
				storeMock.EXPECT().AddNewMapping(gomock.Any(), int64(5), "https://example.com/another-url", "f").Return(domain.MappingInfo{}, assert.AnError)

				return idGenMock, storeMock
			},
		},
		{
			name:                "url already exists returns error",
			originalUrl:         "https://example.com/existing-url",
			expectedMappingInfo: domain.MappingInfo{},
			expectedError:       &domain.UrlExistingError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.IdGenerator, domain.MappingInfoAdder) {
				idGenMock := mocks.NewMockIdGenerator(ctrl)
				storeMock := mocks.NewMockMappingInfoAdder(ctrl)

				idGenMock.EXPECT().GetNextId(gomock.Any()).Return(int64(10), nil)
				storeMock.EXPECT().AddNewMapping(gomock.Any(), int64(10), "https://example.com/existing-url", "k").Return(domain.MappingInfo{}, &domain.UrlExistingError{Msg: "url already exists"})

				return idGenMock, storeMock
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			idGenMock, storeMock := tt.setupMocks(t, ctrl)
			urlShortener := NewUrlShortener(idGenMock, storeMock)

			mappingInfo, err := urlShortener.ShortenUrl(context.Background(), tt.originalUrl)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMappingInfo, mappingInfo)
			}
		})
	}
}
