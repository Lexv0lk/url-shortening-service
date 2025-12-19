package database

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/domain/mocks"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStorage_GetMappingByToken(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		urlToken       string
		expectedResult domain.MappingInfo
		expectedFound  bool

		prepareMocks func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger
	}

	testCases := []testCase{
		{
			name:     "Success - mapping found",
			urlToken: "abc123",
			expectedResult: domain.MappingInfo{
				Id:          1,
				OriginalURL: "https://example.com",
				Token:       "abc123",
			},
			expectedFound: true,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				rows := pgxmock.NewRows([]string{"id", "original_url", "url_token"}).
					AddRow(int64(1), "https://example.com", "abc123")
				mockPool.ExpectQuery(`SELECT id, original_url, url_token FROM mappings WHERE url_token = \$1`).
					WithArgs("abc123").
					WillReturnRows(rows)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
			},
		},
		{
			name:           "Not found - returns empty mapping and false",
			urlToken:       "nonexistent",
			expectedResult: domain.MappingInfo{},
			expectedFound:  false,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectQuery(`SELECT id, original_url, url_token FROM mappings WHERE url_token = \$1`).
					WithArgs("nonexistent").
					WillReturnError(pgx.ErrNoRows)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
			},
		},
		{
			name:           "Database error - returns empty mapping and false",
			urlToken:       "abc123",
			expectedResult: domain.MappingInfo{},
			expectedFound:  false,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectQuery(`SELECT id, original_url, url_token FROM mappings WHERE url_token = \$1`).
					WithArgs("abc123").
					WillReturnError(assert.AnError)
				ctrl := gomock.NewController(t)
				mockLogger := mocks.NewMockLogger(ctrl)
				mockLogger.EXPECT().Error(gomock.Any())
				return mockLogger
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

			logger := tt.prepareMocks(t, mockPool)

			storage := NewPostgresStorage(mockPool, logger)
			result, found := storage.GetMappingByToken(context.Background(), tt.urlToken)

			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestPostgresStorage_AddNewMapping(t *testing.T) {
	t.Parallel()

	testTime := time.Date(2025, 12, 18, 10, 0, 0, 0, time.UTC)

	type testCase struct {
		name           string
		id             int64
		originalUrl    string
		urlToken       string
		expectedResult domain.MappingInfo
		expectedError  error

		prepareMocks func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger
	}

	testCases := []testCase{
		{
			name:        "Success - mapping created",
			id:          1,
			originalUrl: "https://example.com",
			urlToken:    "abc123",
			expectedResult: domain.MappingInfo{
				Id:          1,
				OriginalURL: "https://example.com",
				Token:       "abc123",
				CreatedAt:   testTime,
			},
			expectedError: nil,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				rows := pgxmock.NewRows([]string{"id", "original_url", "url_token", "created_at"}).
					AddRow(int64(1), "https://example.com", "abc123", testTime)
				mockPool.ExpectQuery(`INSERT INTO mappings`).
					WithArgs(int64(1), "https://example.com", "abc123").
					WillReturnRows(rows)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
			},
		},
		{
			name:           "Database error - returns error",
			id:             1,
			originalUrl:    "https://example.com",
			urlToken:       "abc123",
			expectedResult: domain.MappingInfo{},
			expectedError:  assert.AnError,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectQuery(`INSERT INTO mappings`).
					WithArgs(int64(1), "https://example.com", "abc123").
					WillReturnError(assert.AnError)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
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

			logger := tt.prepareMocks(t, mockPool)

			storage := NewPostgresStorage(mockPool, logger)
			result, err := storage.AddNewMapping(context.Background(), tt.id, tt.originalUrl, tt.urlToken)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestPostgresStorage_GetLastId(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		expectedId    int64
		expectedError error

		prepareMocks func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger
	}

	testCases := []testCase{
		{
			name:          "Success - returns last id",
			expectedId:    100,
			expectedError: nil,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				rows := pgxmock.NewRows([]string{"id"}).AddRow(int64(100))
				mockPool.ExpectQuery(`SELECT id FROM mappings ORDER BY id DESC LIMIT 1`).
					WillReturnRows(rows)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
			},
		},
		{
			name:          "No rows - returns 0",
			expectedId:    0,
			expectedError: nil,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectQuery(`SELECT id FROM mappings ORDER BY id DESC LIMIT 1`).
					WillReturnError(pgx.ErrNoRows)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
			},
		},
		{
			name:          "Database error - returns error",
			expectedId:    0,
			expectedError: assert.AnError,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectQuery(`SELECT id FROM mappings ORDER BY id DESC LIMIT 1`).
					WillReturnError(assert.AnError)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
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

			logger := tt.prepareMocks(t, mockPool)

			storage := NewPostgresStorage(mockPool, logger)
			result, err := storage.GetLastId(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedId, result)
			}
		})
	}
}

func TestPostgresStorage_UpdateOriginalUrl(t *testing.T) {
	t.Parallel()

	testCreatedTime := time.Date(2025, 12, 17, 10, 0, 0, 0, time.UTC)
	testUpdatedTime := time.Date(2025, 12, 18, 10, 0, 0, 0, time.UTC)

	type testCase struct {
		name           string
		urlToken       string
		newOriginalUrl string
		expectedResult domain.MappingInfo
		expectedError  error

		prepareMocks func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger
	}

	testCases := []testCase{
		{
			name:           "Success - url updated",
			urlToken:       "abc123",
			newOriginalUrl: "https://newexample.com",
			expectedResult: domain.MappingInfo{
				Id:          1,
				OriginalURL: "https://newexample.com",
				Token:       "abc123",
				CreatedAt:   testCreatedTime,
				UpdatedAt:   testUpdatedTime,
			},
			expectedError: nil,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				rows := pgxmock.NewRows([]string{"id", "original_url", "url_token", "created_at", "updated_at"}).
					AddRow(int64(1), "https://newexample.com", "abc123", testCreatedTime, testUpdatedTime)
				mockPool.ExpectQuery(`UPDATE mappings SET original_url`).
					WithArgs("https://newexample.com", pgxmock.AnyArg(), "abc123").
					WillReturnRows(rows)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
			},
		},
		{
			name:           "Token not found - returns TokenNonExistingError",
			urlToken:       "nonexistent",
			newOriginalUrl: "https://newexample.com",
			expectedResult: domain.MappingInfo{},
			expectedError:  &domain.TokenNonExistingError{},
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectQuery(`UPDATE mappings SET original_url`).
					WithArgs("https://newexample.com", pgxmock.AnyArg(), "nonexistent").
					WillReturnError(pgx.ErrNoRows)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
			},
		},
		{
			name:           "Database error - returns error",
			urlToken:       "abc123",
			newOriginalUrl: "https://newexample.com",
			expectedResult: domain.MappingInfo{},
			expectedError:  assert.AnError,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectQuery(`UPDATE mappings SET original_url`).
					WithArgs("https://newexample.com", pgxmock.AnyArg(), "abc123").
					WillReturnError(assert.AnError)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
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

			logger := tt.prepareMocks(t, mockPool)

			storage := NewPostgresStorage(mockPool, logger)
			result, err := storage.UpdateOriginalUrl(context.Background(), tt.urlToken, tt.newOriginalUrl)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if _, ok := tt.expectedError.(*domain.TokenNonExistingError); ok {
					assert.ErrorIs(t, err, &domain.TokenNonExistingError{})
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestPostgresStorage_DeleteMappingInfo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		urlToken      string
		expectedError error

		prepareMocks func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger
	}

	testCases := []testCase{
		{
			name:          "Success - mapping deleted",
			urlToken:      "abc123",
			expectedError: nil,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectExec(`DELETE FROM mappings WHERE url_token = \$1`).
					WithArgs("abc123").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				return slog.New(slog.NewTextHandler(io.Discard, nil))
			},
		},
		{
			name:          "Token not found - returns TokenNonExistingError",
			urlToken:      "nonexistent",
			expectedError: &domain.TokenNonExistingError{},
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectExec(`DELETE FROM mappings WHERE url_token = \$1`).
					WithArgs("nonexistent").
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
				return slog.New(slog.NewTextHandler(io.Discard, nil))
			},
		},
		{
			name:          "Database error - returns error",
			urlToken:      "abc123",
			expectedError: assert.AnError,
			prepareMocks: func(t *testing.T, mockPool pgxmock.PgxConnIface) domain.Logger {
				mockPool.ExpectExec(`DELETE FROM mappings WHERE url_token = \$1`).
					WithArgs("abc123").
					WillReturnError(assert.AnError)
				return slog.New(slog.NewTextHandler(io.Discard, nil))
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

			logger := tt.prepareMocks(t, mockPool)

			storage := NewPostgresStorage(mockPool, logger)
			err = storage.DeleteMappingInfo(context.Background(), tt.urlToken)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if _, ok := tt.expectedError.(*domain.TokenNonExistingError); ok {
					assert.ErrorIs(t, err, &domain.TokenNonExistingError{})
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
