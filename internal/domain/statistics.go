package domain

import (
	"context"
	"time"
)

type RawStatsEvent struct {
	UrlToken  string    `json:"url_token"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referrer  string    `json:"referrer"`
}

type ProcessedStatsEvent struct {
	UrlToken   string
	Timestamp  time.Time
	Country    string
	City       string
	DeviceType string
	Referrer   string
}

type StatisticsProcessor interface {
	ProcessEvent(ctx context.Context, eventData []byte) error
}

type StatisticsStorage interface {
	AddStatsEvent(ctx context.Context, event ProcessedStatsEvent) error
}

type StatisticsSender interface {
	SendEvent(ctx context.Context, rawEvent RawStatsEvent) error
}
