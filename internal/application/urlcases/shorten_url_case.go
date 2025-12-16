package urlcases

import (
	"context"
	"url-shortening-service/internal/domain"
)

// UrlShortener handles URL shortening operations.
// It generates unique tokens for URLs and stores the mappings.
type UrlShortener struct {
	store       domain.MappingInfoAdder
	idGenerator domain.IdGenerator
}

// NewUrlShortener creates a new UrlShortener instance.
// Parameters:
//   - idGenerator: generates unique IDs for new URL mappings
//   - store: persistent storage for URL mappings
func NewUrlShortener(idGenerator domain.IdGenerator, store domain.MappingInfoAdder) *UrlShortener {
	return &UrlShortener{
		store:       store,
		idGenerator: idGenerator,
	}
}

// ShortenUrl creates a shortened URL for the given original URL.
// It validates the URL, generates a unique ID and token, and stores the mapping.
//
// Returns the created MappingInfo containing the new short URL token.
//
// Returns an error if:
//   - *domain.InvalidUrlError: the URL format is invalid or scheme is unsupported
//   - ID generation fails
//   - Storage operation fails
func (u *UrlShortener) ShortenUrl(ctx context.Context, originalUrl string) (domain.MappingInfo, error) {
	err := domain.ValidateURL(originalUrl)
	if err != nil {
		return domain.MappingInfo{}, err
	}

	id, err := u.idGenerator.GetNextId(ctx)
	if err != nil {
		return domain.MappingInfo{}, err
	}

	urlToken := domain.GenerateToken(id)
	return u.store.AddNewMapping(ctx, id, originalUrl, urlToken)
}
