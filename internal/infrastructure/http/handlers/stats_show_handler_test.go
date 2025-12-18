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

func TestStatsShowHandler_Show(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		urlToken       string
		expectedStatus int

		prepareMocks func(t *testing.T, ctrl *gomock.Controller) (domain.StatisticsCalculator, domain.Logger)
	}

	testCases := []testCase{
		{
			name:           "Success",
			urlToken:       "validToken",
			expectedStatus: http.StatusOK,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.StatisticsCalculator, domain.Logger) {
				statsCalculator := mocks.NewMockStatisticsCalculator(ctrl)
				statsCalculator.EXPECT().CalculateStatistics(gomock.Any(), "validToken").Return(domain.CalculatedStatistics{
					UrlToken:        "validToken",
					TotalClicks:     100,
					UniqueCountries: map[string]int{"US": 50, "DE": 30, "UA": 20},
					UniqueCities:    map[string]int{"New York": 30, "Berlin": 30, "Kyiv": 20, "Los Angeles": 20},
					DeviceTypeStats: map[string]int{"Desktop": 60, "Mobile": 40},
					ReferrerStats:   map[string]int{"google.com": 50, "twitter.com": 30, "direct": 20},
				}, nil)

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return statsCalculator, logger
			},
		},
		{
			name:           "TokenNotFound",
			urlToken:       "missingToken",
			expectedStatus: http.StatusNotFound,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.StatisticsCalculator, domain.Logger) {
				statsCalculator := mocks.NewMockStatisticsCalculator(ctrl)
				statsCalculator.EXPECT().CalculateStatistics(gomock.Any(), "missingToken").Return(domain.CalculatedStatistics{}, &domain.TokenNonExistingError{})

				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				return statsCalculator, logger
			},
		},
		{
			name:           "InternalError",
			urlToken:       "errorToken",
			expectedStatus: http.StatusInternalServerError,
			prepareMocks: func(t *testing.T, ctrl *gomock.Controller) (domain.StatisticsCalculator, domain.Logger) {
				statsCalculator := mocks.NewMockStatisticsCalculator(ctrl)
				statsCalculator.EXPECT().CalculateStatistics(gomock.Any(), "errorToken").Return(domain.CalculatedStatistics{}, assert.AnError)

				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any())
				return statsCalculator, logger
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			statsCalculatorMock, loggerMock := tt.prepareMocks(t, ctrl)
			handler := NewStatsShowHandler(statsCalculatorMock, loggerMock)

			req := httptest.NewRequest(http.MethodGet, "/stats/"+tt.urlToken, nil)
			req.SetPathValue(domain.UrlTokenStr, tt.urlToken)
			w := httptest.NewRecorder()

			handler.Show(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
