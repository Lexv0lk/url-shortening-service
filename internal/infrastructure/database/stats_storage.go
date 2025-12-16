package database

import (
	"context"
	"url-shortening-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgersStatsStorage struct {
	dbpool *pgxpool.Pool
	logger domain.Logger
}

func NewPostgresStatsStorage(dbpool *pgxpool.Pool, logger domain.Logger) *PostgersStatsStorage {
	return &PostgersStatsStorage{
		dbpool: dbpool,
		logger: logger,
	}
}

func (s *PostgersStatsStorage) AddStatsEvent(ctx context.Context, event domain.ProcessedStatsEvent) error {
	sql := `INSERT INTO stats_events (url_token, timestamp, country, city, device_type, referrer) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.dbpool.Exec(ctx, sql, event.UrlToken, event.Timestamp, event.Country, event.City, event.DeviceType, event.Referrer)
	if err != nil {
		s.logger.Error("Failed to add stats event to db: " + err.Error())
		return err
	}

	return nil
}
