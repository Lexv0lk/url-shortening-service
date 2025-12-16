package urlcases

import (
	"context"
	"errors"
	"url-shortening-service/internal/domain"
)

// UrlDeleter handles URL mapping deletion operations.
// It removes mappings from both persistent storage and cache.
type UrlDeleter struct {
	cache   domain.UrlTokenDeleter
	storage domain.MappingInfoDeleter
	logger  domain.Logger
}

// NewUrlDeleter creates a new UrlDeleter instance.
// Parameters:
//   - cache: cache storage for URL mappings (e.g., Redis)
//   - storage: persistent storage for URL mappings (e.g., PostgreSQL)
//   - logger: logger for recording warnings and errors
func NewUrlDeleter(cache domain.UrlTokenDeleter, storage domain.MappingInfoDeleter, logger domain.Logger) *UrlDeleter {
	return &UrlDeleter{
		cache:   cache,
		storage: storage,
		logger:  logger,
	}
}

// DeleteUrl removes a URL mapping by its token from both storage and cache.
// The operation first deletes from persistent storage, then from cache.
// Cache deletion failures are logged as warnings but don't cause the operation to fail.
//
// Returns an error if:
//   - *domain.TokenNonExistingError: the token does not exist in storage
//   - Other errors from storage or cache operations
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
