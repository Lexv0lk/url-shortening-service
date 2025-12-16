package urlcases

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"
)

// UrlUpdater handles URL mapping update operations.
// It updates the original URL associated with an existing token.
type UrlUpdater struct {
	cache   domain.UrlTokenSetter
	storage domain.MappingInfoUpdater
	logger  domain.Logger
}

// NewUrlUpdater creates a new UrlUpdater instance.
// Parameters:
//   - cache: cache storage for URL mappings (e.g., Redis)
//   - storage: persistent storage for URL mappings (e.g., PostgreSQL)
//   - logger: logger for recording info messages
func NewUrlUpdater(cache domain.UrlTokenSetter, storage domain.MappingInfoUpdater, logger domain.Logger) *UrlUpdater {
	return &UrlUpdater{
		cache:   cache,
		storage: storage,
		logger:  logger,
	}
}

// UpdateUrlMapping updates the original URL for an existing URL token.
// It validates the new URL and updates the mapping in persistent storage.
//
// Returns the updated MappingInfo.
//
// Returns an error if:
//   - *domain.InvalidUrlError: the new URL format is invalid or scheme is unsupported
//   - *domain.TokenNonExistingError: the token does not exist in storage
//   - Storage operation fails
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
