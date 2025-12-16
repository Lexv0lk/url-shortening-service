package database

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStatsCalculator calculates aggregated statistics from PostgreSQL.
// It provides statistics about URL access patterns including geographic and device data.
type PostgresStatsCalculator struct {
	dbpool *pgxpool.Pool
	logger domain.Logger
}

// NewPostgresStatsCalculator creates a new PostgresStatsCalculator instance.
// Parameters:
//   - dbpool: PostgreSQL connection pool
//   - logger: logger for recording errors
func NewPostgresStatsCalculator(dbpool *pgxpool.Pool, logger domain.Logger) *PostgresStatsCalculator {
	return &PostgresStatsCalculator{
		dbpool: dbpool,
		logger: logger,
	}
}

// CalculateStatistics computes aggregated statistics for a given URL token.
// It returns total clicks, country distribution, city distribution,
// device type breakdown, and referrer statistics.
//
// Returns an error if:
//   - *domain.TokenNonExistingError: no statistics exist for the given token
//   - Database query fails
func (s *PostgresStatsCalculator) CalculateStatistics(ctx context.Context, urlToken string) (domain.CalculatedStatistics, error) {
	stats := domain.CalculatedStatistics{
		UrlToken:        urlToken,
		UniqueCountries: make(map[string]int),
		UniqueCities:    make(map[string]int),
		DeviceTypeStats: make(map[string]int),
		ReferrerStats:   make(map[string]int),
	}

	err := s.dbpool.QueryRow(ctx, `
		SELECT COUNT(*) FROM stats_events WHERE url_token = $1
	`, urlToken).Scan(&stats.TotalClicks)
	if err != nil {
		return stats, err
	}

	if stats.TotalClicks == 0 {
		return domain.CalculatedStatistics{}, &domain.TokenNonExistingError{Msg: fmt.Sprintf("No statistics found for url token: %s", urlToken)}
	}

	rows, err := s.dbpool.Query(ctx, `
		SELECT country, COUNT(*) 
		FROM stats_events 
		WHERE url_token = $1 
		GROUP BY country
	`, urlToken)
	if err != nil {
		return stats, err
	}
	for rows.Next() {
		var country string
		var count int
		if err := rows.Scan(&country, &count); err != nil {
			return stats, err
		}
		stats.UniqueCountries[country] = count
	}
	rows.Close()

	rows, err = s.dbpool.Query(ctx, `
		SELECT city, COUNT(*) 
		FROM stats_events 
		WHERE url_token = $1 
		GROUP BY city
	`, urlToken)
	if err != nil {
		return stats, err
	}
	for rows.Next() {
		var city string
		var count int
		if err := rows.Scan(&city, &count); err != nil {
			return stats, err
		}
		stats.UniqueCities[city] = count
	}
	rows.Close()

	rows, err = s.dbpool.Query(ctx, `
		SELECT device_type, COUNT(*) 
		FROM stats_events 
		WHERE url_token = $1 
		GROUP BY device_type
	`, urlToken)
	if err != nil {
		return stats, err
	}
	for rows.Next() {
		var deviceType string
		var count int
		if err := rows.Scan(&deviceType, &count); err != nil {
			return stats, err
		}
		stats.DeviceTypeStats[deviceType] = count
	}
	rows.Close()

	rows, err = s.dbpool.Query(ctx, `
		SELECT referrer, COUNT(*) 
		FROM stats_events 
		WHERE url_token = $1 
		GROUP BY referrer
	`, urlToken)
	if err != nil {
		return stats, err
	}
	for rows.Next() {
		var referrer string
		var count int
		if err := rows.Scan(&referrer, &count); err != nil {
			return stats, err
		}
		stats.ReferrerStats[referrer] = count
	}
	rows.Close()

	return stats, nil
}
