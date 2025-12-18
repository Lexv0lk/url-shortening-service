package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateUrlHandler_Update(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		urlToken       string
		requestBody    interface{}
		expectedStatus int

		prepareMocks func(t *testing.T, ctrl *gomock.Controller) (domain.UrlUpdater, domain.Logger)
	}

	testCases := []testCase{
		{
			name:           "Success",
			urlToken:       "validToken",
			requestBody:    UpdateUrlRequest{NewURL: "https://newexample.com"},
			expectedStatus: http.StatusOK,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlUpdater, domain.Logger) {
				urlUpdater := mocks.NewMockUrlUpdater(ctrl)
				urlUpdater.EXPECT().UpdateUrlMapping(gomock.Any(), "validToken", "https://newexample.com").Return(domain.MappingInfo{
					Id:          1,
					OriginalURL: "https://newexample.com",
					Token:       "validToken",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil)

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return urlUpdater, logger
			},
		},
		{
			name:           "InvalidRequestPayload",
			urlToken:       "validToken",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlUpdater, domain.Logger) {
				urlUpdater := mocks.NewMockUrlUpdater(ctrl)
				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return urlUpdater, logger
			},
		},
		{
			name:           "InvalidUrl",
			urlToken:       "validToken",
			requestBody:    UpdateUrlRequest{NewURL: "invalid-url"},
			expectedStatus: http.StatusBadRequest,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlUpdater, domain.Logger) {
				urlUpdater := mocks.NewMockUrlUpdater(ctrl)
				urlUpdater.EXPECT().UpdateUrlMapping(gomock.Any(), "validToken", "invalid-url").Return(domain.MappingInfo{}, &domain.InvalidUrlError{})

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return urlUpdater, logger
			},
		},
		{
			name:           "TokenNotFound",
			urlToken:       "missingToken",
			requestBody:    UpdateUrlRequest{NewURL: "https://newexample.com"},
			expectedStatus: http.StatusNotFound,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlUpdater, domain.Logger) {
				urlUpdater := mocks.NewMockUrlUpdater(ctrl)
				urlUpdater.EXPECT().UpdateUrlMapping(gomock.Any(), "missingToken", "https://newexample.com").Return(domain.MappingInfo{}, &domain.TokenNonExistingError{})

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return urlUpdater, logger
			},
		},
		{
			name:           "InternalError",
			urlToken:       "errorToken",
			requestBody:    UpdateUrlRequest{NewURL: "https://newexample.com"},
			expectedStatus: http.StatusInternalServerError,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlUpdater, domain.Logger) {
				urlUpdater := mocks.NewMockUrlUpdater(ctrl)
				urlUpdater.EXPECT().UpdateUrlMapping(gomock.Any(), "errorToken", "https://newexample.com").Return(domain.MappingInfo{}, assert.AnError)

				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any())
				return urlUpdater, logger
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			urlUpdaterMock, loggerMock := tt.prepareMocks(t, ctrl)
			handler := NewUpdateUrlHandler(urlUpdaterMock, loggerMock)

			var body []byte
			switch v := tt.requestBody.(type) {
			case string:
				body = []byte(v)
			default:
				body, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPut, "/"+tt.urlToken, bytes.NewReader(body))
			req.SetPathValue(domain.UrlTokenStr, tt.urlToken)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Update(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
