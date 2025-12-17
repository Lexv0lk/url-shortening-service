package urlcases

import (
	"context"
	"testing"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUrlDeleter_DeleteUrl(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		urlToken      string
		expectedError error

		setupMocks func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenDeleter, domain.MappingInfoDeleter, domain.Logger)
	}

	testCases := []testCase{
		{
			name:          "successful deletion from storage and cache",
			urlToken:      "abc123",
			expectedError: nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenDeleter, domain.MappingInfoDeleter, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenDeleter(ctrl)
				storageMock := mocks.NewMockMappingInfoDeleter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().DeleteMappingInfo(gomock.Any(), "abc123").Return(nil)
				cacheMock.EXPECT().DeleteMapping(gomock.Any(), "abc123").Return(nil)

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:          "storage returns token not existing error",
			urlToken:      "nonexistent",
			expectedError: &domain.TokenNonExistingError{},
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenDeleter, domain.MappingInfoDeleter, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenDeleter(ctrl)
				storageMock := mocks.NewMockMappingInfoDeleter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().DeleteMappingInfo(gomock.Any(), "nonexistent").Return(&domain.TokenNonExistingError{Msg: "token not found"})

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:          "storage returns generic error",
			urlToken:      "abc123",
			expectedError: assert.AnError,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenDeleter, domain.MappingInfoDeleter, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenDeleter(ctrl)
				storageMock := mocks.NewMockMappingInfoDeleter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().DeleteMappingInfo(gomock.Any(), "abc123").Return(assert.AnError)

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:          "cache returns token not existing error logs warning and succeeds",
			urlToken:      "xyz789",
			expectedError: nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenDeleter, domain.MappingInfoDeleter, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenDeleter(ctrl)
				storageMock := mocks.NewMockMappingInfoDeleter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().DeleteMappingInfo(gomock.Any(), "xyz789").Return(nil)
				cacheMock.EXPECT().DeleteMapping(gomock.Any(), "xyz789").Return(&domain.TokenNonExistingError{Msg: "token not in cache"})
				loggerMock.EXPECT().Warn(gomock.Any()).AnyTimes()

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:          "cache returns generic error",
			urlToken:      "abc123",
			expectedError: assert.AnError,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenDeleter, domain.MappingInfoDeleter, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenDeleter(ctrl)
				storageMock := mocks.NewMockMappingInfoDeleter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().DeleteMappingInfo(gomock.Any(), "abc123").Return(nil)
				cacheMock.EXPECT().DeleteMapping(gomock.Any(), "abc123").Return(assert.AnError)

				return cacheMock, storageMock, loggerMock
			},
		},
		{
			name:          "empty token",
			urlToken:      "",
			expectedError: nil,
			setupMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlTokenDeleter, domain.MappingInfoDeleter, domain.Logger) {
				cacheMock := mocks.NewMockUrlTokenDeleter(ctrl)
				storageMock := mocks.NewMockMappingInfoDeleter(ctrl)
				loggerMock := mocks.NewMockLogger(ctrl)

				storageMock.EXPECT().DeleteMappingInfo(gomock.Any(), "").Return(nil)
				cacheMock.EXPECT().DeleteMapping(gomock.Any(), "").Return(nil)

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
			urlDeleter := NewUrlDeleter(cacheMock, storageMock, loggerMock)

			err := urlDeleter.DeleteUrl(context.Background(), tt.urlToken)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
