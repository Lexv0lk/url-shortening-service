//go:generate mockgen -source=storage.go -destination=./mocks/storage.go -package=mocks
package domain

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// MappedGetSetter combines URL retrieval and mapping creation capabilities.
// Used for cache implementations that need both read and write access.
type MappedGetSetter interface {
	OriginalUrlGetter
	UrlTokenSetter
}

// MappingInfoGetAdder combines mapping info retrieval and creation capabilities.
// Used for storage implementations that need both read and write access to mapping details.
type MappingInfoGetAdder interface {
	MappingInfoGetter
	MappingInfoAdder
}

// OriginalUrlGetter defines the interface for retrieving original URLs by their short token.
type OriginalUrlGetter interface {
	// GetOriginalUrl retrieves the original URL for a given short URL token.
	// Returns the original URL and true if found, or empty string and false if not found.
	GetOriginalUrl(ctx context.Context, shortUrl string) (string, bool)
}

// UrlTokenSetter defines the interface for creating URL mappings.
type UrlTokenSetter interface {
	// SetMapping creates a new mapping between an original URL and its token.
	// Returns an error if the mapping could not be created.
	SetMapping(ctx context.Context, originalUrl, urlToken string) error
}

// UrlTokenDeleter defines the interface for deleting URL mappings from cache.
type UrlTokenDeleter interface {
	// DeleteMapping removes a URL mapping by its token.
	// Returns an error if the deletion fails.
	DeleteMapping(ctx context.Context, urlToken string) error
}

// MappingInfoGetter defines the interface for retrieving full mapping information.
type MappingInfoGetter interface {
	// GetMappingByToken retrieves complete mapping information for a given token.
	// Returns MappingInfo and true if found, or empty MappingInfo and false if not found.
	GetMappingByToken(ctx context.Context, urlToken string) (MappingInfo, bool)
}

// MappingInfoLastIdGetter defines the interface for retrieving the last used mapping ID.
type MappingInfoLastIdGetter interface {
	// GetLastId retrieves the highest ID currently used in the mappings table.
	// Returns the last ID and an error if the query fails.
	GetLastId(ctx context.Context) (int64, error)
}

// MappingInfoAdder defines the interface for adding new URL mappings with full details.
type MappingInfoAdder interface {
	// AddNewMapping creates a new URL mapping with the specified ID, original URL, and token.
	// Returns the created MappingInfo and an error if the operation fails.
	// May return *UrlExistingError if a mapping for this URL already exists.
	AddNewMapping(ctx context.Context, id int64, originalUrl string, shortUrl string) (MappingInfo, error)
}

// MappingInfoUpdater defines the interface for updating existing URL mappings.
type MappingInfoUpdater interface {
	// UpdateOriginalUrl updates the original URL for an existing token.
	// Returns the updated MappingInfo and an error if the operation fails.
	// May return *TokenNonExistingError if the token does not exist.
	UpdateOriginalUrl(ctx context.Context, urlToken string, newOriginalUrl string) (MappingInfo, error)
}

// MappingInfoDeleter defines the interface for deleting URL mappings from persistent storage.
type MappingInfoDeleter interface {
	// DeleteMappingInfo removes a URL mapping by its token from persistent storage.
	// Returns an error if the deletion fails.
	// May return *TokenNonExistingError if the token does not exist.
	DeleteMappingInfo(ctx context.Context, urlToken string) error
}

// IdGenerator defines the interface for generating unique mapping IDs.
type IdGenerator interface {
	// GetNextId generates and returns the next unique ID for URL mappings.
	// Returns the new ID and an error if generation fails.
	GetNextId(ctx context.Context) (int64, error)
}

// KeyStorage defines the interface for basic key-value operations in Redis.
type KeyStorage interface {
	KeySetter
	KeyGetter
	KeyDeleter
}

// KeySetIncrementer defines the interface for setting and incrementing keys in Redis.
type KeySetIncrementer interface {
	KeySetter
	Incr(ctx context.Context, key string) *redis.IntCmd
}

// KeyGetter defines the interface for getting keys from Redis.
type KeyGetter interface {
	Get(ctx context.Context, key string) *redis.StringCmd
}

// KeyDeleter defines the interface for deleting keys from Redis.
type KeyDeleter interface {
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// KeySetter defines the interface for setting keys in Redis.
type KeySetter interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}
