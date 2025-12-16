package mocks

import "context"

// LocalCache is an in-memory mock implementation of URL mapping cache.
// It is intended for testing purposes only.
type LocalCache struct {
	storage map[string]string
}

// NewLocalCache creates a new LocalCache instance with an empty storage map.
func NewLocalCache() *LocalCache {
	return &LocalCache{
		storage: make(map[string]string),
	}
}

// SetMapping stores a URL mapping in the local cache.
// Always returns nil as this mock implementation never fails.
func (c *LocalCache) SetMapping(ctx context.Context, originalUrl, urlToken string) error {
	c.storage[urlToken] = originalUrl
	return nil
}

// GetOriginalUrl retrieves the original URL for a given token from the local cache.
// Returns the URL and true if found, or empty string and false if not found.
func (c *LocalCache) GetOriginalUrl(ctx context.Context, urlToken string) (string, bool) {
	originalUrl, found := c.storage[urlToken]
	return originalUrl, found
}
