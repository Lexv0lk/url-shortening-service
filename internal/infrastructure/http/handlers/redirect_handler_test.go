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

func TestRedirectHandler_Redirect(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		urlToken       string
		expectedStatus int
		expectedHeader string

		prepareMocks func(t *testing.T, ctrl *gomock.Controller) (domain.UrlGetter, domain.StatisticsSender, domain.Logger)
	}

	testCases := []testCase{
		{
			name:           "Success",
			urlToken:       "validToken",
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "https://example.com",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlGetter, domain.StatisticsSender, domain.Logger) {
				urlGetter := mocks.NewMockUrlGetter(ctrl)
				urlGetter.EXPECT().GetOriginalUrl(gomock.Any(), "validToken").Return("https://example.com", nil)

				statsSender := mocks.NewMockStatisticsSender(ctrl)
				statsSender.EXPECT().SendEvent(gomock.Any(), gomock.Any()).Return(nil)

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return urlGetter, statsSender, logger
			},
		},
		{
			name:           "SuccessWithStatsSendError",
			urlToken:       "validToken",
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "https://example.com",
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlGetter, domain.StatisticsSender, domain.Logger) {
				urlGetter := mocks.NewMockUrlGetter(ctrl)
				urlGetter.EXPECT().GetOriginalUrl(gomock.Any(), "validToken").Return("https://example.com", nil)

				statsSender := mocks.NewMockStatisticsSender(ctrl)
				statsSender.EXPECT().SendEvent(gomock.Any(), gomock.Any()).Return(assert.AnError)

				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().Warn(gomock.Any())
				return urlGetter, statsSender, logger
			},
		},
		{
			name:           "TokenNotFound",
			urlToken:       "missingToken",
			expectedStatus: http.StatusNotFound,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlGetter, domain.StatisticsSender, domain.Logger) {
				urlGetter := mocks.NewMockUrlGetter(ctrl)
				urlGetter.EXPECT().GetOriginalUrl(gomock.Any(), "missingToken").Return("", &domain.UrlNonExistingError{})

				statsSender := mocks.NewMockStatisticsSender(ctrl)
				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return urlGetter, statsSender, logger
			},
		},
		{
			name:           "InternalError",
			urlToken:       "errorToken",
			expectedStatus: http.StatusInternalServerError,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlGetter, domain.StatisticsSender, domain.Logger) {
				urlGetter := mocks.NewMockUrlGetter(ctrl)
				urlGetter.EXPECT().GetOriginalUrl(gomock.Any(), "errorToken").Return("", assert.AnError)

				statsSender := mocks.NewMockStatisticsSender(ctrl)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any())
				return urlGetter, statsSender, logger
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			urlGetterMock, statsSenderMock, loggerMock := tt.prepareMocks(t, ctrl)
			handler := NewRedirectHandler(urlGetterMock, statsSenderMock, loggerMock)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.urlToken, nil)
			req.SetPathValue(domain.UrlTokenStr, tt.urlToken)
			w := httptest.NewRecorder()

			handler.Redirect(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, w.Header().Get("Location"))
			}
		})
	}
}
