//go:generate mockgen -source=statistics.go -destination=mocks/statistics.go -package=mocks
package domain

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// RawStatsEvent represents an unprocessed statistics event captured during URL redirection.
// It contains raw data that needs to be processed before storage.
type RawStatsEvent struct {
	UrlToken  string    `json:"url_token"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referrer  string    `json:"referrer"`
}

// ProcessedStatsEvent represents a statistics event after processing.
// IP addresses are resolved to geographic locations and user agents are parsed.
type ProcessedStatsEvent struct {
	UrlToken   string
	Timestamp  time.Time
	Country    string
	City       string
	DeviceType string
	Referrer   string
}

// CalculatedStatistics represents aggregated statistics for a shortened URL.
type CalculatedStatistics struct {
	// UrlToken is the short URL token these statistics belong to.
	UrlToken string `json:"url_token"`
	// TotalClicks is the total number of times this URL was accessed.
	TotalClicks int `json:"total_clicks"`
	// UniqueCountries maps country names to their access counts.
	UniqueCountries map[string]int `json:"unique_countries"`
	// UniqueCities maps city names to their access counts.
	UniqueCities map[string]int `json:"unique_cities"`
	// DeviceTypeStats maps device types to their access counts.
	DeviceTypeStats map[string]int `json:"device_types"`
	// ReferrerStats maps referrer URLs to their access counts.
	ReferrerStats map[string]int `json:"referrer_stats"`
}

// StatisticsProcessor defines the interface for processing raw statistics events.
type StatisticsProcessor interface {
	// ProcessEvent processes a raw statistics event from byte data.
	// Returns an error if parsing or processing fails.
	ProcessEvent(ctx context.Context, eventData []byte) error
}

// StatsEventAdder defines the interface for adding processed statistics events to storage.
type StatsEventAdder interface {
	// AddStatsEvent persists a processed statistics event.
	// Returns an error if the storage operation fails.
	AddStatsEvent(ctx context.Context, event ProcessedStatsEvent) error
}

// StatisticsCalculator defines the interface for calculating aggregated statistics.
type StatisticsCalculator interface {
	// CalculateStatistics computes aggregated statistics for a given URL token.
	// Returns CalculatedStatistics and an error if calculation fails.
	CalculateStatistics(ctx context.Context, urlToken string) (CalculatedStatistics, error)
}

// StatisticsSender defines the interface for sending raw statistics events to a message bus.
type StatisticsSender interface {
	// SendEvent publishes a raw statistics event for asynchronous processing.
	// Returns an error if the event could not be sent.
	SendEvent(ctx context.Context, rawEvent RawStatsEvent) error
}

// MessageFetcher defines the interface for fetching and committing messages from a message bus.
type MessageFetcher interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// MessageWriter defines the interface for writing messages to a message bus.
type MessageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}
