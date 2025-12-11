package mocks

import "context"

type LocalCache struct {
	storage map[string]string
}

func NewLocalCache() *LocalCache {
	return &LocalCache{
		storage: make(map[string]string),
	}
}

func (c *LocalCache) SetMapping(ctx context.Context, originalUrl, urlToken string) error {
	c.storage[urlToken] = originalUrl
	return nil
}

func (c *LocalCache) GetOriginalUrl(ctx context.Context, urlToken string) (string, bool) {
	originalUrl, found := c.storage[urlToken]
	return originalUrl, found
}
