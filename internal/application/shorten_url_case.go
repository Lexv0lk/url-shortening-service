package application

import (
	"context"
	"url-shortening-service/internal/domain"
)

type UrlShortener struct {
	store       domain.MappingInfoAdder
	idGenerator domain.IdGenerator
}

func NewUrlShortener(idGenerator domain.IdGenerator, store domain.MappingInfoAdder) *UrlShortener {
	return &UrlShortener{
		store:       store,
		idGenerator: idGenerator,
	}
}

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
