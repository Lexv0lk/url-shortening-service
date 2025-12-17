package database

import (
	"context"
	"testing"
	"url-shortening-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestGetGroupCount(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		urlToken       string
		groupByField   string
		expectedResult map[string]int
		expectedError  error

		prepareMocks func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	testCases := []testCase{
		{
			name:         "Success - returns country stats",
			urlToken:     "abc123",
			groupByField: "country",
			expectedResult: map[string]int{
				"USA":     10,
				"Germany": 5,
				"France":  3,
			},
			expectedError: nil,
			prepareMocks: func(t *testing.T, mock pgxmock.PgxConnIface) {
				rows := pgxmock.NewRows([]string{"country", "count"}).
					AddRow("USA", 10).
					AddRow("Germany", 5).
					AddRow("France", 3)
				mock.ExpectQuery(`SELECT country, COUNT\(\*\) FROM stats_events WHERE url_token = \$1 GROUP BY country`).
					WithArgs("abc123").
					WillReturnRows(rows)
			},
		},
		{
			name:           "Success - returns empty map when no rows",
			urlToken:       "abc123",
			groupByField:   "city",
			expectedResult: map[string]int{},
			expectedError:  nil,
			prepareMocks: func(t *testing.T, mock pgxmock.PgxConnIface) {
				rows := pgxmock.NewRows([]string{"city", "count"})
				mock.ExpectQuery(`SELECT city, COUNT\(\*\) FROM stats_events WHERE url_token = \$1 GROUP BY city`).
					WithArgs("abc123").
					WillReturnRows(rows)
			},
		},
		{
			name:           "Database query error - returns error",
			urlToken:       "abc123",
			groupByField:   "device_type",
			expectedResult: nil,
			expectedError:  assert.AnError,
			prepareMocks: func(t *testing.T, mock pgxmock.PgxConnIface) {
				mock.ExpectQuery(`SELECT device_type, COUNT\(\*\) FROM stats_events WHERE url_token = \$1 GROUP BY device_type`).
					WithArgs("abc123").
					WillReturnError(assert.AnError)
			},
		},
		{
			name:           "Scan error - returns error",
			urlToken:       "abc123",
			groupByField:   "referrer",
			expectedResult: nil,
			expectedError:  assert.AnError,
			prepareMocks: func(t *testing.T, mock pgxmock.PgxConnIface) {
				rows := pgxmock.NewRows([]string{"referrer", "count"}).
					AddRow("google.com", 10).
					RowError(0, assert.AnError)
				mock.ExpectQuery(`SELECT referrer, COUNT\(\*\) FROM stats_events WHERE url_token = \$1 GROUP BY referrer`).
					WithArgs("abc123").
					WillReturnRows(rows)
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock, err := pgxmock.NewConn()
			if err != nil {
				t.Fatalf("failed to create pgxmock: %v", err)
			}
			defer mock.Close(context.Background())

			tt.prepareMocks(t, mock)

			calculator := NewPostgresStatsCalculator(mock)
			result, err := calculator.getGroupCount(context.Background(), tt.urlToken, tt.groupByField)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestGetTotalClicks(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		urlToken       string
		expectedClicks int
		expectedError  error

		prepareMocks func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	testCases := []testCase{
		{
			name:           "Success - returns total clicks",
			urlToken:       "abc123",
			expectedClicks: 42,
			expectedError:  nil,
			prepareMocks: func(t *testing.T, mock pgxmock.PgxConnIface) {
				rows := pgxmock.NewRows([]string{"count"}).AddRow(42)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stats_events WHERE url_token = \$1`).
					WithArgs("abc123").
					WillReturnRows(rows)
			},
		},
		{
			name:           "Token does not exist - zero clicks returns TokenNonExistingError",
			urlToken:       "nonexistent",
			expectedClicks: 0,
			expectedError:  &domain.TokenNonExistingError{},
			prepareMocks: func(t *testing.T, mock pgxmock.PgxConnIface) {
				rows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stats_events WHERE url_token = \$1`).
					WithArgs("nonexistent").
					WillReturnRows(rows)
			},
		},
		{
			name:           "Database error - returns error",
			urlToken:       "abc123",
			expectedClicks: 0,
			expectedError:  assert.AnError,
			prepareMocks: func(t *testing.T, mock pgxmock.PgxConnIface) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stats_events WHERE url_token = \$1`).
					WithArgs("abc123").
					WillReturnError(assert.AnError)
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock, err := pgxmock.NewConn()
			if err != nil {
				t.Fatalf("failed to create pgxmock: %v", err)
			}
			defer mock.Close(context.Background())

			tt.prepareMocks(t, mock)

			calculator := NewPostgresStatsCalculator(mock)
			clicks, err := calculator.getTotalClicks(context.Background(), tt.urlToken)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedClicks, clicks)
			}
		})
	}
}

func TestCalculateStatistics_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctr, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies())
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(ctr) })

	baseConnStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	dbPool, err := pgxpool.New(ctx, baseConnStr)
	require.NoError(t, err)
	t.Cleanup(dbPool.Close)

	type testCase struct {
		name          string
		urlToken      string
		expectedStats domain.CalculatedStatistics
		expectedError error

		prepareData func(t *testing.T, pool *pgxpool.Pool, urlToken string)
	}

	testCases := []testCase{
		{
			name:     "Success_returns_calculated_statistics",
			urlToken: "abc123",
			expectedStats: domain.CalculatedStatistics{
				UrlToken:    "abc123",
				TotalClicks: 3,
				UniqueCountries: map[string]int{
					"USA":     2,
					"Germany": 1,
				},
				UniqueCities: map[string]int{
					"New York": 1,
					"Boston":   1,
					"Berlin":   1,
				},
				DeviceTypeStats: map[string]int{
					"desktop": 2,
					"mobile":  1,
				},
				ReferrerStats: map[string]int{
					"google.com":   2,
					"facebook.com": 1,
				},
			},
			expectedError: nil,
			prepareData: func(t *testing.T, pool *pgxpool.Pool, urlToken string) {
				t.Helper()
				_, err := pool.Exec(ctx, `
					INSERT INTO stats_events (url_token, country, city, device_type, referrer) VALUES
					($1, 'USA', 'New York', 'desktop', 'google.com'),
					($1, 'USA', 'Boston', 'mobile', 'facebook.com'),
					($1, 'Germany', 'Berlin', 'desktop', 'google.com')
				`, urlToken)
				require.NoError(t, err)
			},
		},
		{
			name:          "Token_does_not_exist_returns_TokenNonExistingError",
			urlToken:      "nonexistent",
			expectedStats: domain.CalculatedStatistics{},
			expectedError: &domain.TokenNonExistingError{},
			prepareData: func(t *testing.T, pool *pgxpool.Pool, urlToken string) {
				// No data inserted - token does not exist
			},
		},
		{
			name:     "Success_single_event_returns_statistics",
			urlToken: "single123",
			expectedStats: domain.CalculatedStatistics{
				UrlToken:    "single123",
				TotalClicks: 1,
				UniqueCountries: map[string]int{
					"France": 1,
				},
				UniqueCities: map[string]int{
					"Paris": 1,
				},
				DeviceTypeStats: map[string]int{
					"tablet": 1,
				},
				ReferrerStats: map[string]int{
					"twitter.com": 1,
				},
			},
			expectedError: nil,
			prepareData: func(t *testing.T, pool *pgxpool.Pool, urlToken string) {
				t.Helper()
				_, err := pool.Exec(ctx, `
					INSERT INTO stats_events (url_token, country, city, device_type, referrer) VALUES
					($1, 'France', 'Paris', 'tablet', 'twitter.com')
				`, urlToken)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema := tt.name
			_, err := dbPool.Exec(ctx, "CREATE SCHEMA "+schema)
			require.NoError(t, err)
			t.Cleanup(func() {
				_, _ = dbPool.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
			})

			schemaPool := newPoolWithSchema(t, ctx, schema, baseConnStr)
			t.Cleanup(schemaPool.Close)

			_, err = schemaPool.Exec(ctx, `
			CREATE TABLE stats_events (
				id              BIGSERIAL PRIMARY KEY,
				url_token       TEXT NOT NULL,
				timestamp       TIMESTAMP WITH TIME ZONE,
				country         TEXT,
				city            TEXT,
				device_type     TEXT,
				referrer        TEXT
			);`)
			require.NoError(t, err)

			tt.prepareData(t, schemaPool, tt.urlToken)

			calculator := NewPostgresStatsCalculator(schemaPool)
			stats, err := calculator.CalculateStatistics(ctx, tt.urlToken)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStats, stats)
			}
		})
	}
}

func newPoolWithSchema(t *testing.T, ctx context.Context, schema, baseConnStr string) *pgxpool.Pool {
	t.Helper()

	cfg, err := pgxpool.ParseConfig(baseConnStr)
	require.NoError(t, err)

	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO "+schema)
		return err
	}

	schemaPool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	return schemaPool
}
