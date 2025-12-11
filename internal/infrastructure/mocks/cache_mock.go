package mocks

type LocalCache struct {
	storage map[string]string
}

func NewLocalCache() *LocalCache {
	return &LocalCache{
		storage: make(map[string]string),
	}
}

func (c *LocalCache) SetMapping(originalUrl, urlToken string) error {
	c.storage[urlToken] = originalUrl
	return nil
}

func (c *LocalCache) GetOriginalUrl(urlToken string) (string, bool) {
	originalUrl, found := c.storage[urlToken]
	return originalUrl, found
}
