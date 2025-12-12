package database

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	dbpool *pgxpool.Pool
	logger domain.Logger
}

func NewPostgresStorage(ctx context.Context, settings domain.PostgresSettings, logger domain.Logger) (*PostgresStorage, error) {
	connectionUrl := settings.GetUrl()

	dbpool, err := pgxpool.New(ctx, connectionUrl)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		dbpool: dbpool,
		logger: logger,
	}, nil
}

func (s *PostgresStorage) Close() {
	s.dbpool.Close()
}

func (s *PostgresStorage) GetMapping(ctx context.Context, urlToken string) (domain.MappingInfo, bool) {
	sql := `SELECT id, original_url, url_token FROM mappings WHERE url_token = $1`
	var mapping domain.MappingInfo

	err := s.dbpool.QueryRow(ctx, sql, urlToken).Scan(&mapping.Id, &mapping.OriginalURL, &mapping.Token)
	if err == pgx.ErrNoRows {
		return domain.MappingInfo{}, false
	} else if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get mapping from db: %v", err))
		return domain.MappingInfo{}, false
	}

	return mapping, true
}

func (s *PostgresStorage) AddNewMapping(ctx context.Context, id int64, originalUrl string, urlToken string) error {
	sql := `INSERT INTO mappings (id, original_url, url_token) VALUES ($1, $2, $3)`

	_, err := s.dbpool.Exec(ctx, sql, id, originalUrl, urlToken)
	if err != nil {
		return fmt.Errorf("failed to add new mapping to db: %w", err)
	}

	return nil
}

func (s *PostgresStorage) GetLastId(ctx context.Context) (int64, error) {
	sql := `SELECT id FROM mappings ORDER BY id DESC LIMIT 1`
	var lastId int64

	err := s.dbpool.QueryRow(ctx, sql).Scan(&lastId)
	if err == pgx.ErrNoRows {
		s.logger.Info("No existing mappings found in database.")
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to get last mapping id from db: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Retrieved last mapping id is: %d", lastId))

	return lastId, nil
}
