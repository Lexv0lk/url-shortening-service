package application

import (
	"context"
	"errors"
	"url-shortening-service/internal/domain"
)

type UrlDeleter struct {
	cache   domain.UrlTokenDeleter
	storage domain.MappingInfoDeleter
	logger  domain.Logger
}

func NewUrlDeleter(cache domain.UrlTokenDeleter, storage domain.MappingInfoDeleter, logger domain.Logger) *UrlDeleter {
	return &UrlDeleter{
		cache:   cache,
		storage: storage,
		logger:  logger,
	}
}

func (ud *UrlDeleter) DeleteUrl(ctx context.Context, urlToken string) error {
	err := ud.storage.DeleteMappingInfo(ctx, urlToken)
	if err != nil {
		return err
	}

	err = ud.cache.DeleteMapping(ctx, urlToken)
	if errors.Is(err, &domain.TokenNonExistingError{}) {
		ud.logger.Warn(err.Error())
	} else if err != nil {
		return err
	}

	return nil
}
