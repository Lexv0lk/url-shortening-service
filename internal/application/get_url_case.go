package application

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"
)

type UrlGetter struct {
	cache  domain.MappedGetSetter
	store  domain.MappingInfoGetter
	logger domain.Logger
}

func NewUrlGetter(cache domain.MappedGetSetter, store domain.MappingInfoGetter, logger domain.Logger) *UrlGetter {
	return &UrlGetter{
		cache:  cache,
		store:  store,
		logger: logger,
	}
}

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
