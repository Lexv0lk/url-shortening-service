package database

import (
	"context"
	"testing"
	"time"
	"url-shortening-service/internal/domain"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgersStatsStorage_AddStatsEvent(t *testing.T) {
	t.Parallel()

	testTime := time.Date(2025, 12, 18, 10, 0, 0, 0, time.UTC)

	type testCase struct {
		name          string
		event         domain.ProcessedStatsEvent
		expectedError error

		prepareMocks func(t *testing.T, mockPool pgxmock.PgxConnIface)
	}

	testCases := []testCase{
		{
			name: "Success - event inserted",
			event: domain.ProcessedStatsEvent{
				UrlToken:   "abc123",
				Timestamp:  testTime,
				Country:    "USA",
				City:       "New York",
				DeviceType: "desktop",
				Referrer:   "google.com",
			},
			expectedError: nil,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) {
				mockPool.ExpectExec(`INSERT INTO stats_events`).
					WithArgs("abc123", testTime, "USA", "New York", "desktop", "google.com").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name: "Database error - returns error",
			event: domain.ProcessedStatsEvent{
				UrlToken:   "abc123",
				Timestamp:  testTime,
				Country:    "Germany",
				City:       "Berlin",
				DeviceType: "mobile",
				Referrer:   "facebook.com",
			},
			expectedError: assert.AnError,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) {
				mockPool.ExpectExec(`INSERT INTO stats_events`).
					WithArgs("abc123", testTime, "Germany", "Berlin", "mobile", "facebook.com").
					WillReturnError(assert.AnError)
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockPool, err := pgxmock.NewConn()
			require.NoError(t, err)
			defer mockPool.Close(context.Background())

			tt.prepareMocks(t, mockPool)

			storage := NewPostgresStatsStorage(mockPool)
			err = storage.AddStatsEvent(context.Background(), tt.event)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
