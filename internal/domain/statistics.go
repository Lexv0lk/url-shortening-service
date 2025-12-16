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

type CalculatedStatistics struct {
	UrlToken        string         `json:"url_token"`
	TotalClicks     int            `json:"total_clicks"`
	UniqueCountries map[string]int `json:"unique_countries"`
	UniqueCities    map[string]int `json:"unique_cities"`
	DeviceTypeStats map[string]int `json:"device_types"`
	ReferrerStats   map[string]int `json:"referrer_stats"`
}

type StatisticsProcessor interface {
	ProcessEvent(ctx context.Context, eventData []byte) error
}

type StatsEventAdder interface {
	AddStatsEvent(ctx context.Context, event ProcessedStatsEvent) error
}

type StatisticsCalculator interface {
	CalculateStatistics(ctx context.Context, urlToken string) (CalculatedStatistics, error)
}

type StatisticsSender interface {
	SendEvent(ctx context.Context, rawEvent RawStatsEvent) error
}
