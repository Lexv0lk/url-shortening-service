package application

import (
	"fmt"
	"url-shortening-service/internal/domain"
)

type MappedGetSetter interface {
	OriginalUrlGetter
	ShortUrlSetter
}

type OriginalUrlGetter interface {
	GetOriginalUrl(shortUrl string) (string, bool)
}

type ShortUrlSetter interface {
	SetMapping(originalUrl, shortUrl string) error
}

type MappingInfoGetter interface {
	GetMapping(urlToken string) (domain.MappingInfo, bool)
}

type UrlGetter struct {
	cache  MappedGetSetter
	store  MappingInfoGetter
	logger domain.Logger
}

func NewUrlGetter(cache MappedGetSetter, store MappingInfoGetter, logger domain.Logger) *UrlGetter {
	return &UrlGetter{
		cache:  cache,
		store:  store,
		logger: logger,
	}
}

func (u *UrlGetter) GetOriginalUrl(urlToken string) (string, error) {
	if originalUrl, found := u.cache.GetOriginalUrl(urlToken); found {
		return originalUrl, nil
	}

	mappingInfo, found := u.store.GetMapping(urlToken)
	if !found {
		return "", &domain.UrlNonExistingError{Msg: fmt.Sprintf("short URL not found for original URL: %s", urlToken)}
	}

	err := u.cache.SetMapping(mappingInfo.OriginalURL, urlToken)
	if err != nil {
		u.logger.Warn("Failed to cache short URL for original URL")
	}

	return mappingInfo.OriginalURL, nil
}
