package database

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"
)

// PostgresStatsCalculator calculates aggregated statistics from PostgreSQL.
// It provides statistics about URL access patterns including geographic and device data.
type PostgresStatsCalculator struct {
	sqlQuerier domain.Querier
}

// NewPostgresStatsCalculator creates a new PostgresStatsCalculator instance.
// Parameters:
//   - sqlExecutor: PostgreSQL connection pool
//   - logger: logger for recording errors
func NewPostgresStatsCalculator(sqlQuerier domain.Querier) *PostgresStatsCalculator {
	return &PostgresStatsCalculator{
		sqlQuerier: sqlQuerier,
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
	var err error

	stats.TotalClicks, err = s.getTotalClicks(ctx, urlToken)
	if err != nil {
		return domain.CalculatedStatistics{}, err
	}

	stats.UniqueCountries, err = s.getGroupCount(ctx, urlToken, "country")
	if err != nil {
		return domain.CalculatedStatistics{}, err
	}

	stats.UniqueCities, err = s.getGroupCount(ctx, urlToken, "city")
	if err != nil {
		return domain.CalculatedStatistics{}, err
	}

	stats.DeviceTypeStats, err = s.getGroupCount(ctx, urlToken, "device_type")
	if err != nil {
		return domain.CalculatedStatistics{}, err
	}

	stats.ReferrerStats, err = s.getGroupCount(ctx, urlToken, "referrer")
	if err != nil {
		return domain.CalculatedStatistics{}, err
	}

	return stats, nil
}

func (s *PostgresStatsCalculator) getGroupCount(ctx context.Context, urlToken, groupBy string) (map[string]int, error) {
	result := make(map[string]int)
	query := fmt.Sprintf(`
		SELECT %s, COUNT(*) 
		FROM stats_events 
		WHERE url_token = $1 
		GROUP BY %s
	`, groupBy, groupBy)

	rows, err := s.sqlQuerier.Query(ctx, query, urlToken)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var groupName string
		var count int
		if err := rows.Scan(&groupName, &count); err != nil {
			return nil, err
		}

		result[groupName] = count
	}
	rows.Close()

	return result, nil
}

func (s *PostgresStatsCalculator) getTotalClicks(ctx context.Context, urlToken string) (int, error) {
	var totalClicks int

	err := s.sqlQuerier.QueryRow(ctx, `
		SELECT COUNT(*) FROM stats_events WHERE url_token = $1
	`, urlToken).Scan(&totalClicks)
	if err != nil {
		return 0, err
	}

	if totalClicks == 0 {
		return 0, &domain.TokenNonExistingError{Msg: fmt.Sprintf("No statistics found for url token: %s", urlToken)}
	}

	return totalClicks, nil
}
