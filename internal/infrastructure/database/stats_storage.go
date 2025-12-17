package database

import (
	"context"
	"url-shortening-service/internal/domain"
)

// PostgersStatsStorage implements statistics event storage using PostgreSQL.
// It persists processed redirect statistics events for later analysis.
type PostgersStatsStorage struct {
	sqlExecutor domain.Executor
}

// NewPostgresStatsStorage creates a new PostgersStatsStorage instance.
// Parameters:
//   - sqlExecutor: PostgreSQL connection pool
func NewPostgresStatsStorage(sqlQuerier domain.Executor) *PostgersStatsStorage {
	return &PostgersStatsStorage{
		sqlExecutor: sqlQuerier,
	}
}

// AddStatsEvent persists a processed statistics event to PostgreSQL.
// It stores URL token, timestamp, country, city, device type, and referrer.
//
// Returns an error if the database operation fails.
func (s *PostgersStatsStorage) AddStatsEvent(ctx context.Context, event domain.ProcessedStatsEvent) error {
	sql := `INSERT INTO stats_events (url_token, timestamp, country, city, device_type, referrer) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.sqlExecutor.Exec(ctx, sql, event.UrlToken, event.Timestamp, event.Country, event.City, event.DeviceType, event.Referrer)
	if err != nil {
		return err
	}

	return nil
}
