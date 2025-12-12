package application

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"
)

type UrlUpdater struct {
	cache   domain.UrlTokenSetter
	storage domain.MappingInfoUpdater
	logger  domain.Logger
}

func NewUrlUpdater(cache domain.UrlTokenSetter, storage domain.MappingInfoUpdater, logger domain.Logger) *UrlUpdater {
	return &UrlUpdater{
		cache:   cache,
		storage: storage,
		logger:  logger,
	}
}

func (u *UrlUpdater) UpdateUrlMapping(ctx context.Context, urlToken, newOriginalUrl string) (domain.MappingInfo, error) {
	err := domain.ValidateURL(newOriginalUrl)
	if err != nil {
		return domain.MappingInfo{}, err
	}

	newInfo, err := u.storage.UpdateOriginalUrl(ctx, urlToken, newOriginalUrl)
	if err != nil {
		return domain.MappingInfo{}, err
	}

	u.logger.Info(fmt.Sprintf("Updated URL mapping for token: %s", urlToken))
	return newInfo, nil
}
