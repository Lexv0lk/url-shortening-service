package database

import (
	"context"
	"url-shortening-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgersStatsStorage implements statistics event storage using PostgreSQL.
// It persists processed redirect statistics events for later analysis.
type PostgersStatsStorage struct {
	dbpool *pgxpool.Pool
	logger domain.Logger
}

// NewPostgresStatsStorage creates a new PostgersStatsStorage instance.
// Parameters:
//   - dbpool: PostgreSQL connection pool
//   - logger: logger for recording errors
func NewPostgresStatsStorage(dbpool *pgxpool.Pool, logger domain.Logger) *PostgersStatsStorage {
	return &PostgersStatsStorage{
		dbpool: dbpool,
		logger: logger,
	}
}

// AddStatsEvent persists a processed statistics event to PostgreSQL.
// It stores URL token, timestamp, country, city, device type, and referrer.
//
// Returns an error if the database operation fails.
func (s *PostgersStatsStorage) AddStatsEvent(ctx context.Context, event domain.ProcessedStatsEvent) error {
	sql := `INSERT INTO stats_events (url_token, timestamp, country, city, device_type, referrer) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.dbpool.Exec(ctx, sql, event.UrlToken, event.Timestamp, event.Country, event.City, event.DeviceType, event.Referrer)
	if err != nil {
		s.logger.Error("Failed to add stats event to db: " + err.Error())
		return err
	}

	return nil
}
