package urlcases

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"
)

// UrlGetter retrieves original URLs by their short token.
// It implements a cache-aside pattern: first checking cache, then falling back to storage.
type UrlGetter struct {
	cache  domain.MappedGetSetter
	store  domain.MappingInfoGetter
	logger domain.Logger
}

// NewUrlGetter creates a new UrlGetter instance.
// Parameters:
//   - cache: cache storage supporting get and set operations (e.g., Redis)
//   - store: persistent storage for retrieving mapping information
//   - logger: logger for recording warnings
func NewUrlGetter(cache domain.MappedGetSetter, store domain.MappingInfoGetter, logger domain.Logger) *UrlGetter {
	return &UrlGetter{
		cache:  cache,
		store:  store,
		logger: logger,
	}
}

// GetOriginalUrl retrieves the original URL for a given short URL token.
// It first checks the cache, and on cache miss, queries the persistent storage
// and populates the cache for future requests.
//
// Returns an error if:
//   - *domain.UrlNonExistingError: the URL token was not found in storage
func (u *UrlGetter) GetOriginalUrl(ctx context.Context, urlToken string) (string, error) {
	if originalUrl, found := u.cache.GetOriginalUrl(ctx, urlToken); found {
		return originalUrl, nil
	}

	mappingInfo, found := u.store.GetMappingByToken(ctx, urlToken)
	if !found {
		return "", &domain.UrlNonExistingError{Msg: fmt.Sprintf("short URL not found for original URL: %s", urlToken)}
	}

	err := u.cache.SetMapping(ctx, mappingInfo.OriginalURL, urlToken)
	if err != nil {
		u.logger.Warn("Failed to cache short URL for original URL")
	}

	return mappingInfo.OriginalURL, nil
}
