package database

import (
	"context"
	"fmt"
	"time"
	"url-shortening-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	dbpool *pgxpool.Pool
	logger domain.Logger
}

func NewPostgresStorage(dbpool *pgxpool.Pool, logger domain.Logger) *PostgresStorage {
	return &PostgresStorage{
		dbpool: dbpool,
		logger: logger,
	}
}

func (s *PostgresStorage) GetMappingByToken(ctx context.Context, urlToken string) (domain.MappingInfo, bool) {
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

func (s *PostgresStorage) AddNewMapping(ctx context.Context, id int64, originalUrl string, urlToken string) (domain.MappingInfo, error) {
	sql := `INSERT INTO mappings (id, original_url, url_token) VALUES ($1, $2, $3) RETURNING id, original_url, url_token, created_at`
	var result domain.MappingInfo

	err := s.dbpool.QueryRow(ctx, sql, id, originalUrl, urlToken).Scan(&result.Id, &result.OriginalURL, &result.Token, &result.CreatedAt)
	if err != nil {
		return domain.MappingInfo{}, fmt.Errorf("failed to add new mapping to db: %w", err)
	}

	return result, nil
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

func (s *PostgresStorage) UpdateOriginalUrl(ctx context.Context, urlToken string, newOriginalUrl string) (domain.MappingInfo, error) {
	sql := `UPDATE mappings SET original_url = $1, updated_at = $2 WHERE url_token = $3 RETURNING id, original_url, url_token, created_at, updated_at`
	var updatedMapping domain.MappingInfo

	err := s.dbpool.QueryRow(ctx, sql, newOriginalUrl, time.Now(), urlToken).Scan(&updatedMapping.Id, &updatedMapping.OriginalURL, &updatedMapping.Token, &updatedMapping.CreatedAt, &updatedMapping.UpdatedAt)
	if err == pgx.ErrNoRows {
		return domain.MappingInfo{}, &domain.TokenNonExistingError{Msg: fmt.Sprintf("No mapping with token %s found", urlToken)}
	} else if err != nil {
		return domain.MappingInfo{}, fmt.Errorf("failed to update original URL in db: %w", err)
	}

	return updatedMapping, nil
}

func (s *PostgresStorage) DeleteMappingInfo(ctx context.Context, urlToken string) error {
	sql := `DELETE FROM mappings WHERE url_token = $1`

	cmdTag, err := s.dbpool.Exec(ctx, sql, urlToken)
	if err != nil {
		return fmt.Errorf("failed to delete mapping from db: %w", err)
	} else if cmdTag.RowsAffected() == 0 {
		return &domain.TokenNonExistingError{Msg: fmt.Sprintf("No mapping with token %s found", urlToken)}
	}

	return nil
}
