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

func TestShortenUrlHandler_Create(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		requestBody    interface{}
		expectedStatus int

		prepareMocks func(t *testing.T, ctrl *gomock.Controller) (domain.UrlShortener, domain.Logger)
	}

	testCases := []testCase{
		{
			name:           "Success",
			requestBody:    ShortenUrlRequest{URL: "https://example.com"},
			expectedStatus: http.StatusCreated,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlShortener, domain.Logger) {
				urlShortener := mocks.NewMockUrlShortener(ctrl)
				urlShortener.EXPECT().ShortenUrl(gomock.Any(), "https://example.com").Return(domain.MappingInfo{
					Id:          1,
					OriginalURL: "https://example.com",
					Token:       "abc123",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil)

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return urlShortener, logger
			},
		},
		{
			name:           "InvalidRequestPayload",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlShortener, domain.Logger) {
				urlShortener := mocks.NewMockUrlShortener(ctrl)
				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return urlShortener, logger
			},
		},
		{
			name:           "InvalidUrl",
			requestBody:    ShortenUrlRequest{URL: "invalid-url"},
			expectedStatus: http.StatusBadRequest,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlShortener, domain.Logger) {
				urlShortener := mocks.NewMockUrlShortener(ctrl)
				urlShortener.EXPECT().ShortenUrl(gomock.Any(), "invalid-url").Return(domain.MappingInfo{}, &domain.InvalidUrlError{})

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return urlShortener, logger
			},
		},
		{
			name:           "InternalError",
			requestBody:    ShortenUrlRequest{URL: "https://example.com"},
			expectedStatus: http.StatusInternalServerError,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.UrlShortener, domain.Logger) {
				urlShortener := mocks.NewMockUrlShortener(ctrl)
				urlShortener.EXPECT().ShortenUrl(gomock.Any(), "https://example.com").Return(domain.MappingInfo{}, assert.AnError)

				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any())
				return urlShortener, logger
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			urlShortenerMock, loggerMock := tt.prepareMocks(t, ctrl)
			handler := NewAddUrlHandler(urlShortenerMock, loggerMock)

			var body []byte
			switch v := tt.requestBody.(type) {
			case string:
				body = []byte(v)
			default:
				body, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Create(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
