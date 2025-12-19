package database

import (
	"context"
	"url-shortening-service/internal/domain"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickhouseStatsStorage struct {
	conn driver.Conn
}

func NewClickhouseStatsStorage(conn driver.Conn) *ClickhouseStatsStorage {
	return &ClickhouseStatsStorage{
		conn: conn,
	}
}

func (s *ClickhouseStatsStorage) AddStatsEvent(ctx context.Context, event domain.ProcessedStatsEvent) error {
	req := `INSERT INTO stats_events (url_token, timestamp, country, city, device_type, referrer)`

	batch, err := s.conn.PrepareBatch(ctx, req)
	if err != nil {
		return err
	}
	defer batch.Close()

	err = batch.Append(
		event.UrlToken,
		event.Timestamp,
		event.Country,
		event.City,
		event.DeviceType,
		event.Referrer,
	)
	if err != nil {
		return err
	}

	return batch.Send()
}
