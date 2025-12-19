package database

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickhouseStatsCalculator struct {
	sqlQuerier driver.Conn
}

func NewClickhouseStatsCalculator(sqlQuerier driver.Conn) *ClickhouseStatsCalculator {
	return &ClickhouseStatsCalculator{
		sqlQuerier: sqlQuerier,
	}
}

func (s *ClickhouseStatsCalculator) CalculateStatistics(ctx context.Context, urlToken string) (domain.CalculatedStatistics, error) {
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

func (s *ClickhouseStatsCalculator) getGroupCount(ctx context.Context, urlToken, groupBy string) (map[string]int, error) {
	result := make(map[string]int)
	query := fmt.Sprintf(`
		SELECT %s, COUNT() 
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
		var count uint64
		if err := rows.Scan(&groupName, &count); err != nil {
			return nil, err
		}

		result[groupName] = int(count)
	}
	rows.Close()

	return result, nil
}

func (s *ClickhouseStatsCalculator) getTotalClicks(ctx context.Context, urlToken string) (int, error) {
	var totalClicks uint64

	err := s.sqlQuerier.QueryRow(ctx, `
		SELECT COUNT() FROM stats_events WHERE url_token = $1
	`, urlToken).Scan(&totalClicks)
	if err != nil {
		return 0, err
	}

	if totalClicks == 0 {
		return 0, &domain.TokenNonExistingError{Msg: fmt.Sprintf("No statistics found for url token: %s", urlToken)}
	}

	return int(totalClicks), nil
}
