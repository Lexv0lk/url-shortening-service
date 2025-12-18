package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeleteUrlHandler_Delete(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		urlToken       string
		expectedStatus int

		prepareMocks func(t *testing.T, ctrl *gomock.Controller) (domain.UrlDeleter, domain.Logger)
	}

	testCases := []testCase{
		{
			name:           "Success",
			urlToken:       "validToken",
			expectedStatus: http.StatusNoContent,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlDeleter, domain.Logger) {
				deleter := mocks.NewMockUrlDeleter(ctrl)
				deleter.EXPECT().DeleteUrl(gomock.Any(), "validToken").Return(nil)

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return deleter, logger
			},
		},
		{
			name:           "TokenNotFound",
			urlToken:       "missingToken",
			expectedStatus: http.StatusNotFound,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlDeleter, domain.Logger) {
				deleter := mocks.NewMockUrlDeleter(ctrl)
				deleter.EXPECT().DeleteUrl(gomock.Any(), "missingToken").Return(&domain.TokenNonExistingError{})

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return deleter, logger
			},
		},
		{
			name:           "InternalError",
			urlToken:       "errorToken",
			expectedStatus: http.StatusInternalServerError,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlDeleter, domain.Logger) {
				deleter := mocks.NewMockUrlDeleter(ctrl)
				deleter.EXPECT().DeleteUrl(gomock.Any(), "errorToken").Return(assert.AnError)

				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any())
				return deleter, logger
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			urlDeleterMock, loggerMock := tt.prepareMocks(t, ctrl)
			handler := NewDeleteUrlHandler(urlDeleterMock, loggerMock)

			req := httptest.NewRequest(http.MethodDelete, "/"+tt.urlToken, nil)
			req.SetPathValue(domain.UrlTokenStr, tt.urlToken)
			w := httptest.NewRecorder()

			handler.Delete(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
